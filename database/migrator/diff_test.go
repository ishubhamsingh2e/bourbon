package migrator_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ishubhamsingh2e/bourbon/database/dialect"
	"github.com/ishubhamsingh2e/bourbon/database/migrator"
)

type MockPrompter struct {
	BoolAnswers   map[string]bool
	StringAnswers map[string]string
	DefaultBool   bool
}

func (m *MockPrompter) PromptBool(question string, defaultVal bool) bool {
	for k, v := range m.BoolAnswers {
		if strings.Contains(question, k) {
			return v
		}
	}
	return m.DefaultBool
}

func (m *MockPrompter) PromptString(question string, defaultVal string) string {
	for k, v := range m.StringAnswers {
		if strings.Contains(question, k) {
			return v
		}
	}
	return defaultVal
}

func TestASTScannerAndDiff(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Create a model in apps/users/model.go
	usersDir := filepath.Join(tempDir, "apps", "users")
	if err := os.MkdirAll(usersDir, 0755); err != nil {
		t.Fatalf("failed creating users dir: %v", err)
	}

	modelCodeV1 := `package users

import "time"

type User struct {
	ID        string    ` + "`json:\"id\" db:\"id\"`" + `
	Email     string    ` + "`json:\"email\" db:\"email,not_null,unique\"`" + `
	Name      string    ` + "`json:\"name\" db:\"name,not_null\"`" + `
	CreatedAt time.Time ` + "`json:\"created_at\" db:\"created_at\"`" + `
	UpdatedAt time.Time ` + "`json:\"updated_at\" db:\"updated_at\"`" + `
}

func (u *User) TableName() string {
	return "users"
}
`
	modelPath := filepath.Join(usersDir, "model.go")
	if err := os.WriteFile(modelPath, []byte(modelCodeV1), 0644); err != nil {
		t.Fatalf("failed writing model.go: %v", err)
	}

	d, _ := dialect.Get("sqlite")

	// 2. Scan models
	models, err := migrator.ScanProjectModels(tempDir, "users", d)
	if err != nil {
		t.Fatalf("ScanProjectModels failed: %v", err)
	}
	if len(models) != 1 {
		t.Fatalf("expected 1 model, got %d", len(models))
	}
	if models[0].Name != "User" || models[0].TableName != "users" {
		t.Errorf("unexpected model: %+v", models[0])
	}
	if len(models[0].Fields) != 5 {
		t.Errorf("expected 5 fields, got %d", len(models[0].Fields))
	}

	// 3. Initial MakeMigrations
	prompter := &migrator.NonInteractivePrompter{AutoYes: true}
	migrationsDir := filepath.Join(tempDir, "db", "migrations")
	res, err := migrator.MakeMigrations(tempDir, migrationsDir, "users", "initial_users", false, prompter, "sqlite")
	if err != nil {
		t.Fatalf("MakeMigrations failed: %v", err)
	}
	if res == nil {
		t.Fatalf("expected migration to be created, got nil")
	}
	if _, err := os.Stat(res.FilePath); err != nil {
		t.Fatalf("migration file not found: %s", res.FilePath)
	}

	// 4. Running MakeMigrations again without changes should return nil
	res2, err := migrator.MakeMigrations(tempDir, migrationsDir, "users", "", false, prompter, "sqlite")
	if err != nil {
		t.Fatalf("second MakeMigrations failed: %v", err)
	}
	if res2 != nil {
		t.Errorf("expected no changes on second run, got %+v", res2)
	}

	// 5. Update model: Rename Name -> FullName, Add Age, Remove Email
	modelCodeV2 := `package users

import "time"

type User struct {
	ID        string    ` + "`json:\"id\" db:\"id\"`" + `
	FullName  string    ` + "`json:\"full_name\" db:\"full_name,not_null\"`" + `
	Age       int       ` + "`json:\"age\" db:\"age\" bourbon:\"default:0\"`" + `
	CreatedAt time.Time ` + "`json:\"created_at\" db:\"created_at\"`" + `
	UpdatedAt time.Time ` + "`json:\"updated_at\" db:\"updated_at\"`" + `
}

func (u *User) TableName() string {
	return "users"
}
`
	if err := os.WriteFile(modelPath, []byte(modelCodeV2), 0644); err != nil {
		t.Fatalf("failed writing model v2: %v", err)
	}

	mockPrompter := &MockPrompter{
		BoolAnswers: map[string]bool{
			"User.name to User.full_name": true,
			"User.email to User.age":      false,
		},
		DefaultBool: false,
	}

	res3, err := migrator.MakeMigrations(tempDir, migrationsDir, "users", "update_users", false, mockPrompter, "sqlite")
	if err != nil {
		t.Fatalf("third MakeMigrations failed: %v", err)
	}
	if res3 == nil {
		t.Fatalf("expected changes in v2, got nil")
	}

	content, err := os.ReadFile(res3.FilePath)
	if err != nil {
		t.Fatalf("failed reading migration v2: %v", err)
	}
	contentStr := string(content)

	// Verify rename column
	if !strings.Contains(contentStr, "RENAME COLUMN") || !strings.Contains(contentStr, "full_name") {
		t.Errorf("expected RENAME COLUMN in SQL, got:\n%s", contentStr)
	}

	// Verify add column age
	if !strings.Contains(contentStr, "ADD COLUMN") || !strings.Contains(contentStr, "age") {
		t.Errorf("expected ADD COLUMN age in SQL, got:\n%s", contentStr)
	}

	// 6. Test SQLMigrate
	sqlUp, err := migrator.SQLMigrate(migrationsDir, "initial_users", false)
	if err != nil {
		t.Fatalf("SQLMigrate failed: %v", err)
	}
	if !strings.Contains(sqlUp, "CREATE TABLE IF NOT EXISTS") {
		t.Errorf("expected CREATE TABLE in SQLMigrate, got: %s", sqlUp)
	}

	sqlDown, err := migrator.SQLMigrate(migrationsDir, "initial_users", true)
	if err != nil {
		t.Fatalf("SQLMigrate backwards failed: %v", err)
	}
	if !strings.Contains(sqlDown, "DROP TABLE IF EXISTS") {
		t.Errorf("expected DROP TABLE in SQLMigrate backwards, got: %s", sqlDown)
	}
}
