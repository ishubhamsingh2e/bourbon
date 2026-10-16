// Package schema provides table, column, index, and relationship metadata reflection and DDL generation across SQL dialects.
package schema

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/ishubhamsingh2e/bourbon/database/dialect"
	"github.com/ishubhamsingh2e/bourbon/model"
)

// ColumnDefinition describes a single database column metadata.
type ColumnDefinition struct {
	Name         string
	SQLType      string
	IsPrimaryKey bool
	IsAutoIncr   bool
	IsNotNull    bool
	IsUnique     bool
	DefaultVal   string
	HasDefault   bool
	ForeignKey   *ForeignKeyDefinition
	IsIndex      bool
	IndexName    string
}

// ForeignKeyDefinition describes a foreign key constraint.
type ForeignKeyDefinition struct {
	TargetTable string
	TargetCol   string
	OnDelete    string
	OnUpdate    string
}

// TableDefinition describes a complete table schema with columns, indexes, and relationships.
type TableDefinition struct {
	TableName   string
	Columns     []ColumnDefinition
	ForeignKeys []ForeignKeyDefinition
}

// InspectModel inspects a Go struct model and extracts its complete table schema definition.
func InspectModel(d dialect.Dialect, v any) (*TableDefinition, error) {
	if v == nil {
		return nil, fmt.Errorf("schema: cannot inspect nil model")
	}

	typ := reflect.TypeOf(v)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf("schema: expected struct, got %s", typ.Kind())
	}

	tableName := ""
	if m, ok := v.(model.Model); ok && m.TableName() != "" {
		tableName = m.TableName()
	} else if inst := reflect.New(typ).Interface(); inst != nil {
		if m, ok := inst.(model.Model); ok && m.TableName() != "" {
			tableName = m.TableName()
		}
	}
	if tableName == "" {
		tableName = model.Inspect(v).TableName
	}

	table := &TableDefinition{
		TableName: tableName,
		Columns:   make([]ColumnDefinition, 0),
	}

	inspectFields(d, typ, table)
	return table, nil
}

func inspectFields(d dialect.Dialect, typ reflect.Type, table *TableDefinition) {
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if !field.IsExported() {
			continue
		}

		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			inspectFields(d, field.Type, table)
			continue
		}

		dbTag := field.Tag.Get("db")
		bourbonTag := field.Tag.Get("bourbon")
		if bourbonTag == "" {
			bourbonTag = field.Tag.Get("bun")
		}

		if dbTag == "-" || bourbonTag == "-" {
			continue
		}

		// Skip relation slices or pointer structs from being created as columns
		if isRelationField(field.Type, bourbonTag) {
			continue
		}

		col := ColumnDefinition{
			Name: toSnakeCase(field.Name),
		}

		// Parse db tag for column name
		if dbTag != "" {
			parts := strings.Split(dbTag, ",")
			if parts[0] != "" {
				col.Name = parts[0]
			}
			for _, opt := range parts[1:] {
				opt = strings.TrimSpace(opt)
				switch opt {
				case "primary", "pk":
					col.IsPrimaryKey = true
				case "not_null", "notnull":
					col.IsNotNull = true
				case "unique":
					col.IsUnique = true
				case "autoincrement", "auto_increment":
					col.IsAutoIncr = true
				case "index":
					col.IsIndex = true
				}
				if strings.HasPrefix(opt, "references:") {
					col.ForeignKey = parseForeignKey(opt, parts[1:]...)
				}
			}
		}

		// Parse bourbon / bun tags
		if bourbonTag != "" {
			parts := strings.Split(bourbonTag, ",")
			for _, opt := range parts {
				opt = strings.TrimSpace(opt)
				switch {
				case opt == "pk" || opt == "primary":
					col.IsPrimaryKey = true
				case opt == "notnull" || opt == "not_null":
					col.IsNotNull = true
				case opt == "unique":
					col.IsUnique = true
				case opt == "autoincrement" || opt == "auto_increment":
					col.IsAutoIncr = true
				case opt == "index":
					col.IsIndex = true
				case strings.HasPrefix(opt, "type:"):
					col.SQLType = strings.TrimPrefix(opt, "type:")
				case strings.HasPrefix(opt, "default:"):
					col.DefaultVal = strings.TrimPrefix(opt, "default:")
					col.HasDefault = true
				case strings.HasPrefix(opt, "references:"):
					col.ForeignKey = parseForeignKey(opt, parts...)
				case strings.HasPrefix(opt, "on_delete:"):
					if col.ForeignKey != nil {
						col.ForeignKey.OnDelete = strings.ToUpper(strings.TrimPrefix(opt, "on_delete:"))
					}
				case strings.HasPrefix(opt, "on_update:"):
					if col.ForeignKey != nil {
						col.ForeignKey.OnUpdate = strings.ToUpper(strings.TrimPrefix(opt, "on_update:"))
					}
				}
			}
		}

		if col.SQLType == "" {
			col.SQLType = inferSQLType(d, field.Type, col.IsPrimaryKey)
		}

		// Auto primary key constraint for id
		if strings.EqualFold(col.Name, "id") {
			col.IsPrimaryKey = true
		}

		table.Columns = append(table.Columns, col)
	}
}

func isRelationField(t reflect.Type, tag string) bool {
	if strings.Contains(tag, "rel:") {
		return true
	}
	if t.Kind() == reflect.Slice {
		elem := t.Elem()
		if elem.Kind() == reflect.Struct || (elem.Kind() == reflect.Ptr && elem.Elem().Kind() == reflect.Struct) {
			return true
		}
	}
	if t.Kind() == reflect.Ptr {
		elem := t.Elem()
		if elem.Kind() == reflect.Struct && elem != reflect.TypeOf(time.Time{}) {
			return true
		}
	}
	return false
}

