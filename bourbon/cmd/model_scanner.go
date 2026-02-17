package cmd

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// ModelInfo represents a Go struct model
type ModelInfo struct {
	Name        string
	Fields      []FieldInfo
	PackageName string
	FilePath    string
}

// FieldInfo represents a struct field
type FieldInfo struct {
	Name      string
	Type      string
	Tag       string
	IsPointer bool
}

// ScanModels scans the app directory for model structs
func ScanModels(appName string) ([]ModelInfo, error) {
	modelsPath := filepath.Join("apps", appName, "models.go")

	// Check if models.go exists
	if _, err := os.Stat(modelsPath); os.IsNotExist(err) {
		return []ModelInfo{}, nil // No models yet
	}

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, modelsPath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse models.go: %w", err)
	}

	var models []ModelInfo

	ast.Inspect(node, func(n ast.Node) bool {
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			return true
		}

		// Skip if it embeds BaseModel (it's likely a model)
		hasBaseModel := false
		var fields []FieldInfo

		for _, field := range structType.Fields.List {
			// Check for embedded BaseModel
			if len(field.Names) == 0 {
				if ident, ok := field.Type.(*ast.SelectorExpr); ok {
					if ident.Sel.Name == "BaseModel" {
						hasBaseModel = true
						continue
					}
				}
			}

			// Parse regular fields
			for _, name := range field.Names {
				fieldInfo := FieldInfo{
					Name: name.Name,
					Type: exprToString(field.Type),
				}

				if field.Tag != nil {
					fieldInfo.Tag = field.Tag.Value
				}

				// Check if pointer
				if _, ok := field.Type.(*ast.StarExpr); ok {
					fieldInfo.IsPointer = true
				}

				fields = append(fields, fieldInfo)
			}
		}

		// Only include if it has BaseModel (indicating it's a GORM model)
		if hasBaseModel {
			models = append(models, ModelInfo{
				Name:        typeSpec.Name.Name,
				Fields:      fields,
				PackageName: node.Name.Name,
				FilePath:    modelsPath,
			})
		}

		return true
	})

	return models, nil
}

// exprToString converts an AST expression to a type string
func exprToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + exprToString(t.X)
	case *ast.SelectorExpr:
		return exprToString(t.X) + "." + t.Sel.Name
	case *ast.ArrayType:
		return "[]" + exprToString(t.Elt)
	default:
		return "interface{}"
	}
}

// GenerateMigrationCode generates migration code for models
func GenerateMigrationCode(models []ModelInfo, appName string) string {
	if len(models) == 0 {
		return ""
	}

	var code strings.Builder

	// Generate AutoMigrate call
	code.WriteString("\t\t\t// Auto-migrate models\n")
	code.WriteString("\t\t\treturn db.AutoMigrate(\n")

	for _, model := range models {
		code.WriteString(fmt.Sprintf("\t\t\t\t&%s.%s{},\n", appName, model.Name))
	}

	code.WriteString("\t\t\t)")

	return code.String()
}

// GenerateRollbackCode generates rollback code for models
func GenerateRollbackCode(models []ModelInfo) string {
	if len(models) == 0 {
		return ""
	}

	var code strings.Builder

	code.WriteString("\t\t\t// Drop tables\n")
	code.WriteString("\t\t\treturn db.Migrator().DropTable(\n")

	for _, model := range models {
		code.WriteString(fmt.Sprintf("\t\t\t\t\"%s\",\n", toSnakeCase(model.Name)))
	}

	code.WriteString("\t\t\t)")

	return code.String()
}

// toSnakeCase converts CamelCase to snake_case and pluralizes (GORM convention)
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	tableName := strings.ToLower(result.String())

	// Pluralize (GORM convention)
	// Simple pluralization - add 's' for most cases
	// GORM handles this automatically, so we need to match it
	if !strings.HasSuffix(tableName, "s") {
		tableName += "s"
	}

	return tableName
}

// GetTableNames extracts table names from models
func GetTableNames(models []ModelInfo) []string {
	var tables []string
	for _, model := range models {
		tables = append(tables, toSnakeCase(model.Name))
	}
	return tables
}

// GenerateInlineStructs generates inline struct definitions for migration
func GenerateInlineStructs(models []ModelInfo) string {
	if len(models) == 0 {
		return ""
	}

	var code strings.Builder

	for _, model := range models {
		code.WriteString(fmt.Sprintf("\t\ttype %s struct {\n", model.Name))

		// Add BaseModel fields
		code.WriteString("\t\t\tID        uint      `gorm:\"primarykey\"`\n")
		code.WriteString("\t\t\tCreatedAt time.Time\n")
		code.WriteString("\t\t\tUpdatedAt time.Time\n")
		code.WriteString("\t\t\tDeletedAt gorm.DeletedAt `gorm:\"index\"`\n")

		// Add model-specific fields
		for _, field := range model.Fields {
			tagStr := ""
			if field.Tag != "" {
				tagStr = " " + field.Tag
			}
			code.WriteString(fmt.Sprintf("\t\t\t%s %s%s\n", field.Name, field.Type, tagStr))
		}

		code.WriteString("\t\t}\n")
	}

	return code.String()
}

// GenerateInlineAutoMigrate generates AutoMigrate call with inline types
func GenerateInlineAutoMigrate(models []ModelInfo) string {
	if len(models) == 0 {
		return ""
	}

	var code strings.Builder
	code.WriteString("\t\treturn db.AutoMigrate(\n")

	for _, model := range models {
		code.WriteString(fmt.Sprintf("\t\t\t&%s{},\n", model.Name))
	}

	code.WriteString("\t\t)")

	return code.String()
}

// GenerateInlineDropTable generates DropTable call with inline table names
func GenerateInlineDropTable(models []ModelInfo) string {
	if len(models) == 0 {
		return ""
	}

	var code strings.Builder
	code.WriteString("\t\treturn db.Migrator().DropTable(\n")

	for _, model := range models {
		code.WriteString(fmt.Sprintf("\t\t\t\"%s\",\n", toSnakeCase(model.Name)))
	}

	code.WriteString("\t\t)")

	return code.String()
}

// GenerateDropColumnCode generates code to drop deleted columns
func GenerateDropColumnCode(deletedFields map[string][]string) string {
	if len(deletedFields) == 0 {
		return ""
	}

	var code strings.Builder
	code.WriteString("\t\t// Drop deleted columns\n")

	for modelName, fields := range deletedFields {
		tableName := toSnakeCase(modelName)
		for _, fieldName := range fields {
			columnName := fieldToSnakeCase(fieldName)
			// Use table name in string format for DropColumn
			code.WriteString(fmt.Sprintf("\t\tif err := db.Migrator().DropColumn(\"%s\", \"%s\"); err != nil {\n",
				tableName, columnName))
			code.WriteString("\t\t\treturn err\n")
			code.WriteString("\t\t}\n")
		}
	}

	return code.String()
}

// fieldToSnakeCase converts CamelCase to snake_case WITHOUT pluralization (for columns)
func fieldToSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// GenerateInlineAutoMigrateNoReturn generates AutoMigrate call without return statement
func GenerateInlineAutoMigrateNoReturn(models []ModelInfo) string {
	if len(models) == 0 {
		return ""
	}

	var code strings.Builder
	code.WriteString("\t\tif err := db.AutoMigrate(\n")

	for _, model := range models {
		code.WriteString(fmt.Sprintf("\t\t\t&%s{},\n", model.Name))
	}

	code.WriteString("\t\t); err != nil {\n")
	code.WriteString("\t\t\treturn err\n")
	code.WriteString("\t\t}\n")

	return code.String()
}
