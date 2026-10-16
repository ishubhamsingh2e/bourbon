package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ishubhamsingh2e/bourbon/cmd/bourbon/internal/config"
	"github.com/ishubhamsingh2e/bourbon/cmd/bourbon/internal/generator"
	"github.com/ishubhamsingh2e/bourbon/database/migrator"
	"github.com/spf13/cobra"
)

func detectProjectLayout() (structure, srcDir, module string, err error) {
	rootDir, manifest, err := config.FindProjectRoot(".")
	if err != nil {
		return "", "", "", err
	}

	structure = manifest.Project.Structure
	if structure == "" {
		structure = "feature"
	}

	srcDir = manifest.Paths.SourceDir
	if srcDir == "" {
		if structure == "layered" {
			srcDir = "controllers"
		} else {
			srcDir = "apps"
		}
	}

	module = manifest.Project.GoModule
	if module == "" {
		module = "app"
	}

	// Always operate relative to project root
	if rootDir != "." {
		_ = os.Chdir(rootDir)
	}

	return structure, srcDir, module, nil
}

func newMakeCmd() *cobra.Command {
	var force bool

	makeCmd := &cobra.Command{
		Use:   "make",
		Short: "Generate boilerplate components and scaffolding",
	}

	makeCmd.PersistentFlags().BoolVarP(&force, "force", "f", false, "Overwrite existing files")

	makeCmd.AddCommand(
		newMakeAppCmd(&force),
		newMakeControllerCmd(&force),
		newMakeServiceCmd(&force),
		newMakeRepositoryCmd(&force),
		newMakeModelCmd(&force),
		newMakeMigrationCmd(&force),
		newMakeMiddlewareCmd(&force),
		newMakeJobCmd(&force),
		newMakeTestCmd(&force),
	)

	return makeCmd
}