func inferSQLType(d dialect.Dialect, t reflect.Type, isPK bool) string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t == reflect.TypeOf(time.Time{}) {
		return "TIMESTAMP"
	}

	switch t.Kind() {
	case reflect.String:
		if isPK {
			return d.UUIDType()
		}
		return "VARCHAR(255)"
	case reflect.Bool:
		return d.BooleanType()
	case reflect.Int, reflect.Int32, reflect.Uint, reflect.Uint32:
		return "INTEGER"
	case reflect.Int64, reflect.Uint64:
		return "BIGINT"
	case reflect.Float32:
		return "FLOAT"
	case reflect.Float64:
		return "DOUBLE PRECISION"
	case reflect.Slice:
		if t.Elem().Kind() == reflect.Uint8 {
			return "BLOB"
		}
		return d.JSONType()
	case reflect.Map, reflect.Struct:
		return d.JSONType()
	default:
		return "TEXT"
	}
}

func parseForeignKey(ref string, allOptions ...string) *ForeignKeyDefinition {
	raw := strings.TrimPrefix(ref, "references:")
	// Expected format: table(col)
	open := strings.Index(raw, "(")
	close := strings.Index(raw, ")")
	if open == -1 || close == -1 || close <= open {
		return nil
	}

	targetTable := raw[:open]
	targetCol := raw[open+1 : close]

	fk := &ForeignKeyDefinition{
		TargetTable: targetTable,
		TargetCol:   targetCol,
	}

	for _, opt := range allOptions {
		opt = strings.TrimSpace(opt)
		if strings.HasPrefix(opt, "on_delete:") {
			fk.OnDelete = strings.ToUpper(strings.TrimPrefix(opt, "on_delete:"))
		} else if strings.HasPrefix(opt, "on_update:") {
			fk.OnUpdate = strings.ToUpper(strings.TrimPrefix(opt, "on_update:"))
		}
	}

	return fk
}

// GenerateCreateTableSQL generates the complete CREATE TABLE and index DDL for a model under a specific dialect.
func GenerateCreateTableSQL(d dialect.Dialect, table *TableDefinition) string {
	var sb strings.Builder
	qTable := d.QuoteIdentifier(table.TableName)

	sb.WriteString(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n", qTable))

	var colDefs []string
	for _, col := range table.Columns {
		var parts []string
		parts = append(parts, fmt.Sprintf("    %s", d.QuoteIdentifier(col.Name)))

		colType := col.SQLType
		if col.IsPrimaryKey && col.IsAutoIncr {
			colType = d.AutoIncrementKeyword()
		}
		parts = append(parts, colType)

		if col.IsPrimaryKey {
			parts = append(parts, "PRIMARY KEY")
		}
		if col.IsNotNull && !col.IsPrimaryKey {
			parts = append(parts, "NOT NULL")
		}
		if col.IsUnique && !col.IsPrimaryKey {
			parts = append(parts, "UNIQUE")
		}
		if col.HasDefault {
			parts = append(parts, fmt.Sprintf("DEFAULT %s", col.DefaultVal))
		}

		if col.ForeignKey != nil {
			fkClause := fmt.Sprintf("REFERENCES %s(%s)",
				d.QuoteIdentifier(col.ForeignKey.TargetTable),
				d.QuoteIdentifier(col.ForeignKey.TargetCol),
			)
			if col.ForeignKey.OnDelete != "" {
				fkClause += fmt.Sprintf(" ON DELETE %s", col.ForeignKey.OnDelete)
			}
			if col.ForeignKey.OnUpdate != "" {
				fkClause += fmt.Sprintf(" ON UPDATE %s", col.ForeignKey.OnUpdate)
			}
			parts = append(parts, fkClause)
		}

		colDefs = append(colDefs, strings.Join(parts, " "))
	}

	sb.WriteString(strings.Join(colDefs, ",\n"))
	sb.WriteString("\n);")

	// Generate Indexes
	for _, col := range table.Columns {
		if col.IsIndex && !col.IsPrimaryKey && !col.IsUnique {
			idxName := fmt.Sprintf("idx_%s_%s", table.TableName, col.Name)
			sb.WriteString(fmt.Sprintf("\nCREATE INDEX IF NOT EXISTS %s ON %s (%s);",
				d.QuoteIdentifier(idxName),
				qTable,
				d.QuoteIdentifier(col.Name),
			))
		}
	}

	return sb.String()
}

// GenerateDropTableSQL generates DROP TABLE and index drop statements.
func GenerateDropTableSQL(d dialect.Dialect, table *TableDefinition) string {
	var sb strings.Builder
	qTable := d.QuoteIdentifier(table.TableName)

	for _, col := range table.Columns {
		if col.IsIndex && !col.IsPrimaryKey && !col.IsUnique {
			idxName := fmt.Sprintf("idx_%s_%s", table.TableName, col.Name)
			sb.WriteString(fmt.Sprintf("DROP INDEX IF EXISTS %s;\n", d.QuoteIdentifier(idxName)))
		}
	}

	sb.WriteString(fmt.Sprintf("DROP TABLE IF EXISTS %s;", qTable))
	return sb.String()
}

func toSnakeCase(s string) string {
	if s == "ID" {
		return "id"
	}
	var res []rune
	runes := []rune(s)
	for i, r := range runes {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				prev := runes[i-1]
				nextIsLower := i+1 < len(runes) && (runes[i+1] >= 'a' && runes[i+1] <= 'z')
				if (prev >= 'a' && prev <= 'z') || ((prev >= 'A' && prev <= 'Z') && nextIsLower) {
					res = append(res, '_')
				}
			}
			res = append(res, r+32)
		} else {
			res = append(res, r)
		}
	}
	return string(res)
}
