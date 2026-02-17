package cli

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func createApp(name string) {
	// Ensure we're in project root by checking for go.mod
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		fmt.Println("Error: Must run from project root (go.mod not found)")
		return
	}

	fmt.Printf("Creating app: %s\n", name)

	appDir := filepath.Join("apps", name)

	dirs := []string{
		appDir,
		filepath.Join(appDir, "migrations"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("Error creating directory %s: %v\n", dir, err)
			return
		}
	}

	files := map[string]string{
		filepath.Join(appDir, "models.go"):      modelFileTemplate,
		filepath.Join(appDir, "controllers.go"): controllerFileTemplate,
		filepath.Join(appDir, "routes.go"):      routesFileTemplate,
	}

	data := map[string]string{"AppName": name}

	for path, tmpl := range files {
		content := renderTemplate(tmpl, data)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			fmt.Printf("Error creating file %s: %v\n", path, err)
			return
		}
	}

	fmt.Printf("App created: %s\n", name)
	fmt.Printf("\nAdd '%s' to settings.toml under [apps.installed]\n", name)
}

func makeMigrationsForAllApps(migrationName string, force bool) {
	// Ensure we're in project root
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		fmt.Println("Error: Must run from project root (go.mod not found)")
		return
	}

	// Check if apps directory exists
	if _, err := os.Stat("apps"); os.IsNotExist(err) {
		fmt.Println("No apps found. Create an app with: bourbon create:app <name>")
		return
	}

	// Read all apps
	entries, err := os.ReadDir("apps")
	if err != nil {
		fmt.Printf("Error reading apps directory: %v\n", err)
		return
	}

	fmt.Println("Migrations:")
	appsWithChanges := 0
	for _, entry := range entries {
		if entry.IsDir() {
			appName := entry.Name()
			modelsPath := filepath.Join("apps", appName, "models.go")

			// Check if app has models and if they've changed
			if hasModels(modelsPath) && (force || hasModelChanges(appName)) {
				if err := makeMigration(appName, migrationName, force); err != nil {
					fmt.Printf("Error creating migration for %s: %v\n", appName, err)
					continue
				}
				appsWithChanges++
			}
		}
	}

	if appsWithChanges == 0 {
		if force {
			fmt.Println("No apps with models found.")
		} else {
			fmt.Println("No changes detected.")
		}
	}
}

func makeMigrationForApp(appName, migrationName string, force bool) {
	if err := makeMigration(appName, migrationName, force); err != nil {
		fmt.Printf("❌ %v\n", err)
	}
}

// getProjectModule reads the module name from go.mod
func getProjectModule() (string, error) {
	content, err := os.ReadFile("go.mod")
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}

	return "", fmt.Errorf("module name not found in go.mod")
}

// ModelInfo represents a detected model in the code
type ModelInfo struct {
	Name    string
	Package string
}

// detectModels parses models.go and extracts model structs
func detectModels(modelsPath string) ([]ModelInfo, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, modelsPath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var models []ModelInfo
	packageName := node.Name.Name

	// Look for struct declarations
	ast.Inspect(node, func(n ast.Node) bool {
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			return true
		}

		// Check if struct has at least one field (not empty struct)
		if structType.Fields == nil || len(structType.Fields.List) == 0 {
			return true
		}

		// Check if struct has gorm or json tags (indicating it's a model)
		hasModelTags := false
		for _, field := range structType.Fields.List {
			if field.Tag != nil {
				tagValue := field.Tag.Value
				// Look for common model tags
				if strings.Contains(tagValue, "gorm:") ||
					strings.Contains(tagValue, "json:") ||
					strings.Contains(tagValue, "db:") {
					hasModelTags = true
					break
				}
			}
		}

		if hasModelTags {
			// Skip BaseModel as it's a shared model that shouldn't be migrated
			if typeSpec.Name.Name != "BaseModel" {
				models = append(models, ModelInfo{
					Name:    typeSpec.Name.Name,
					Package: packageName,
				})
			}
		}

		return true
	})

	return models, nil
}

// hasModels checks if an app has models defined
func hasModels(modelsPath string) bool {
	models, err := detectModels(modelsPath)
	if err != nil {
		return false
	}
	return len(models) > 0
}

// getModelsHash returns a hash of the models.go file to detect changes
func getModelsHash(modelsPath string) (string, error) {
	content, err := os.ReadFile(modelsPath)
	if err != nil {
		return "", err
	}
	hash := md5.Sum(content)
	return hex.EncodeToString(hash[:]), nil
}

// hasModelChanges checks if models have changed since last migration
func hasModelChanges(appName string) bool {
	modelsPath := filepath.Join("apps", appName, "models.go")
	migrationsDir := filepath.Join("apps", appName, "migrations")
	hashFile := filepath.Join(migrationsDir, ".models_hash")

	// Get current hash
	currentHash, err := getModelsHash(modelsPath)
	if err != nil {
		return true // Assume changes if we can't read
	}

	// Read previous hash
	previousHash, err := os.ReadFile(hashFile)
	if err != nil {
		return true // No previous hash = changes
	}

	return string(previousHash) != currentHash
}