func newMakeAppCmd(force *bool) *cobra.Command {
	return &cobra.Command{
		Use:     "app <name>",
		Aliases: []string{"feature", "module"},
		Short:   "Generate a complete application feature slice (model, repository, service, controller, routes, test)",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			structure, srcDir, module, err := detectProjectLayout()
			if err != nil {
				return err
			}

			eng := generator.New()
			snakeName := eng.RenderString("snake", name)
			pascalName := eng.RenderString("pascal", name)
			pluralName := eng.RenderString("plural", snakeName)

			var (
				ctrlDest, ctrlTmpl string
				srvDest, srvTmpl   string
				repoDest, repoTmpl string
				modDest, modTmpl   string
				routDest, routTmpl string
				testDest, testTmpl string
				pkgName            string
			)

			if structure == "feature" {
				pkgName = snakeName
				baseDir := filepath.Join(srcDir, snakeName)
				ctrlDest = filepath.Join(baseDir, "controller.go")
				srvDest = filepath.Join(baseDir, "service.go")
				repoDest = filepath.Join(baseDir, "repository.go")
				modDest = filepath.Join(baseDir, "model.go")
				routDest = filepath.Join(baseDir, "routes.go")
				testDest = filepath.Join(baseDir, fmt.Sprintf("%s_test.go", snakeName))

				ctrlTmpl = `package {{.Package}}

import (
	"encoding/json"
	"net/http"

	"github.com/ishubhamsingh2e/bourbon/router"
)

type {{.Pascal}}Controller struct {
	service *{{.Pascal}}Service
}

func New{{.Pascal}}Controller(s *{{.Pascal}}Service) *{{.Pascal}}Controller {
	if s == nil {
		s = New{{.Pascal}}Service(nil)
	}
	return &{{.Pascal}}Controller{service: s}
}

func (ctrl *{{.Pascal}}Controller) Index(w http.ResponseWriter, r *http.Request) {
	items, err := ctrl.service.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data": items,
	})
}

func (ctrl *{{.Pascal}}Controller) Show(w http.ResponseWriter, r *http.Request) {
	id := router.URLParam(r, "id")
	item, err := ctrl.service.GetByID(r.Context(), id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "{{.Pascal}} not found"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(item)
}

func (ctrl *{{.Pascal}}Controller) Create(w http.ResponseWriter, r *http.Request) {
	var item {{.Pascal}}
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := ctrl.service.Create(r.Context(), &item); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(item)
}
`

				srvTmpl = `package {{.Package}}

import "context"

type {{.Pascal}}Service struct {
	repo *{{.Pascal}}Repository
}

func New{{.Pascal}}Service(r *{{.Pascal}}Repository) *{{.Pascal}}Service {
	if r == nil {
		r = New{{.Pascal}}Repository()
	}
	return &{{.Pascal}}Service{repo: r}
}

func (s *{{.Pascal}}Service) List(ctx context.Context) ([]*{{.Pascal}}, error) {
	return s.repo.FindAll(ctx)
}

func (s *{{.Pascal}}Service) GetByID(ctx context.Context, id string) (*{{.Pascal}}, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *{{.Pascal}}Service) Create(ctx context.Context, item *{{.Pascal}}) error {
	return s.repo.Save(ctx, item)
}
`

				repoTmpl = `package {{.Package}}

import "context"

type {{.Pascal}}Repository struct{}

func New{{.Pascal}}Repository() *{{.Pascal}}Repository {
	return &{{.Pascal}}Repository{}
}

func (r *{{.Pascal}}Repository) FindAll(ctx context.Context) ([]*{{.Pascal}}, error) {
	return []*{{.Pascal}}{}, nil
}

func (r *{{.Pascal}}Repository) FindByID(ctx context.Context, id string) (*{{.Pascal}}, error) {
	return &{{.Pascal}}{ID: id}, nil
}

func (r *{{.Pascal}}Repository) Save(ctx context.Context, item *{{.Pascal}}) error {
	return nil
}
`

				modTmpl = `package {{.Package}}

import "time"

type {{.Pascal}} struct {
	ID        string    ` + "`json:\"id\" db:\"id\"`" + `
	CreatedAt time.Time ` + "`json:\"created_at\" db:\"created_at\"`" + `
	UpdatedAt time.Time ` + "`json:\"updated_at\" db:\"updated_at\"`" + `
}

func (m *{{.Pascal}}) TableName() string {
	return "{{.Plural}}"
}
`

				routTmpl = `package {{.Package}}

import (
	"github.com/ishubhamsingh2e/bourbon/router"
)

// RegisterRoutes registers {{.Pascal}} application endpoints on the router.
func RegisterRoutes(r router.Router) {
	ctrl := New{{.Pascal}}Controller(nil)

	r.Route("/{{.Plural}}", func(g router.Router) {
		g.Get("/", ctrl.Index)
		g.Post("/", ctrl.Create)
		g.Get("/{id}", ctrl.Show)
	})
}
`

				testTmpl = `package {{.Package}}

import (
	"testing"
)

func Test{{.Pascal}}Flow(t *testing.T) {
	service := New{{.Pascal}}Service(nil)
	if service == nil {
		t.Fatal("expected service instance, got nil")
	}
}
`
			} else {
				pkgName = "controllers"
				ctrlDest = filepath.Join("controllers", fmt.Sprintf("%s_controller.go", snakeName))
				srvDest = filepath.Join("services", fmt.Sprintf("%s_service.go", snakeName))
				repoDest = filepath.Join("repositories", fmt.Sprintf("%s_repository.go", snakeName))
				modDest = filepath.Join("models", fmt.Sprintf("%s.go", snakeName))
				routDest = filepath.Join("routes", fmt.Sprintf("%s_routes.go", snakeName))
				testDest = filepath.Join("tests", fmt.Sprintf("%s_test.go", snakeName))

				ctrlTmpl = `package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/ishubhamsingh2e/bourbon/router"
	"{{.Module}}/models"
	"{{.Module}}/services"
)

type {{.Pascal}}Controller struct {
	service *services.{{.Pascal}}Service
}

func New{{.Pascal}}Controller(s *services.{{.Pascal}}Service) *{{.Pascal}}Controller {
	if s == nil {
		s = services.New{{.Pascal}}Service(nil)
	}
	return &{{.Pascal}}Controller{service: s}
}

func (ctrl *{{.Pascal}}Controller) Index(w http.ResponseWriter, r *http.Request) {
	items, err := ctrl.service.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data": items,
	})
}

func (ctrl *{{.Pascal}}Controller) Show(w http.ResponseWriter, r *http.Request) {
	id := router.URLParam(r, "id")
	item, err := ctrl.service.GetByID(r.Context(), id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "{{.Pascal}} not found"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(item)
}

func (ctrl *{{.Pascal}}Controller) Create(w http.ResponseWriter, r *http.Request) {
	var item models.{{.Pascal}}
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := ctrl.service.Create(r.Context(), &item); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(item)
}
`

				srvTmpl = `package services

import (
	"context"

	"{{.Module}}/models"
	"{{.Module}}/repositories"
)

type {{.Pascal}}Service struct {
	repo *repositories.{{.Pascal}}Repository
}

func New{{.Pascal}}Service(r *repositories.{{.Pascal}}Repository) *{{.Pascal}}Service {
	if r == nil {
		r = repositories.New{{.Pascal}}Repository()
	}
	return &{{.Pascal}}Service{repo: r}
}

func (s *{{.Pascal}}Service) List(ctx context.Context) ([]*models.{{.Pascal}}, error) {
	return s.repo.FindAll(ctx)
}

func (s *{{.Pascal}}Service) GetByID(ctx context.Context, id string) (*models.{{.Pascal}}, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *{{.Pascal}}Service) Create(ctx context.Context, item *models.{{.Pascal}}) error {
	return s.repo.Save(ctx, item)
}
`

				repoTmpl = `package repositories

import (
	"context"

	"{{.Module}}/models"
)

type {{.Pascal}}Repository struct{}

func New{{.Pascal}}Repository() *{{.Pascal}}Repository {
	return &{{.Pascal}}Repository{}
}

func (r *{{.Pascal}}Repository) FindAll(ctx context.Context) ([]*models.{{.Pascal}}, error) {
	return []*models.{{.Pascal}}{}, nil
}

func (r *{{.Pascal}}Repository) FindByID(ctx context.Context, id string) (*models.{{.Pascal}}, error) {
	return &models.{{.Pascal}}{ID: id}, nil
}

func (r *{{.Pascal}}Repository) Save(ctx context.Context, item *models.{{.Pascal}}) error {
	return nil
}
`

				modTmpl = `package models

import "time"

type {{.Pascal}} struct {
	ID        string    ` + "`json:\"id\" db:\"id\"`" + `
	CreatedAt time.Time ` + "`json:\"created_at\" db:\"created_at\"`" + `
	UpdatedAt time.Time ` + "`json:\"updated_at\" db:\"updated_at\"`" + `
}

func (m *{{.Pascal}}) TableName() string {
	return "{{.Plural}}"
}
`

				routTmpl = `package routes

import (
	"github.com/ishubhamsingh2e/bourbon/router"
	"{{.Module}}/controllers"
)

// Register{{.Pascal}}Routes registers {{.Pascal}} application endpoints on the router.
func Register{{.Pascal}}Routes(r router.Router) {
	ctrl := controllers.New{{.Pascal}}Controller(nil)

	r.Route("/{{.Plural}}", func(g router.Router) {
		g.Get("/", ctrl.Index)
		g.Post("/", ctrl.Create)
		g.Get("/{id}", ctrl.Show)
	})
}
`

				testTmpl = `package tests

import (
	"testing"

	"{{.Module}}/services"
)

func Test{{.Pascal}}Flow(t *testing.T) {
	service := services.New{{.Pascal}}Service(nil)
	if service == nil {
		t.Fatal("expected service instance, got nil")
	}
}
`
			}

			data := map[string]string{
				"Module":  module,
				"Package": pkgName,
				"Pascal":  pascalName,
				"Snake":   snakeName,
				"Plural":  pluralName,
			}

			filesToGenerate := []struct {
				dest string
				tmpl string
			}{
				{ctrlDest, ctrlTmpl},
				{srvDest, srvTmpl},
				{repoDest, repoTmpl},
				{modDest, modTmpl},
				{routDest, routTmpl},
				{testDest, testTmpl},
			}

			for _, f := range filesToGenerate {
				code, err := eng.RenderGo(f.tmpl, data)
				if err != nil {
					return fmt.Errorf("failed rendering %s: %w", f.dest, err)
				}
				if err := eng.WriteFile(f.dest, code, *force); err != nil {
					return err
				}
				cmd.Printf("Created: %s\n", f.dest)
			}

			cmd.Printf("\nSuccessfully scaffolded %s application.\n", pascalName)
			if structure == "feature" {
				cmd.Printf("To mount routes, add `%s.RegisterRoutes(r)` inside internal/routes/routes.go\n", snakeName)
			} else {
				cmd.Printf("To mount routes, add `routes.Register%sRoutes(r)` inside internal/routes/routes.go\n", pascalName)
			}
			return nil
		},
	}
}

