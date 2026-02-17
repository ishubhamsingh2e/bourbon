package dev

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Watcher struct {
	cmd     *exec.Cmd
	lastMod map[string]time.Time
}

func NewWatcher() *Watcher {
	return &Watcher{
		lastMod: make(map[string]time.Time),
	}
}

func (w *Watcher) Start() error {
	w.cmd = exec.Command("go", "run", ".")
	w.cmd.Stdout = os.Stdout
	w.cmd.Stderr = os.Stderr
	return w.cmd.Start()
}

func (w *Watcher) Stop() error {
	if w.cmd != nil && w.cmd.Process != nil {
		w.cmd.Process.Kill()
		return w.cmd.Wait()
	}
	return nil
}

func (w *Watcher) CheckChanges() bool {
	changed := false

	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		if filepath.Ext(path) != ".go" || strings.Contains(path, "tmp/") {
			return nil
		}

		modTime := info.ModTime()
		if prevTime, exists := w.lastMod[path]; !exists || modTime.After(prevTime) {
			w.lastMod[path] = modTime
			if exists {
				changed = true
			}
		}

		return nil
	})

	return changed
}