// saveModelsHash saves the current models hash
func saveModelsHash(appName string) error {
	modelsPath := filepath.Join("apps", appName, "models.go")
	migrationsDir := filepath.Join("apps", appName, "migrations")
	hashFile := filepath.Join(migrationsDir, ".models_hash")

	currentHash, err := getModelsHash(modelsPath)
	if err != nil {
		return err
	}

	return os.WriteFile(hashFile, []byte(currentHash), 0644)
}

func makeMigration(appName, migrationName string, force bool) error {
	// Ensure we're in project root by checking for go.mod
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		fmt.Println("Error: Must run from project root (go.mod not found)")
		return fmt.Errorf("not in project root")
	}

	// Check if app exists
	appDir := filepath.Join("apps", appName)
	if _, err := os.Stat(appDir); os.IsNotExist(err) {
		return fmt.Errorf("App '%s' does not exist. Create it with: bourbon create:app %s", appName, appName)
	}

	migrationsDir := filepath.Join(appDir, "migrations")
	if err := os.MkdirAll(migrationsDir, 0755); err != nil {
		return fmt.Errorf("Error creating migrations directory: %v", err)
	}

	// Detect models from models.go
	modelsPath := filepath.Join(appDir, "models.go")
	models, err := detectModels(modelsPath)
	if err != nil {
		return fmt.Errorf("Error parsing models: %v", err)
	}

	// Get project module name
	projectModule, err := getProjectModule()
	if err != nil {
		return fmt.Errorf("Error reading go.mod: %v", err)
	}

	// Get next migration number for this app
	number := getNextMigrationNumber(migrationsDir)

	// Use sequential numbering if no name provided (like Django)
	if migrationName == "" {
		if number == 1 {
			migrationName = "initial"
		} else {
			migrationName = fmt.Sprintf("auto_%04d", number)
		}
	}

	// Generate timestamp-based ID (YYYYMMDDHHmmss format)
	timestamp := generateTimestamp()
	fileName := fmt.Sprintf("%04d_%s.go", number, migrationName)
	path := filepath.Join(migrationsDir, fileName)

	// Generate model definitions and migration calls
	var modelDefs, autoMigrateCalls, tableNames []string

	if len(models) > 0 {
		// Read the actual model definitions from models.go
		modelsContent, _ := os.ReadFile(modelsPath)
		modelsStr := string(modelsContent)

		for _, model := range models {
			// Extract struct definition
			structDef := extractStructDefinition(modelsStr, model.Name)
			if structDef != "" {
				modelDefs = append(modelDefs, structDef)
				autoMigrateCalls = append(autoMigrateCalls, fmt.Sprintf("&%s{}", model.Name))
				// Generate table name (lowercase with underscores)
				tableNames = append(tableNames, fmt.Sprintf("\"%s\"", toSnakeCase(model.Name)))
			}
		}
	}

	hasModelsStr := "false"
	timeImport := ""
	modelMigrationCode := "// Add your migration logic here\n\t\t\treturn nil"
	modelRollbackCode := "// Add your rollback logic here\n\t\t\treturn nil"

	if len(models) > 0 {
		hasModelsStr = "true"
		timeImport = "\n\t\"time\""
		modelMigrationCode = fmt.Sprintf("return tx.AutoMigrate(%s)", strings.Join(autoMigrateCalls, ", "))
		modelRollbackCode = fmt.Sprintf("return tx.Migrator().DropTable(%s)", strings.Join(tableNames, ", "))
	}

	data := map[string]string{
		"AppName":          appName,
		"ProjectModule":    projectModule,
		"MigrationName":    toPascalCase(migrationName),
		"Number":           fmt.Sprintf("%04d", number),
		"NumberInt":        strconv.Itoa(number),
		"FullName":         migrationName,
		"Timestamp":        timestamp,
		"ModelDefinitions": strings.Join(modelDefs, "\n\n\t\t\t"),
		"MigrationCode":    modelMigrationCode,
		"RollbackCode":     modelRollbackCode,
		"HasModels":        hasModelsStr,
		"TimeImport":       timeImport,
	}

	content := renderTemplate(migrationWithModelsTemplate, data)

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("Error writing migration file: %v", err)
	}

	// Save models hash to detect future changes
	if err := saveModelsHash(appName); err != nil {
		fmt.Printf("Warning: Could not save models hash: %v\n", err)
	}

	fmt.Printf("  %s:\n", appName)
	if len(models) > 0 {
		fmt.Printf("    - %s (ID: %s_%s)\n", fileName, timestamp, migrationName)
		fmt.Printf("      Detected %d model(s):\n", len(models))
		for _, model := range models {
			fmt.Printf("      • %s\n", model.Name)
		}
	} else {
		fmt.Printf("    - %s (ID: %s_%s)\n", fileName, timestamp, migrationName)
		fmt.Printf("      No models detected - empty migration created\n")
	}

	return nil
}