func newMakeControllerCmd(force *bool) *cobra.Command {
	return &cobra.Command{
		Use:   "controller <name>",
		Short: "Generate a new HTTP controller",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			structure, srcDir, _, err := detectProjectLayout()
			if err != nil {
				return err
			}

			eng := generator.New()
			tmpl := `package {{.Package}}

import (
	"encoding/json"
	"net/http"
)

type {{.Name | pascal}}Controller struct{}

func New{{.Name | pascal}}Controller() *{{.Name | pascal}}Controller {
	return &{{.Name | pascal}}Controller{}
}

func (ctrl *{{.Name | pascal}}Controller) Index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "{{.Name | pascal}} list",
	})
}
`
			pkg := "controllers"
			var dest string
			if structure == "feature" {
				pkg = eng.RenderString("snake", name)
				dest = filepath.Join(srcDir, eng.RenderString("snake", name), "controller.go")
			} else {
				if srcDir == "internal" {
					dest = filepath.Join("internal", "controllers", fmt.Sprintf("%s_controller.go", eng.RenderString("snake", name)))
				} else {
					dest = filepath.Join("controllers", fmt.Sprintf("%s_controller.go", eng.RenderString("snake", name)))
				}
			}

			code, err := eng.RenderGo(tmpl, map[string]string{
				"Name":    name,
				"Package": pkg,
			})
			if err != nil {
				return fmt.Errorf("failed rendering controller: %w", err)
			}

			if err := eng.WriteFile(dest, code, *force); err != nil {
				return err
			}

			cmd.Printf("Created controller: %s\n", dest)
			return nil
		},
	}
}

