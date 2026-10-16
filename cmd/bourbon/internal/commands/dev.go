package commands

import (
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/ishubhamsingh2e/bourbon/cmd/bourbon/internal/config"
	"github.com/spf13/cobra"
)

// DevProxy holds requests while reloading application process.
type DevProxy struct {
	mu        sync.RWMutex
	ready     bool
	targetURL *url.URL
	proxy     *httputil.ReverseProxy
}

// NewDevProxy initializes a live-reload reverse proxy.
func NewDevProxy(targetPort int) *DevProxy {
	target, _ := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", targetPort))
	return &DevProxy{
		targetURL: target,
		proxy:     httputil.NewSingleHostReverseProxy(target),
	}
}

func (p *DevProxy) SetReady(ready bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.ready = ready
}

func (p *DevProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for i := 0; i < 50; i++ {
		p.mu.RLock()
		isReady := p.ready
		p.mu.RUnlock()

		if isReady {
			p.proxy.ServeHTTP(w, r)
			return
		}
		time.Sleep(50 * time.Millisecond)
	}

	http.Error(w, "Dev server compiling...", http.StatusServiceUnavailable)
}

func isIgnoredDir(path string) bool {
	base := filepath.Base(path)
	if strings.HasPrefix(base, ".") && base != "." {
		return true
	}
	switch base {
	case "storage", "tmp", "node_modules", "vendor", "bin":
		return true
	}
	return false
}

func isWatchedFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	base := filepath.Base(path)

	// Explicitly exclude .sql per instructions
	if ext == ".sql" {
		return false
	}

	if base == ".env" || strings.HasPrefix(base, ".env.") {
		return true
	}

	switch ext {
	case ".go", ".html", ".yaml", ".yml":
		return true
	}
	return false
}

// DevProcessManager oversees compiling and executing the application process.
type DevProcessManager struct {
	rootDir   string
	tmpBinary string
	cmd       *exec.Cmd
	mu        sync.Mutex
}

func (m *DevProcessManager) Build() error {
	buildCmd := exec.Command("go", "build", "-o", m.tmpBinary, ".")
	buildCmd.Dir = m.rootDir
	out, err := buildCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

func (m *DevProcessManager) Restart() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 1. Build first (resilient: do not terminate running server if build fails)
	if err := m.Build(); err != nil {
		fmt.Printf("\n[bourbon] Build failed:\n%s\n\n", err.Error())
		return err
	}

	// 2. Build succeeded: stop previous process
	m.stopCurrent()

	// 3. Start new server process
	child := exec.Command(m.tmpBinary)
	child.Dir = m.rootDir
	child.Stdout = os.Stdout
	child.Stderr = os.Stderr
	child.Stdin = os.Stdin
	child.Env = append(os.Environ(), "APP_ENV=development")

	if err := child.Start(); err != nil {
		fmt.Printf("[bourbon] Failed starting server: %v\n", err)
		return err
	}

	m.cmd = child
	fmt.Println("[bourbon] Server started (watching for changes)")
	return nil
}

func (m *DevProcessManager) stopCurrent() {
	if m.cmd != nil && m.cmd.Process != nil {
		_ = m.cmd.Process.Signal(os.Interrupt)

		done := make(chan error, 1)
		go func() {
			done <- m.cmd.Wait()
		}()

		select {
		case <-done:
		case <-time.After(1 * time.Second):
			_ = m.cmd.Process.Kill()
		}
		m.cmd = nil
	}
}

func (m *DevProcessManager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stopCurrent()
	if m.tmpBinary != "" {
		_ = os.Remove(m.tmpBinary)
	}
}

func newDevCmd() *cobra.Command {
	var dryRun bool

	devCmd := &cobra.Command{
		Use:   "dev",
		Short: "Start the live-reload development server with file watching",
		RunE: func(cmd *cobra.Command, args []string) error {
			rootDir, manifest, err := config.FindProjectRoot(".")
			if err != nil {
				return err
			}
			if rootDir != "." {
				_ = os.Chdir(rootDir)
			}

			appName := "app"
			if manifest != nil && manifest.Project.Name != "" {
				appName = manifest.Project.Name
			}

			cmd.Printf("Bourbon Dev Server starting for '%s'...\n", appName)
			cmd.Println("Watching file changes (.go, .html, .env, .yaml)...")

			if dryRun {
				cmd.Println("Dry-run dev mode completed successfully.")
				return nil
			}

			if _, err := exec.LookPath("go"); err != nil {
				return fmt.Errorf("'go' executable not found in PATH")
			}

			tmpDir, err := os.MkdirTemp("", "bourbon-dev-*")
			if err != nil {
				return err
			}
			defer os.RemoveAll(tmpDir)

			tmpBinary := filepath.Join(tmpDir, "server")
			mgr := &DevProcessManager{
				rootDir:   rootDir,
				tmpBinary: tmpBinary,
			}
			defer mgr.Stop()

			// Initial build and launch
			if err := mgr.Restart(); err != nil {
				cmd.Println("[bourbon] Initial build had errors; waiting for file changes to retry...")
			}

			watcher, err := fsnotify.NewWatcher()
			if err != nil {
				return fmt.Errorf("failed creating file watcher: %w", err)
			}
			defer watcher.Close()

			// Recursively add directories to watch
			err = filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, walkErr error) error {
				if walkErr != nil {
					return walkErr
				}
				if d.IsDir() {
					if isIgnoredDir(path) {
						return filepath.SkipDir
					}
					return watcher.Add(path)
				}
				return nil
			})
			if err != nil {
				return fmt.Errorf("failed scanning directories: %w", err)
			}

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

			var debounceTimer *time.Timer
			var debounceMu sync.Mutex

			for {
				select {
				case <-sigChan:
					cmd.Println("\n[bourbon] Shutting down dev server...")
					return nil
				case watchErr, ok := <-watcher.Errors:
					if !ok {
						return nil
					}
					cmd.Printf("[bourbon] Watcher error: %v\n", watchErr)
				case event, ok := <-watcher.Events:
					if !ok {
						return nil
					}

					// If new directory created, watch it
					if event.Has(fsnotify.Create) {
						if fi, err := os.Stat(event.Name); err == nil && fi.IsDir() {
							if !isIgnoredDir(event.Name) {
								_ = watcher.Add(event.Name)
							}
							continue
						}
					}

					if !isWatchedFile(event.Name) {
						continue
					}

					// Debounce restarts
					debounceMu.Lock()
					if debounceTimer != nil {
						debounceTimer.Stop()
					}
					debounceTimer = time.AfterFunc(200*time.Millisecond, func() {
						relPath, _ := filepath.Rel(rootDir, event.Name)
						fmt.Printf("\n[bourbon] Change detected: %s (rebuilding...)\n", relPath)
						_ = mgr.Restart()
					})
					debounceMu.Unlock()
				}
			}
		},
	}

	devCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Simulate dev server startup without blocking")
	return devCmd
}