func getNextMigrationNumber(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 1
	}

	maxNumber := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}

		// Extract number from filename like "0001_create_users.go"
		name := entry.Name()
		if len(name) >= 4 {
			if num, err := strconv.Atoi(name[:4]); err == nil {
				if num > maxNumber {
					maxNumber = num
				}
			}
		}
	}

	return maxNumber + 1
}

func toPascalCase(s string) string {
	words := strings.Split(s, "_")
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}
	return strings.Join(words, "")
}

func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

func generateTimestamp() string {
	now := time.Now()
	return now.Format("20060102150405")
}

func extractStructDefinition(source, structName string) string {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", source, parser.ParseComments)
	if err != nil {
		return ""
	}

	var structDef string
	ast.Inspect(node, func(n ast.Node) bool {
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok || typeSpec.Name.Name != structName {
			return true
		}

		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			return true
		}

		// Build struct definition
		var fields []string
		fields = append(fields, fmt.Sprintf("type %s struct {", structName))

		for _, field := range structType.Fields.List {
			// Handle embedded fields (no name)
			if len(field.Names) == 0 {
				fieldType := getFieldType(field.Type)

				// If it's models.BaseModel or BaseModel, expand it inline
				if fieldType == "models.BaseModel" || fieldType == "BaseModel" {
					// Expand BaseModel fields inline
					fields = append(fields, "\t\t\t\tID        uint           `gorm:\"primaryKey\" json:\"id\"`")
					fields = append(fields, "\t\t\t\tCreatedAt time.Time      `json:\"created_at\"`")
					fields = append(fields, "\t\t\t\tUpdatedAt time.Time      `json:\"updated_at\"`")
					fields = append(fields, "\t\t\t\tDeletedAt gorm.DeletedAt `gorm:\"index\" json:\"-\"`")
				} else {
					tag := ""
					if field.Tag != nil {
						tag = " " + field.Tag.Value
					}
					fields = append(fields, fmt.Sprintf("\t\t\t\t%s%s", fieldType, tag))
				}
			} else {
				// Named fields
				for _, name := range field.Names {
					fieldType := getFieldType(field.Type)
					tag := ""
					if field.Tag != nil {
						tag = " " + field.Tag.Value
					}
					fields = append(fields, fmt.Sprintf("\t\t\t\t%s %s%s", name.Name, fieldType, tag))
				}
			}
		}
		fields = append(fields, "\t\t\t}")

		structDef = strings.Join(fields, "\n")
		return false
	})

	return structDef
}

func getFieldType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", getFieldType(t.X), t.Sel.Name)
	case *ast.StarExpr:
		return "*" + getFieldType(t.X)
	case *ast.ArrayType:
		return "[]" + getFieldType(t.Elt)
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", getFieldType(t.Key), getFieldType(t.Value))
	default:
		return "interface{}"
	}
}

const modelFileTemplate = `package {{.AppName}}

import (
	"github.com/ishubhamsingh2e/bourbon/bourbon/models"
)

// Example model - uncomment and modify as needed
// type YourModel struct {
// 	models.BaseModel
// 	Name string ` + "`gorm:\"size:255\" json:\"name\"`" + `
// }

`

const controllerFileTemplate = `package {{.AppName}}

import (
	"github.com/ishubhamsingh2e/bourbon/bourbon/core"
	"github.com/ishubhamsingh2e/bourbon/bourbon/http"
)

`

const routesFileTemplate = `package {{.AppName}}

import (
	"github.com/ishubhamsingh2e/bourbon/bourbon/core"
	bourbonHttp "github.com/ishubhamsingh2e/bourbon/bourbon/http"
)

// RegisterRoutes registers all routes for this app under the given prefix
// prefix examples: "/", "/api", "/admin", etc.
func RegisterRoutes(app *core.Application, prefix string) {
	// Create a route group for this app
	group := app.Router.Group(prefix)
	
	// Register your routes here
	// Example:
	// group.Get("/items", listItemsHandler)
	// group.Post("/items", createItemHandler)
	// group.Get("/items/:id", getItemHandler)
}
`

const migrationWithModelsTemplate = `package migrations

import ({{.TimeImport}}
	
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/ishubhamsingh2e/bourbon/bourbon/core"
	"gorm.io/gorm"
)

func init() {
	migration := &gormigrate.Migration{
		ID: "{{.Timestamp}}_{{.FullName}}",
		Migrate: func(tx *gorm.DB) error {
			// Define models inline for this migration
			{{.ModelDefinitions}}

			// Migrate
			{{.MigrationCode}}
		},
		Rollback: func(tx *gorm.DB) error {
			// Rollback
			{{.RollbackCode}}
		},
	}
	
	core.RegisterAppMigration("{{.AppName}}", migration)
}
`