func newMakeServiceCmd(force *bool) *cobra.Command {
	return &cobra.Command{
		Use:   "service <name>",
		Short: "Generate a new business service",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			structure, srcDir, _, err := detectProjectLayout()
			if err != nil {
				return err
			}

			eng := generator.New()
			tmpl := `package {{.Package}}

import "context"

type {{.Name | pascal}}Service struct{}

func New{{.Name | pascal}}Service() *{{.Name | pascal}}Service {
	return &{{.Name | pascal}}Service{}
}

func (s *{{.Name | pascal}}Service) Execute(ctx context.Context) error {
	return nil
}
`
			pkg := "services"
			var dest string
			if structure == "feature" {
				pkg = eng.RenderString("snake", name)
				dest = filepath.Join(srcDir, eng.RenderString("snake", name), "service.go")
			} else {
				if srcDir == "internal" {
					dest = filepath.Join("internal", "services", fmt.Sprintf("%s_service.go", eng.RenderString("snake", name)))
				} else {
					dest = filepath.Join("services", fmt.Sprintf("%s_service.go", eng.RenderString("snake", name)))
				}
			}

			code, err := eng.RenderGo(tmpl, map[string]string{
				"Name":    name,
				"Package": pkg,
			})
			if err != nil {
				return fmt.Errorf("failed rendering service: %w", err)
			}

			if err := eng.WriteFile(dest, code, *force); err != nil {
				return err
			}

			cmd.Printf("Created service: %s\n", dest)
			return nil
		},
	}
}

