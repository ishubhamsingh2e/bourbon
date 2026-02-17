package cmd

import (
	"fmt"
	"strings"
)

// GenerateMigrationCodeFromChanges generates migration code following gormigrate best practices
func GenerateMigrationCodeFromChanges(changes *MigrationChanges) string {
	var code strings.Builder

	// Generate CreateTable for new models
	for _, model := range changes.NewModels {
		code.WriteString(generateCreateTableCode(model))
		code.WriteString("\n")
	}

	// Generate AddColumn for new fields
	for modelName, fields := range changes.NewFields {
		for _, field := range fields {
			code.WriteString(generateAddColumnCode(modelName, field))
			code.WriteString("\n")
		}
	}

	// Generate DropColumn for deleted fields
	for modelName, fields := range changes.DeletedFields {
		for _, field := range fields {
			code.WriteString(generateDropColumnCode(modelName, field))
			code.WriteString("\n")
		}
	}

	// Generate DropTable for deleted models
	for _, modelName := range changes.DeletedModels {
		code.WriteString(generateDropTableCode(modelName))
		code.WriteString("\n")
	}

	result := code.String()
	if result == "" {
		return "\t\treturn nil"
	}

	return strings.TrimSuffix(result, "\n") + "\n\t\treturn nil"
}

// GenerateRollbackCodeFromChanges generates rollback code
func GenerateRollbackCodeFromChanges(changes *MigrationChanges) string {
	var code strings.Builder

	// Rollback: Drop tables that were created
	for _, model := range changes.NewModels {
		code.WriteString(generateDropTableCode(model.Name))
		code.WriteString("\n")
	}

	// Rollback: Drop columns that were added
	for modelName, fields := range changes.NewFields {
		for _, field := range fields {
			code.WriteString(generateDropColumnCode(modelName, field))
			code.WriteString("\n")
		}
	}

	// Rollback: Add back columns that were dropped
	for modelName, fields := range changes.DeletedFields {
		for _, field := range fields {
			code.WriteString(generateAddColumnCode(modelName, field))
			code.WriteString("\n")
		}
	}

	// Rollback: Create tables that were dropped
	// Note: This is imperfect as we don't have full model definition
	for _, modelName := range changes.DeletedModels {
		code.WriteString(fmt.Sprintf("\t\t// TODO: Recreate table %s\n", modelName))
	}

	result := code.String()
	if result == "" {
		return "\t\treturn nil"
	}

	return strings.TrimSuffix(result, "\n") + "\n\t\treturn nil"
}

// generateCreateTableCode generates CreateTable code with inline struct
func generateCreateTableCode(model ModelInfo) string {
	var code strings.Builder

	// Define minimal struct with all fields
	code.WriteString(fmt.Sprintf("\t\ttype %s struct {\n", fieldToSnakeCase(model.Name)))
	code.WriteString("\t\t\tID        uint      `gorm:\"primarykey\"`\n")
	code.WriteString("\t\t\tCreatedAt time.Time\n")
	code.WriteString("\t\t\tUpdatedAt time.Time\n")
	code.WriteString("\t\t\tDeletedAt gorm.DeletedAt `gorm:\"index\"`\n")

	for _, field := range model.Fields {
		tagStr := ""
		if field.Tag != "" {
			tagStr = " " + field.Tag
		}
		code.WriteString(fmt.Sprintf("\t\t\t%s %s%s\n", field.Name, field.Type, tagStr))
	}

	code.WriteString("\t\t}\n")
	code.WriteString(fmt.Sprintf("\t\tif err := tx.Migrator().CreateTable(&%s{}); err != nil {\n", fieldToSnakeCase(model.Name)))
	code.WriteString("\t\t\treturn err\n")
	code.WriteString("\t\t}")

	return code.String()
}

// generateAddColumnCode generates AddColumn code with minimal struct
func generateAddColumnCode(modelName string, field FieldInfo) string {
	var code strings.Builder

	// Define minimal struct with only the field being added
	code.WriteString(fmt.Sprintf("\t\ttype %s struct {\n", fieldToSnakeCase(modelName)))

	tagStr := ""
	if field.Tag != "" {
		tagStr = " " + field.Tag
	}
	code.WriteString(fmt.Sprintf("\t\t\t%s %s%s\n", field.Name, field.Type, tagStr))
	code.WriteString("\t\t}\n")
	code.WriteString(fmt.Sprintf("\t\tif err := tx.Migrator().AddColumn(&%s{}, \"%s\"); err != nil {\n",
		fieldToSnakeCase(modelName), field.Name))
	code.WriteString("\t\t\treturn err\n")
	code.WriteString("\t\t}")

	return code.String()
}

// generateDropColumnCode generates DropColumn code with minimal struct
func generateDropColumnCode(modelName string, field FieldInfo) string {
	var code strings.Builder

	// Define minimal struct with only the field being dropped
	code.WriteString(fmt.Sprintf("\t\ttype %s struct {\n", fieldToSnakeCase(modelName)))

	tagStr := ""
	if field.Tag != "" {
		tagStr = " " + field.Tag
	}
	code.WriteString(fmt.Sprintf("\t\t\t%s %s%s\n", field.Name, field.Type, tagStr))
	code.WriteString("\t\t}\n")
	code.WriteString(fmt.Sprintf("\t\tif err := tx.Migrator().DropColumn(&%s{}, \"%s\"); err != nil {\n",
		fieldToSnakeCase(modelName), field.Name))
	code.WriteString("\t\t\treturn err\n")
	code.WriteString("\t\t}")

	return code.String()
}

// generateDropTableCode generates DropTable code
func generateDropTableCode(modelName string) string {
	tableName := toSnakeCase(modelName)
	return fmt.Sprintf("\t\tif err := tx.Migrator().DropTable(\"%s\"); err != nil {\n\t\t\treturn err\n\t\t}", tableName)
}
