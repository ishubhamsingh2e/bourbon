package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ishubhamsingh2e/bourbon/cmd/bourbon/internal/config"
	"github.com/ishubhamsingh2e/bourbon/cmd/bourbon/internal/generator"
	"github.com/spf13/cobra"
)

func findLocalBourbonRoot(fromDir string) string {
	if root := os.Getenv("BOURBON_ROOT"); root != "" {
		if abs, err := filepath.Abs(root); err == nil {
			return abs
		}
	}

	dir, err := filepath.Abs(fromDir)
	if err != nil {
		return ""
	}
	for {
		modPath := filepath.Join(dir, "go.mod")
		if data, err := os.ReadFile(modPath); err == nil {
			lines := strings.Split(string(data), "\n")
			if len(lines) > 0 && strings.TrimSpace(lines[0]) == "module github.com/ishubhamsingh2e/bourbon" {
				return dir
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// Check directory of running executable
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		for {
			modPath := filepath.Join(exeDir, "go.mod")
			if data, err := os.ReadFile(modPath); err == nil {
				lines := strings.Split(string(data), "\n")
				if len(lines) > 0 && strings.TrimSpace(lines[0]) == "module github.com/ishubhamsingh2e/bourbon" {
					return exeDir
				}
			}
			parent := filepath.Dir(exeDir)
			if parent == exeDir {
				break
			}
			exeDir = parent
		}
	}

	return ""
}

func newNewCmd() *cobra.Command {
	var (
		profile     string
		structure   string
		db          string
		cache       string
		auth        string
		module      string
		useDefaults bool
		noRemember  bool
	)

	newCmd := &cobra.Command{
		Use:   "new <project-name>",
		Short: "Create a new Bourbon application project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName := args[0]
			targetDir := projectName

			if module == "" {
				module = projectName
			}

			selectedProfile := "fullstack"
			if profile != "" {
				selectedProfile = profile
			}
			selectedStructure := "feature"
			if structure != "" {
				selectedStructure = structure
			}
			selectedDB := "sqlite"
			if db != "" {
				selectedDB = db
			}
			selectedCache := "none"
			if cache != "" {
				selectedCache = cache
			}
			selectedAuth := "none"
			if auth != "" {
				selectedAuth = auth
			}

			// 3. Create Project Directory
			if err := os.MkdirAll(targetDir, 0755); err != nil {
				return fmt.Errorf("failed creating directory '%s': %w", targetDir, err)
			}

			eng := generator.New()

			// 4. Build and Write Manifest (bourbon.yaml)
			manifest := config.DefaultManifest(projectName, module, selectedProfile, selectedStructure)
			if err := os.WriteFile(filepath.Join(targetDir, "bourbon.yaml"), []byte(manifest.FormatYAML()), 0644); err != nil {
				return fmt.Errorf("failed writing bourbon.yaml: %w", err)
			}

			// 5. Compute default URLs based on choices
			defaultDBURL := ""
			switch selectedDB {
			case "sqlite":
				defaultDBURL = "file:storage.db?cache=shared&mode=rwc"
			case "postgres":
				defaultDBURL = fmt.Sprintf("postgres://postgres:postgres@localhost:5432/%s_dev?sslmode=disable", eng.RenderString("snake", projectName))
			case "mysql":
				defaultDBURL = fmt.Sprintf("root:root@tcp(localhost:3306)/%s_dev?parseTime=true", eng.RenderString("snake", projectName))
			}

			// 6. Dynamically build .env and .env.example with only selected options
			var envBuf strings.Builder
			envBuf.WriteString(fmt.Sprintf("APP_NAME=%s\nAPP_ENV=development\nPORT=8080\nHOST=0.0.0.0\n", projectName))

			if selectedDB != "none" && selectedDB != "" {
				envBuf.WriteString(fmt.Sprintf("\n# Database Configuration\nDB_DRIVER=%s\nDATABASE_URL=%s\n", selectedDB, defaultDBURL))
			}

			if selectedCache != "none" && selectedCache != "" {
				envBuf.WriteString(fmt.Sprintf("\n# Cache Configuration\nCACHE_DRIVER=%s\n", selectedCache))
				if selectedCache == "redis" {
					envBuf.WriteString("REDIS_URL=redis://localhost:6379/0\n")
				}
			}

			if selectedAuth != "none" && selectedAuth != "" {
				envBuf.WriteString(fmt.Sprintf("\n# Security & Authentication\nAUTH_STRATEGY=%s\n", selectedAuth))
				switch selectedAuth {
				case "jwt":
					envBuf.WriteString("JWT_SECRET=bourbon-super-secret-development-key-change-in-prod\n")
				case "session":
					envBuf.WriteString("SESSION_SECRET=bourbon-session-development-key-change-in-prod\n")
				}
			}

			envContent := envBuf.String()
			_ = os.WriteFile(filepath.Join(targetDir, ".env"), []byte(envContent), 0644)
			_ = os.WriteFile(filepath.Join(targetDir, ".env.example"), []byte(envContent), 0644)

			// 7. Write go.mod
			goModContent := fmt.Sprintf("module %s\n\ngo 1.23\n\nrequire github.com/ishubhamsingh2e/bourbon v%s\n", module, Version)
			absTarget, _ := filepath.Abs(targetDir)
			if localRoot := findLocalBourbonRoot(absTarget); localRoot != "" && localRoot != absTarget {
				if rel, err := filepath.Rel(absTarget, localRoot); err == nil {
					goModContent += fmt.Sprintf("\nreplace github.com/ishubhamsingh2e/bourbon => %s\n", rel)
				}
			}

			if err := os.WriteFile(filepath.Join(targetDir, "go.mod"), []byte(goModContent), 0644); err != nil {
				return fmt.Errorf("failed writing go.mod: %w", err)
			}

			// 8. Create config/ directory and settings.go tailored strictly to selected options
			cfgDir := filepath.Join(targetDir, "config")
			if err := os.MkdirAll(cfgDir, 0755); err != nil {
				return err
			}

			var structFields strings.Builder
			var driverMethods strings.Builder

			structFields.WriteString(fmt.Sprintf("	AppName     string `env:\"APP_NAME,default=%s\"`\n", projectName))
			structFields.WriteString("	Environment string `env:\"APP_ENV,default=development\"`\n")
			structFields.WriteString("	Port        int    `env:\"PORT,default=8080\"`\n")
			structFields.WriteString("	Host        string `env:\"HOST,default=0.0.0.0\"`\n")

			if selectedDB != "none" && selectedDB != "" {
				structFields.WriteString(fmt.Sprintf("	DatabaseURL string `env:\"DATABASE_URL,default=%s\"`\n", defaultDBURL))
				switch selectedDB {
				case "sqlite":
					driverMethods.WriteString(`// DBDriver returns the configured database driver provider.
func (s *Settings) DBDriver() database.Driver {
	return sqlite.Config{
		DSN: s.DatabaseURL,
	}
}

`)
				case "postgres":
					driverMethods.WriteString(`// DBDriver returns the configured database driver provider.
func (s *Settings) DBDriver() database.Driver {
	return postgres.Config{
		DSN: s.DatabaseURL,
	}
}

`)
				case "mysql":
					driverMethods.WriteString(`// DBDriver returns the configured database driver provider.
func (s *Settings) DBDriver() database.Driver {
	return mysql.Config{
		DSN: s.DatabaseURL,
	}
}

`)
				}
			} else {
				driverMethods.WriteString(`// DBDriver returns nil when no database is configured.
func (s *Settings) DBDriver() database.Driver {
	return nil
}

`)
			}

			driverMethods.WriteString(`// CacheDriver returns the configured cache driver provider.
func (s *Settings) CacheDriver() cache.Driver {
	return memcache.Config{
		CleanupInterval: 1 * time.Minute,
	}
}

// StorageDriver returns the configured file storage driver provider.
func (s *Settings) StorageDriver() storage.Driver {
	return localfs.Config{
		RootDir: "storage/uploads",
		BaseURL: "/storage",
	}
}

// EventDriver returns the configured event bus driver provider.
func (s *Settings) EventDriver() events.Driver {
	return inprocess.Config{
		BufferSize: 1024,
	}
}
`)

			dbImport := ""
			switch selectedDB {
			case "sqlite":
				dbImport = "\t\"github.com/ishubhamsingh2e/bourbon/contrib/sqlite\"\n"
			case "postgres":
				dbImport = "\t\"github.com/ishubhamsingh2e/bourbon/contrib/postgres\"\n"
			case "mysql":
				dbImport = "\t\"github.com/ishubhamsingh2e/bourbon/contrib/mysql\"\n"
			}

			settingsTmpl := fmt.Sprintf(`package config

import (
	"time"

	"github.com/ishubhamsingh2e/bourbon/cache"
	"github.com/ishubhamsingh2e/bourbon/cache/memcache"
	bconfig "github.com/ishubhamsingh2e/bourbon/config"
%s	"github.com/ishubhamsingh2e/bourbon/database"
	"github.com/ishubhamsingh2e/bourbon/events"
	"github.com/ishubhamsingh2e/bourbon/events/inprocess"
	"github.com/ishubhamsingh2e/bourbon/storage"
	"github.com/ishubhamsingh2e/bourbon/storage/localfs"
)

// Settings represents the strongly-typed application configuration.
type Settings struct {
%s}

// Load initializes application settings from .env and environment variables.
func Load() *Settings {
	var s Settings
	_ = bconfig.Load(&s)
	return &s
}

%s// AppConfig translates settings into the core Bourbon framework configuration.
func (s *Settings) AppConfig() *bconfig.Config {
	env := bconfig.EnvDevelopment
	if s.Environment == "production" {
		env = bconfig.EnvProduction
	} else if s.Environment == "staging" {
		env = bconfig.EnvStaging
	}

	return &bconfig.Config{
		App: bconfig.AppConfig{
			Name: s.AppName,
		},
		HTTP: bconfig.HTTPConfig{
			Host: s.Host,
			Port: s.Port,
		},
		Env: env,
	}
}
`, dbImport, structFields.String(), driverMethods.String())

			settingsCode, err := eng.RenderGo(settingsTmpl, nil)
			if err != nil {
				return fmt.Errorf("failed rendering config/settings.go: %w", err)
			}
			_ = os.WriteFile(filepath.Join(cfgDir, "settings.go"), settingsCode, 0644)

			// 9. Scaffold internal/platform/container.go
			platDir := filepath.Join(targetDir, "internal", "platform")
			_ = os.MkdirAll(platDir, 0755)

			platTmpl := `package platform

import (
	"fmt"

	"github.com/ishubhamsingh2e/bourbon/cache"
	"github.com/ishubhamsingh2e/bourbon/database"
	"github.com/ishubhamsingh2e/bourbon/events"
	"github.com/ishubhamsingh2e/bourbon/mail/devmail"
	"github.com/ishubhamsingh2e/bourbon/observability"
	"github.com/ishubhamsingh2e/bourbon/storage"
	"{{.Module}}/config"
)

// Container manages all technical infrastructure adapters and connection pools.
type Container struct {
	DB      database.DB
	Cache   cache.Cache
	Bus     events.Bus
	Mailer  *devmail.Mailer
	Metrics *observability.MetricsCollector
	Storage storage.Storage
}

// NewContainer initializes infrastructure adapters according to application settings.
func NewContainer(cfg *config.Settings) (*Container, error) {
	var db database.DB
	if driver := cfg.DBDriver(); driver != nil {
		var err error
		db, err = driver.Open()
		if err != nil {
			return nil, fmt.Errorf("failed opening database: %w", err)
		}
	}

	var ca cache.Cache
	if driver := cfg.CacheDriver(); driver != nil {
		var err error
		ca, err = driver.Open()
		if err != nil {
			return nil, fmt.Errorf("failed opening cache: %w", err)
		}
	}

	var st storage.Storage
	if driver := cfg.StorageDriver(); driver != nil {
		var err error
		st, err = driver.Open()
		if err != nil {
			return nil, fmt.Errorf("failed opening storage: %w", err)
		}
	}

	var bus events.Bus
	if driver := cfg.EventDriver(); driver != nil {
		var err error
		bus, err = driver.Open()
		if err != nil {
			return nil, fmt.Errorf("failed opening event bus: %w", err)
		}
	}

	mailer := devmail.New()
	metrics := observability.NewMetricsCollector()

	return &Container{
		DB:      db,
		Cache:   ca,
		Bus:     bus,
		Mailer:  mailer,
		Metrics: metrics,
		Storage: st,
	}, nil
}
`
			platBytes, _ := eng.RenderGo(platTmpl, map[string]string{"Module": module})
			_ = os.WriteFile(filepath.Join(platDir, "container.go"), platBytes, 0644)

			// 10. Scaffold internal/modules/system/routes.go
			sysDir := filepath.Join(targetDir, "internal", "modules", "system")
			_ = os.MkdirAll(sysDir, 0755)

			sysTmpl := `package system

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ishubhamsingh2e/bourbon/router"
	"{{.Module}}/config"
	"{{.Module}}/internal/platform"
)

// RegisterRoutes mounts all system and diagnostics endpoints.
func RegisterRoutes(r router.Router, cfg *config.Settings, plat *platform.Container) {
	r.Get("/healthz", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  "healthy",
			"service": cfg.AppName,
			"env":     cfg.Environment,
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	r.Get("/readyz", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
	})

	if plat.Metrics != nil {
		r.Get("/metrics", plat.Metrics.ServeHTTP)
	}
}
`
			sysBytes, _ := eng.RenderGo(sysTmpl, map[string]string{"Module": module})
			_ = os.WriteFile(filepath.Join(sysDir, "routes.go"), sysBytes, 0644)

			// 11. Write main.go (Clean Entrypoint with Enterprise Layout)
			var mainGoTmpl string
			if selectedStructure == "feature" {
				mainGoTmpl = `package main

import (
	"context"
	"log"
	"time"

	"github.com/ishubhamsingh2e/bourbon/app"
	"github.com/ishubhamsingh2e/bourbon/middleware"
	"{{.Module}}/config"
	"{{.Module}}/internal/modules/system"
	"{{.Module}}/internal/modules/users"
	"{{.Module}}/internal/platform"
)

func main() {
	cfg := config.Load()

	plat, err := platform.NewContainer(cfg)
	if err != nil {
		log.Fatalf("Failed initializing platform: %v", err)
	}

	userRepo := users.NewRepository(plat.DB)
	userService := users.NewService(userRepo)

	application := app.New(
		app.WithConfig(cfg.AppConfig()),
	)

	r := application.Router()

	// 1. Global Middleware Stack
	r.Use(middleware.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.RequestID())
	r.Use(middleware.CORS())
	r.Use(middleware.Timeout(30 * time.Second))
	if plat.Metrics != nil {
		r.Use(plat.Metrics.Middleware())
	}

	// 2. Feature Routes
	system.RegisterRoutes(r, cfg, plat)
	users.RegisterRoutes(r, userService)

	log.Printf("Starting %s on :%d", cfg.AppName, cfg.Port)
	if err := application.Run(context.Background()); err != nil {
		log.Fatalf("Server terminated: %v", err)
	}
}
`
			} else {
				mainGoTmpl = `package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/ishubhamsingh2e/bourbon/app"
	"github.com/ishubhamsingh2e/bourbon/middleware"
	"github.com/ishubhamsingh2e/bourbon/router"
	"{{.Module}}/config"
	"{{.Module}}/controllers"
)

func main() {
	cfg := config.Load()

	application := app.New(
		app.WithConfig(cfg.AppConfig()),
	)

	r := application.Router()

	// 1. Global Middleware Stack
	r.Use(middleware.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.RequestID())
	r.Use(middleware.CORS())

	// 2. Health & Observability Endpoints
	r.Get("/healthz", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  "healthy",
			"service": cfg.AppName,
			"env":     cfg.Environment,
		})
	})
	r.Get("/readyz", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
	})

	// 3. Controller Routes
	usersCtrl := controllers.NewUsersController()
	r.Route("/users", func(g router.Router) {
		g.Get("/", usersCtrl.Index)
		g.Post("/", usersCtrl.Create)
		g.Get("/{id}", usersCtrl.Show)
	})

	log.Printf("Starting %s on :%d", cfg.AppName, cfg.Port)
	if err := application.Run(context.Background()); err != nil {
		log.Fatalf("Server terminated: %v", err)
	}
}
`
			}

			mainGoBytes, err := eng.RenderGo(mainGoTmpl, map[string]string{
				"Module": module,
			})
			if err != nil {
				return fmt.Errorf("failed rendering main.go: %w", err)
			}
			if err := os.WriteFile(filepath.Join(targetDir, "main.go"), mainGoBytes, 0644); err != nil {
				return fmt.Errorf("failed writing main.go: %w", err)
			}

			// 12. Write main_test.go
			var testGoTmpl string
			if selectedStructure == "feature" {
				testGoTmpl = `package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ishubhamsingh2e/bourbon/app"
	"{{.Module}}/config"
	"{{.Module}}/internal/modules/system"
	"{{.Module}}/internal/modules/users"
	"{{.Module}}/internal/platform"
)

func setupTestApp() *app.App {
	cfg := config.Load()
	plat, _ := platform.NewContainer(cfg)
	userRepo := users.NewRepository(plat.DB)
	userService := users.NewService(userRepo)

	application := app.New(app.WithConfig(cfg.AppConfig()))
	r := application.Router()

	system.RegisterRoutes(r, cfg, plat)
	users.RegisterRoutes(r, userService)
	return application
}

func TestHealthEndpoint(t *testing.T) {
	application := setupTestApp()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	application.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got: %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "healthy") {
		t.Errorf("expected body to contain 'healthy', got: %s", w.Body.String())
	}
}

func TestUsersEndpoint(t *testing.T) {
	application := setupTestApp()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	application.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got: %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "users") {
		t.Errorf("expected body to contain 'users', got: %s", w.Body.String())
	}
}
`
			} else {
				testGoTmpl = `package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ishubhamsingh2e/bourbon/app"
	"github.com/ishubhamsingh2e/bourbon/router"
	"{{.Module}}/config"
	"{{.Module}}/controllers"
)

func setupTestApp() *app.App {
	cfg := config.Load()
	application := app.New(app.WithConfig(cfg.AppConfig()))
	r := application.Router()

	r.Get("/healthz", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(` + "`" + `{"status":"healthy"}` + "`" + `))
	})

	usersCtrl := controllers.NewUsersController()
	r.Route("/users", func(g router.Router) {
		g.Get("/", usersCtrl.Index)
		g.Post("/", usersCtrl.Create)
		g.Get("/{id}", usersCtrl.Show)
	})
	return application
}

func TestHealthEndpoint(t *testing.T) {
	application := setupTestApp()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	application.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got: %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "healthy") {
		t.Errorf("expected body to contain 'healthy', got: %s", w.Body.String())
	}
}

func TestUsersEndpoint(t *testing.T) {
	application := setupTestApp()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	application.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got: %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "users") {
		t.Errorf("expected body to contain 'users', got: %s", w.Body.String())
	}
}
`
			}

			testGoBytes, err := eng.RenderGo(testGoTmpl, map[string]string{
				"Module": module,
			})
			if err == nil {
				_ = os.WriteFile(filepath.Join(targetDir, "main_test.go"), testGoBytes, 0644)
			}

			// 13. Initial Feature/Layered Scaffolding
			if selectedStructure == "feature" {
				userDir := filepath.Join(targetDir, "internal", "modules", "users")
				_ = os.MkdirAll(userDir, 0755)

				ctrlCode, _ := eng.RenderGo(`package users

import (
	"encoding/json"
	"net/http"

	"github.com/ishubhamsingh2e/bourbon/router"
)

type Controller struct {
	service *Service
}

func NewController(s *Service) *Controller {
	return &Controller{service: s}
}

func (ctrl *Controller) Index(w http.ResponseWriter, r *http.Request) {
	items, err := ctrl.service.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"users": items,
	})
}

func (ctrl *Controller) Show(w http.ResponseWriter, r *http.Request) {
	id := router.URLParam(r, "id")
	item, err := ctrl.service.GetByID(r.Context(), id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "user not found"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(item)
}

func (ctrl *Controller) Create(w http.ResponseWriter, r *http.Request) {
	var u User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := ctrl.service.Create(r.Context(), &u); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(u)
}
`, nil)
				_ = os.WriteFile(filepath.Join(userDir, "controller.go"), ctrlCode, 0644)

				srvCode, _ := eng.RenderGo(`package users

import (
	"context"
	"fmt"
	"time"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetByID(ctx context.Context, id string) (*User, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]*User, error) {
	return s.repo.List(ctx)
}

func (s *Service) Create(ctx context.Context, u *User) error {
	if u.Email == "" {
		return fmt.Errorf("email is required")
	}
	if u.ID == "" {
		u.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	return s.repo.Create(ctx, u)
}
`, nil)
				_ = os.WriteFile(filepath.Join(userDir, "service.go"), srvCode, 0644)

				repoCode, _ := eng.RenderGo(`package users

import (
	"context"
	"time"

	"github.com/ishubhamsingh2e/bourbon/database"
)

type Repository interface {
	FindByID(ctx context.Context, id string) (*User, error)
	Create(ctx context.Context, u *User) error
	List(ctx context.Context) ([]*User, error)
}

type sqlRepository struct {
	db database.DB
}

func NewRepository(db database.DB) Repository {
	return &sqlRepository{db: db}
}

func (r *sqlRepository) FindByID(ctx context.Context, id string) (*User, error) {
	if r.db == nil {
		return &User{ID: id, Email: "user@example.com", Name: "Demo User", CreatedAt: time.Now()}, nil
	}
	query := "SELECT id, email, name, created_at, updated_at FROM users WHERE id = ?"
	row := r.db.QueryRow(ctx, query, id)
	if row == nil {
		return &User{ID: id, Email: "user@example.com", Name: "Demo User", CreatedAt: time.Now()}, nil
	}

	var u User
	err := row.Scan(&u.ID, &u.Email, &u.Name, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return &User{ID: id, Email: "user@example.com", Name: "Demo User", CreatedAt: time.Now()}, nil
	}
	return &u, nil
}

func (r *sqlRepository) Create(ctx context.Context, u *User) error {
	if r.db == nil {
		return nil
	}
	query := "INSERT INTO users (id, email, name, created_at, updated_at) VALUES (?, ?, ?, ?, ?)"
	_, err := r.db.Exec(ctx, query, u.ID, u.Email, u.Name, u.CreatedAt, u.UpdatedAt)
	return err
}

func (r *sqlRepository) List(ctx context.Context) ([]*User, error) {
	if r.db == nil {
		return []*User{
			{ID: "1", Email: "admin@bourbon.dev", Name: "Admin User", CreatedAt: time.Now()},
		}, nil
	}
	query := "SELECT id, email, name, created_at, updated_at FROM users"
	rows, err := r.db.Query(ctx, query)
	if err != nil || rows == nil {
		return []*User{
			{ID: "1", Email: "admin@bourbon.dev", Name: "Admin User", CreatedAt: time.Now()},
		}, nil
	}
	defer rows.Close()

	var result []*User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, &u)
	}
	return result, nil
}
`, nil)
				_ = os.WriteFile(filepath.Join(userDir, "repository.go"), repoCode, 0644)

				modelCode, _ := eng.RenderGo(`package users

import "time"

type User struct {
	ID        string    `+"`json:\"id\" db:\"id\"`"+`
	Email     string    `+"`json:\"email\" db:\"email\"`"+`
	Name      string    `+"`json:\"name\" db:\"name\"`"+`
	CreatedAt time.Time `+"`json:\"created_at\" db:\"created_at\"`"+`
	UpdatedAt time.Time `+"`json:\"updated_at\" db:\"updated_at\"`"+`
}

func (u *User) TableName() string {
	return "users"
}
`, nil)
				_ = os.WriteFile(filepath.Join(userDir, "model.go"), modelCode, 0644)

				routesCode, _ := eng.RenderGo(`package users

import (
	"github.com/ishubhamsingh2e/bourbon/router"
)

// RegisterRoutes mounts all user feature endpoints on the provided router.
func RegisterRoutes(r router.Router, service *Service) {
	ctrl := NewController(service)

	r.Route("/users", func(g router.Router) {
		g.Get("/", ctrl.Index)
		g.Post("/", ctrl.Create)
		g.Get("/{id}", ctrl.Show)
	})
}
`, nil)
				_ = os.WriteFile(filepath.Join(userDir, "routes.go"), routesCode, 0644)
			} else {
				ctrlDir := filepath.Join(targetDir, "controllers")
				_ = os.MkdirAll(ctrlDir, 0755)

				ctrlCode, _ := eng.RenderGo(`package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/ishubhamsingh2e/bourbon/router"
)

type UsersController struct{}

func NewUsersController() *UsersController {
	return &UsersController{}
}

func (ctrl *UsersController) Index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"users": []map[string]string{
			{"id": "1", "name": "Bourbon Developer"},
		},
	})
}

func (ctrl *UsersController) Show(w http.ResponseWriter, r *http.Request) {
	id := router.URLParam(r, "id")
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"id":   id,
		"name": "Bourbon Developer",
	})
}

func (ctrl *UsersController) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "created",
	})
}
`, nil)
				_ = os.WriteFile(filepath.Join(ctrlDir, "users_controller.go"), ctrlCode, 0644)
			}

			// 13. Write db/migrations/20260819000000_create_users_table.sql
			migDir := filepath.Join(targetDir, "db", "migrations")
			_ = os.MkdirAll(migDir, 0755)
			migContent := `-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(36) PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
`
			_ = os.WriteFile(filepath.Join(migDir, "20260819000000_create_users_table.sql"), []byte(migContent), 0644)

			// 14. Write cmd/migrate/main.go if database is configured
			if selectedDB != "none" && selectedDB != "" {
				migCmdDir := filepath.Join(targetDir, "cmd", "migrate")
				_ = os.MkdirAll(migCmdDir, 0755)
				migCmdTmpl := `package main

import (
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"{{.Module}}/config"

	"github.com/ishubhamsingh2e/bourbon/database/migrator"
)

func main() {
	cfg := config.Load()
	driver := cfg.DBDriver()
	if driver == nil {
		log.Fatal("migrator: no database driver configured in config.Settings")
	}

	db, err := driver.Open()
	if err != nil {
		log.Fatalf("migrator: failed to open database: %v", err)
	}
	defer db.Close()

	action := "up"
	if len(os.Args) > 1 {
		action = os.Args[1]
	}

	migrationDir := "db/migrations"

	switch action {
	case "up", "migrate":
		if err := migrator.Migrate(db, migrationDir); err != nil {
			log.Fatalf("migration failed: %v", err)
		}
		fmt.Println("Migrations applied successfully.")

	case "rollback", "down":
		if err := migrator.Rollback(db, migrationDir); err != nil {
			log.Fatalf("rollback failed: %v", err)
		}
		fmt.Println("Rollback completed successfully.")

	case "status":
		statuses, err := migrator.Status(db, migrationDir)
		if err != nil {
			log.Fatalf("status check failed: %v", err)
		}

		if len(statuses) == 0 {
			fmt.Println("No migrations found in " + migrationDir)
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "STATUS\tMIGRATION\tAPPLIED AT")
		for _, s := range statuses {
			statusStr := "Pending"
			appliedAtStr := "-"
			if s.Applied {
				statusStr = "Applied"
				if s.AppliedAt != nil {
					appliedAtStr = s.AppliedAt.Format("2006-01-02 15:04:05")
				}
			}
			fmt.Fprintf(w, "%s\t%d_%s\t%s\n", statusStr, s.Version, s.Name, appliedAtStr)
		}
		_ = w.Flush()

	default:
		fmt.Fprintf(os.Stderr, "Unknown migration command: %s (supported: up, rollback, status)\n", action)
		os.Exit(1)
	}
}
`
				migCode, _ := eng.RenderGo(migCmdTmpl, map[string]string{"Module": module})
				_ = os.WriteFile(filepath.Join(migCmdDir, "main.go"), migCode, 0644)
			}

			// 15. Run go mod tidy in target directory
			tidyCmd := exec.Command("go", "mod", "tidy")
			tidyCmd.Dir = targetDir
			_ = tidyCmd.Run()

			cmd.Printf("Created Bourbon project '%s' (%s profile, %s structure) in ./%s\n", projectName, selectedProfile, selectedStructure, targetDir)
			cmd.Printf("To get started:\n  cd %s\n  bourbon migrate\n  bourbon dev\n", targetDir)
			return nil
		},
	}

	newCmd.Flags().StringVar(&profile, "profile", "", "Application profile (fullstack, api, web, worker)")
	newCmd.Flags().StringVar(&structure, "structure", "", "Architecture structure (feature, layered)")
	newCmd.Flags().StringVar(&db, "db", "", "Database backend (postgres, sqlite, mysql, none)")
	newCmd.Flags().StringVar(&cache, "cache", "", "Cache backend (redis, memory, none)")
	newCmd.Flags().StringVar(&auth, "auth", "", "Authentication strategy (session, jwt, none)")
	newCmd.Flags().StringVar(&module, "module", "", "Go module name (defaults to project name)")
	newCmd.Flags().BoolVar(&useDefaults, "defaults", false, "Use remembered developer defaults without displaying interactive prompts")
	newCmd.Flags().BoolVar(&noRemember, "no-remember", false, "Do not remember or update choices from this project creation")

	return newCmd
}