func newMakeRepositoryCmd(force *bool) *cobra.Command {
	return &cobra.Command{
		Use:   "repository <name>",
		Short: "Generate a new data repository",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			structure, srcDir, _, err := detectProjectLayout()
			if err != nil {
				return err
			}

			eng := generator.New()
			tmpl := `package {{.Package}}

import "context"

type {{.Name | pascal}}Repository struct{}

func New{{.Name | pascal}}Repository() *{{.Name | pascal}}Repository {
	return &{{.Name | pascal}}Repository{}
}

func (r *{{.Name | pascal}}Repository) FindByID(ctx context.Context, id string) (any, error) {
	return nil, nil
}
`
			pkg := "repositories"
			var dest string
			if structure == "feature" {
				pkg = eng.RenderString("snake", name)
				dest = filepath.Join(srcDir, eng.RenderString("snake", name), "repository.go")
			} else {
				if srcDir == "internal" {
					dest = filepath.Join("internal", "repositories", fmt.Sprintf("%s_repository.go", eng.RenderString("snake", name)))
				} else {
					dest = filepath.Join("repositories", fmt.Sprintf("%s_repository.go", eng.RenderString("snake", name)))
				}
			}

			code, err := eng.RenderGo(tmpl, map[string]string{
				"Name":    name,
				"Package": pkg,
			})
			if err != nil {
				return fmt.Errorf("failed rendering repository: %w", err)
			}

			if err := eng.WriteFile(dest, code, *force); err != nil {
				return err
			}

			cmd.Printf("Created repository: %s\n", dest)
			return nil
		},
	}
}

func newMakeModelCmd(force *bool) *cobra.Command {
	return &cobra.Command{
		Use:   "model <name>",
		Short: "Generate a new database model struct",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			structure, srcDir, _, err := detectProjectLayout()
			if err != nil {
				return err
			}

			eng := generator.New()
			tmpl := `package {{.Package}}

import "time"

type {{.Name | pascal}} struct {
	ID        string    ` + "`json:\"id\" db:\"id\"`" + `
	CreatedAt time.Time ` + "`json:\"created_at\" db:\"created_at\"`" + `
	UpdatedAt time.Time ` + "`json:\"updated_at\" db:\"updated_at\"`" + `
}

func (m *{{.Name | pascal}}) TableName() string {
	return "{{.Name | snake | plural}}"
}
`
			pkg := "models"
			var dest string
			if structure == "feature" {
				pkg = eng.RenderString("snake", name)
				dest = filepath.Join(srcDir, eng.RenderString("snake", name), "model.go")
			} else {
				if srcDir == "internal" {
					dest = filepath.Join("internal", "models", fmt.Sprintf("%s.go", eng.RenderString("snake", name)))
				} else {
					dest = filepath.Join("models", fmt.Sprintf("%s.go", eng.RenderString("snake", name)))
				}
			}

			code, err := eng.RenderGo(tmpl, map[string]string{
				"Name":    name,
				"Package": pkg,
			})
			if err != nil {
				return fmt.Errorf("failed rendering model: %w", err)
			}

			if err := eng.WriteFile(dest, code, *force); err != nil {
				return err
			}

			cmd.Printf("Created model: %s\n", dest)
			return nil
		},
	}
}

func newMakeMigrationCmd(_ *bool) *cobra.Command {
	return &cobra.Command{
		Use:   "migration [name]",
		Short: "Generate a new SQL migration file",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := "migration"
			if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
				name = args[0]
			}
			_, _, _, err := detectProjectLayout()
			if err != nil {
				return err
			}

			dest, err := migrator.Create("db/migrations", name, "", "")
			if err != nil {
				return fmt.Errorf("failed creating migration file: %w", err)
			}

			cmd.Printf("Created migration: %s\n", dest)
			return nil
		},
	}
}

func newMakeMiddlewareCmd(force *bool) *cobra.Command {
	return &cobra.Command{
		Use:   "middleware <name>",
		Short: "Generate a new HTTP middleware",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			_, _, _, err := detectProjectLayout()
			if err != nil {
				return err
			}

			eng := generator.New()
			tmpl := `package middleware

import (
	"net/http"

	"github.com/ishubhamsingh2e/bourbon/router"
)

func {{.Name | pascal}}() router.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}
}
`
			code, err := eng.RenderGo(tmpl, map[string]string{"Name": name})
			if err != nil {
				return fmt.Errorf("failed rendering middleware: %w", err)
			}

			dest := filepath.Join("middleware", fmt.Sprintf("%s.go", eng.RenderString("snake", name)))
			if _, err := os.Stat("internal"); err == nil {
				dest = filepath.Join("internal", "middleware", fmt.Sprintf("%s.go", eng.RenderString("snake", name)))
			}

			if err := eng.WriteFile(dest, code, *force); err != nil {
				return err
			}

			cmd.Printf("Created middleware: %s\n", dest)
			return nil
		},
	}
}

func newMakeJobCmd(force *bool) *cobra.Command {
	return &cobra.Command{
		Use:   "job <name>",
		Short: "Generate a new background job worker",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			_, _, _, err := detectProjectLayout()
			if err != nil {
				return err
			}

			eng := generator.New()
			tmpl := `package jobs

import "context"

type {{.Name | pascal}}Job struct {
	Payload []byte
}

func New{{.Name | pascal}}Job(payload []byte) *{{.Name | pascal}}Job {
	return &{{.Name | pascal}}Job{Payload: payload}
}

func (j *{{.Name | pascal}}Job) Name() string {
	return "{{.Name | snake}}"
}

func (j *{{.Name | pascal}}Job) PayloadData() []byte {
	return j.Payload
}

func (j *{{.Name | pascal}}Job) Execute(ctx context.Context) error {
	return nil
}
`
			code, err := eng.RenderGo(tmpl, map[string]string{"Name": name})
			if err != nil {
				return fmt.Errorf("failed rendering job: %w", err)
			}

			dest := filepath.Join("jobs", fmt.Sprintf("%s_job.go", eng.RenderString("snake", name)))
			if _, err := os.Stat("internal"); err == nil {
				dest = filepath.Join("internal", "jobs", fmt.Sprintf("%s_job.go", eng.RenderString("snake", name)))
			}

			if err := eng.WriteFile(dest, code, *force); err != nil {
				return err
			}

			cmd.Printf("Created job: %s\n", dest)
			return nil
		},
	}
}

func newMakeTestCmd(force *bool) *cobra.Command {
	return &cobra.Command{
		Use:   "test <name>",
		Short: "Generate a new test suite for a component",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			structure, srcDir, _, err := detectProjectLayout()
			if err != nil {
				return err
			}

			eng := generator.New()
			tmpl := `package {{.Package}}

import (
	"testing"
)

func Test{{.Name | pascal}}(t *testing.T) {
	// Add test assertions
}
`
			pkg := "main"
			var dest string
			if structure == "feature" {
				pkg = eng.RenderString("snake", name)
				dest = filepath.Join(srcDir, eng.RenderString("snake", name), fmt.Sprintf("%s_test.go", eng.RenderString("snake", name)))
			} else {
				dest = filepath.Join("tests", fmt.Sprintf("%s_test.go", eng.RenderString("snake", name)))
			}

			code, err := eng.RenderGo(tmpl, map[string]string{
				"Name":    name,
				"Package": pkg,
			})
			if err != nil {
				return fmt.Errorf("failed rendering test: %w", err)
			}

			if err := eng.WriteFile(dest, code, *force); err != nil {
				return err
			}

			cmd.Printf("Created test: %s\n", dest)
			return nil
		},
	}
}
