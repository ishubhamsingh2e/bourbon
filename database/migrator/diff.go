// Package migrator provides embedded database migration management, schema versioning, and migration file scaffolding for Bourbon.
package migrator

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ishubhamsingh2e/bourbon/database/dialect"
)

// Prompter provides an interface for asking interactive questions during makemigrations.
type Prompter interface {
	PromptBool(question string, defaultVal bool) bool
	PromptString(question string, defaultVal string) string
}

// DefaultCLIPrompter reads from stdin and writes to stdout.
type DefaultCLIPrompter struct {
	Reader io.Reader
	Writer io.Writer
}

// NewDefaultCLIPrompter creates a standard terminal prompter.
func NewDefaultCLIPrompter() *DefaultCLIPrompter {
	return &DefaultCLIPrompter{
		Reader: os.Stdin,
		Writer: os.Stdout,
	}
}

// PromptBool asks a [y/N] or [Y/n] question.
func (p *DefaultCLIPrompter) PromptBool(question string, defaultVal bool) bool {
	hint := "[y/N]"
	if defaultVal {
		hint = "[Y/n]"
	}
	fmt.Fprintf(p.Writer, "%s %s: ", question, hint)
	scanner := bufio.NewScanner(p.Reader)
	if scanner.Scan() {
		input := strings.ToLower(strings.TrimSpace(scanner.Text()))
		if input == "" {
			return defaultVal
		}
		return input == "y" || input == "yes"
	}
	return defaultVal
}

// PromptString asks a question and returns the string response.
func (p *DefaultCLIPrompter) PromptString(question string, defaultVal string) string {
	if defaultVal != "" {
		fmt.Fprintf(p.Writer, "%s [%s]: ", question, defaultVal)
	} else {
		fmt.Fprintf(p.Writer, "%s: ", question)
	}
	scanner := bufio.NewScanner(p.Reader)
	if scanner.Scan() {
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			return defaultVal
		}
		return input
	}
	return defaultVal
}

// NonInteractivePrompter always returns defaults without prompting.
type NonInteractivePrompter struct {
	AutoYes bool
}

// PromptBool returns auto-yes or the default value.
func (p *NonInteractivePrompter) PromptBool(question string, defaultVal bool) bool {
	if p.AutoYes {
		return true
	}
	return defaultVal
}

// PromptString returns the default value.
func (p *NonInteractivePrompter) PromptString(question string, defaultVal string) string {
	return defaultVal
}

// FieldAST describes an AST-parsed field.
type FieldAST struct {
	Name         string `json:"name"`
	ColumnName   string `json:"column_name"`
	Type         string `json:"type"`
	SQLType      string `json:"sql_type"`
	Tag          string `json:"tag"`
	IsPrimaryKey bool   `json:"is_primary_key"`
	IsNotNull    bool   `json:"is_not_null"`
	IsUnique     bool   `json:"is_unique"`
	DefaultVal   string `json:"default_val"`
	HasDefault   bool   `json:"has_default"`
	RenamedFrom  string `json:"renamed_from,omitempty"`
}

// ModelAST describes an AST-parsed model struct.
type ModelAST struct {
	Name      string     `json:"name"`
	TableName string     `json:"table_name"`
	App       string     `json:"app"`
	Fields    []FieldAST `json:"fields"`
	FilePath  string     `json:"file_path"`
}

// AppState stores the snapshot of an app's models.
type AppState struct {
	LastHash      string               `json:"last_hash"`
	LastMigration string               `json:"last_migration"`
	Models        map[string]*ModelAST `json:"models"`
}

// MigrationState stores project-wide schema snapshots across apps.
type MigrationState struct {
	Apps map[string]*AppState `json:"apps"`
}

// LoadMigrationState loads the persisted schema state from .bourbon/migration_state.json.
func LoadMigrationState(rootDir string) (*MigrationState, error) {
	statePath := filepath.Join(rootDir, ".bourbon", "migration_state.json")
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		return &MigrationState{
			Apps: make(map[string]*AppState),
		}, nil
	}

	data, err := os.ReadFile(statePath)
	if err != nil {
		return nil, fmt.Errorf("failed reading migration state: %w", err)
	}

	var state MigrationState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed parsing migration state: %w", err)
	}
	if state.Apps == nil {
		state.Apps = make(map[string]*AppState)
	}
	return &state, nil
}

// SaveMigrationState persists the schema state to .bourbon/migration_state.json.
func SaveMigrationState(rootDir string, state *MigrationState) error {
	stateDir := filepath.Join(rootDir, ".bourbon")
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return fmt.Errorf("failed creating .bourbon directory: %w", err)
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed serializing migration state: %w", err)
	}
	statePath := filepath.Join(stateDir, "migration_state.json")
	return os.WriteFile(statePath, data, 0644)
}

// ScanProjectModels scans the project directories to extract all model declarations.
func ScanProjectModels(rootDir string, targetApp string, d dialect.Dialect) ([]ModelAST, error) {
	if d == nil {
		d, _ = dialect.Get("sqlite")
	}

	var models []ModelAST
	searchDirs := []string{
		filepath.Join(rootDir, "apps"),
		filepath.Join(rootDir, "internal", "modules"),
		filepath.Join(rootDir, "internal", "models"),
		filepath.Join(rootDir, "models"),
	}

	for _, dir := range searchDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}

		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}

			// Determine app name from path relative to search dir
			rel, _ := filepath.Rel(rootDir, path)
			parts := strings.Split(filepath.ToSlash(rel), "/")
			appName := "app"
			if len(parts) >= 2 && (parts[0] == "apps" || (parts[0] == "internal" && parts[1] == "modules")) {
				if parts[0] == "apps" {
					appName = parts[1]
				} else if len(parts) >= 3 {
					appName = parts[2]
				}
			} else if len(parts) >= 2 {
				appName = parts[len(parts)-2]
			}

			if targetApp != "" && appName != targetApp {
				return nil
			}

			parsedModels, err := parseModelsFromFile(path, appName, d)
			if err == nil {
				models = append(models, parsedModels...)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return models, nil
}

func parseModelsFromFile(filePath string, appName string, d dialect.Dialect) ([]ModelAST, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var models []ModelAST

	// First find any TableName() method overrides in this file
	tableNames := make(map[string]string)
	for _, decl := range node.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv == nil || len(fn.Recv.List) == 0 || fn.Name.Name != "TableName" {
			continue
		}
		recvType := exprToTypeName(fn.Recv.List[0].Type)
		if fn.Body != nil && len(fn.Body.List) > 0 {
			if ret, ok := fn.Body.List[len(fn.Body.List)-1].(*ast.ReturnStmt); ok && len(ret.Results) == 1 {
				if lit, ok := ret.Results[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
					tableNames[recvType] = strings.Trim(lit.Value, `"`)
				}
			}
		}
	}

	// Parse struct type declarations
	for _, decl := range node.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			structName := typeSpec.Name.Name
			hasBaseModel := false
			var fields []FieldAST

			for _, field := range structType.Fields.List {
				// Check embedded BaseModel
				if len(field.Names) == 0 {
					typeName := exprToTypeName(field.Type)
					if typeName == "BaseModel" || strings.HasSuffix(typeName, ".BaseModel") {
						hasBaseModel = true
						fields = append(fields,
							FieldAST{
								Name:         "ID",
								ColumnName:   "id",
								Type:         "string",
								SQLType:      d.UUIDType(),
								IsPrimaryKey: true,
								IsNotNull:    true,
							},
							FieldAST{
								Name:       "CreatedAt",
								ColumnName: "created_at",
								Type:       "time.Time",
								SQLType:    "TIMESTAMP WITH TIME ZONE",
								HasDefault: true,
								DefaultVal: "CURRENT_TIMESTAMP",
							},
							FieldAST{
								Name:       "UpdatedAt",
								ColumnName: "updated_at",
								Type:       "time.Time",
								SQLType:    "TIMESTAMP WITH TIME ZONE",
								HasDefault: true,
								DefaultVal: "CURRENT_TIMESTAMP",
							},
						)
					}
					continue
				}

				tagValue := ""
				if field.Tag != nil {
					tagValue = strings.Trim(field.Tag.Value, "`")
				}

				for _, name := range field.Names {
					if !ast.IsExported(name.Name) {
						continue
					}

					goType := exprToTypeName(field.Type)
					fast := parseFieldAST(name.Name, goType, tagValue, d)
					if fast != nil {
						fields = append(fields, *fast)
					}
				}
			}

			// Only treat as model if it embeds BaseModel or has TableName() method or fields with db tags
			tableName := tableNames[structName]
			if tableName == "" {
				tableName = deriveTableName(toSnakeCase(structName))
			}

			isLikelyModel := hasBaseModel || tableNames[structName] != "" || hasDBTags(fields)
			if isLikelyModel && len(fields) > 0 {
				models = append(models, ModelAST{
					Name:      structName,
					TableName: tableName,
					App:       appName,
					Fields:    fields,
					FilePath:  filePath,
				})
			}
		}
	}

	return models, nil
}

func hasDBTags(fields []FieldAST) bool {
	for _, f := range fields {
		if strings.Contains(f.Tag, "db:") || strings.Contains(f.Tag, "bourbon:") {
			return true
		}
	}
	return false
}

func parseFieldAST(name string, goType string, tag string, d dialect.Dialect) *FieldAST {
	dbTag := getTagValue(tag, "db")
	bourbonTag := getTagValue(tag, "bourbon")

	if dbTag == "-" || bourbonTag == "-" {
		return nil
	}

	colName := toSnakeCase(name)
	fast := &FieldAST{
		Name:       name,
		ColumnName: colName,
		Type:       goType,
		Tag:        tag,
	}

	// Parse db tag
	if dbTag != "" {
		parts := strings.Split(dbTag, ",")
		if parts[0] != "" {
			fast.ColumnName = parts[0]
		}
		for _, opt := range parts[1:] {
			opt = strings.TrimSpace(opt)
			switch opt {
			case "primary", "pk":
				fast.IsPrimaryKey = true
			case "not_null", "notnull":
				fast.IsNotNull = true
			case "unique":
				fast.IsUnique = true
			}
			if strings.HasPrefix(opt, "old_name:") {
				fast.RenamedFrom = strings.TrimPrefix(opt, "old_name:")
			}
		}
	}

	// Parse bourbon tag
	if bourbonTag != "" {
		parts := strings.Split(bourbonTag, ",")
		for _, opt := range parts {
			opt = strings.TrimSpace(opt)
			switch {
			case opt == "pk" || opt == "primary":
				fast.IsPrimaryKey = true
			case opt == "notnull" || opt == "not_null":
				fast.IsNotNull = true
			case opt == "unique":
				fast.IsUnique = true
			case strings.HasPrefix(opt, "type:"):
				fast.SQLType = strings.TrimPrefix(opt, "type:")
			case strings.HasPrefix(opt, "default:"):
				fast.DefaultVal = strings.TrimPrefix(opt, "default:")
				fast.HasDefault = true
			case strings.HasPrefix(opt, "renamed_from:") || strings.HasPrefix(opt, "old_name:"):
				fast.RenamedFrom = strings.TrimPrefix(strings.TrimPrefix(opt, "renamed_from:"), "old_name:")
			}
		}
	}

	if fast.ColumnName == "id" {
		fast.IsPrimaryKey = true
		fast.IsNotNull = true
	}

	if fast.SQLType == "" {
		fast.SQLType = inferASTSQLType(d, goType, fast.IsPrimaryKey)
	}

	return fast
}

func getTagValue(tag string, key string) string {
	prefix := key + `:"`
	idx := strings.Index(tag, prefix)
	if idx == -1 {
		return ""
	}
	rem := tag[idx+len(prefix):]
	end := strings.Index(rem, `"`)
	if end == -1 {
		return ""
	}
	return rem[:end]
}

func inferASTSQLType(d dialect.Dialect, goType string, isPK bool) string {
	cleanType := strings.TrimPrefix(goType, "*")
	switch cleanType {
	case "string":
		if isPK {
			return d.UUIDType()
		}
		return "VARCHAR(255)"
	case "bool":
		return d.BooleanType()
	case "int", "int32", "uint", "uint32":
		return "INTEGER"
	case "int64", "uint64":
		return "BIGINT"
	case "float32":
		return "FLOAT"
	case "float64":
		return "DOUBLE PRECISION"
	case "time.Time":
		return "TIMESTAMP WITH TIME ZONE"
	case "[]byte":
		return "BLOB"
	default:
		if strings.HasPrefix(cleanType, "[]") || strings.HasPrefix(cleanType, "map[") {
			return d.JSONType()
		}
		return "TEXT"
	}
}

func exprToTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + exprToTypeName(t.X)
	case *ast.SelectorExpr:
		return exprToTypeName(t.X) + "." + t.Sel.Name
	case *ast.ArrayType:
		return "[]" + exprToTypeName(t.Elt)
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", exprToTypeName(t.Key), exprToTypeName(t.Value))
	default:
		return "any"
	}
}

// ComputeModelsHash computes a deterministic SHA256 hash for a list of models.
func ComputeModelsHash(models []ModelAST) string {
	h := sha256.New()
	sorted := make([]ModelAST, len(models))
	copy(sorted, models)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Name < sorted[j].Name
	})

	for _, m := range sorted {
		h.Write([]byte(m.Name + ":" + m.TableName + "\n"))
		for _, f := range m.Fields {
			h.Write(fmt.Appendf(nil, "  %s:%s:%s:%s:%t:%t\n", f.Name, f.ColumnName, f.Type, f.SQLType, f.IsPrimaryKey, f.IsNotNull))
		}
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

// ModelChanges represents detected schema differences.
type ModelChanges struct {
	AppName         string
	CreatedModels   []ModelAST
	DeletedModels   []string                     // table names
	RenamedModels   map[string]string            // oldTable -> newTable
	AddedColumns    map[string][]FieldAST        // tableName -> added fields
	DeletedColumns  map[string][]FieldAST        // tableName -> deleted fields
	RenamedColumns  map[string]map[string]string // tableName -> (oldCol -> newCol)
	ModifiedColumns map[string][]FieldAST        // tableName -> modified fields
	Descriptions    []string
}

// HasChanges returns true if any schema diff exists.
func (mc *ModelChanges) HasChanges() bool {
	return len(mc.CreatedModels) > 0 ||
		len(mc.DeletedModels) > 0 ||
		len(mc.RenamedModels) > 0 ||
		len(mc.AddedColumns) > 0 ||
		len(mc.DeletedColumns) > 0 ||
		len(mc.RenamedColumns) > 0 ||
		len(mc.ModifiedColumns) > 0
}

// DetectSchemaChanges compares stored state with scanned models, interactively resolving renames.
func DetectSchemaChanges(appName string, currentModels []ModelAST, previousState *AppState, prompter Prompter, d dialect.Dialect) (*ModelChanges, error) {
	if prompter == nil {
		prompter = &NonInteractivePrompter{AutoYes: false}
	}
	if d == nil {
		d, _ = dialect.Get("sqlite")
	}

	changes := &ModelChanges{
		AppName:         appName,
		CreatedModels:   make([]ModelAST, 0),
		DeletedModels:   make([]string, 0),
		RenamedModels:   make(map[string]string),
		AddedColumns:    make(map[string][]FieldAST),
		DeletedColumns:  make(map[string][]FieldAST),
		RenamedColumns:  make(map[string]map[string]string),
		ModifiedColumns: make(map[string][]FieldAST),
		Descriptions:    make([]string, 0),
	}

	if previousState == nil || len(previousState.Models) == 0 {
		for _, m := range currentModels {
			changes.CreatedModels = append(changes.CreatedModels, m)
			changes.Descriptions = append(changes.Descriptions, fmt.Sprintf("Create model %s (table %s)", m.Name, m.TableName))
		}
		return changes, nil
	}

	storedModels := make(map[string]*ModelAST)
	for k, v := range previousState.Models {
		// Copy stored model
		cp := *v
		cp.Fields = make([]FieldAST, len(v.Fields))
		copy(cp.Fields, v.Fields)
		storedModels[k] = &cp
	}

	currentModelMap := make(map[string]*ModelAST)
	for i := range currentModels {
		currentModelMap[currentModels[i].Name] = &currentModels[i]
	}

	// 1. Detect Model / Table Renames & New Models
	for _, current := range currentModels {
		if stored, exists := storedModels[current.Name]; exists {
			// Check if table name changed on existing model
			if stored.TableName != current.TableName {
				q := fmt.Sprintf("Did you rename table %s to %s?", stored.TableName, current.TableName)
				if prompter.PromptBool(q, false) {
					changes.RenamedModels[stored.TableName] = current.TableName
					changes.Descriptions = append(changes.Descriptions, fmt.Sprintf("Rename table %s to %s", stored.TableName, current.TableName))
				}
			}
			continue
		}

		// Current model is not in stored by name. Check if an old stored model was renamed.
		foundRename := false
		for storedName, stored := range storedModels {
			if _, stillExists := currentModelMap[storedName]; !stillExists {
				q := fmt.Sprintf("Did you rename the %s model to %s?", storedName, current.Name)
				if prompter.PromptBool(q, false) {
					changes.RenamedModels[stored.TableName] = current.TableName
					changes.Descriptions = append(changes.Descriptions, fmt.Sprintf("Rename model %s to %s", storedName, current.Name))
					foundRename = true
					delete(storedModels, storedName)
					storedModels[current.Name] = stored // map to compare fields
					break
				}
			}
		}

		if !foundRename {
			changes.CreatedModels = append(changes.CreatedModels, current)
			changes.Descriptions = append(changes.Descriptions, fmt.Sprintf("Create model %s", current.Name))
		}
	}

	// 2. Detect Deleted Models
	for storedName, stored := range storedModels {
		if _, exists := currentModelMap[storedName]; !exists {
			changes.DeletedModels = append(changes.DeletedModels, stored.TableName)
			changes.Descriptions = append(changes.Descriptions, fmt.Sprintf("Delete model %s (table %s)", storedName, stored.TableName))
		}
	}

	// 3. Detect Field Changes (Add, Delete, Rename, Modify)
	for _, current := range currentModels {
		stored, exists := storedModels[current.Name]
		if !exists {
			continue // Already in CreatedModels
		}

		tableName := current.TableName
		storedFieldMap := make(map[string]FieldAST)
		for _, f := range stored.Fields {
			storedFieldMap[f.ColumnName] = f
		}

		currentFieldMap := make(map[string]FieldAST)
		for _, f := range current.Fields {
			currentFieldMap[f.ColumnName] = f
		}

		var candidateAdded []FieldAST
		var candidateDeleted []FieldAST

		for _, cf := range current.Fields {
			if sf, ok := storedFieldMap[cf.ColumnName]; ok {
				// Field exists by column name, check if type/constraint changed
				if sf.SQLType != cf.SQLType || sf.IsNotNull != cf.IsNotNull || sf.IsUnique != cf.IsUnique {
					changes.ModifiedColumns[tableName] = append(changes.ModifiedColumns[tableName], cf)
					changes.Descriptions = append(changes.Descriptions, fmt.Sprintf("Alter field %s on %s", cf.Name, current.Name))
				}
			} else if cf.RenamedFrom != "" {
				// Explicit tag rename
				if changes.RenamedColumns[tableName] == nil {
					changes.RenamedColumns[tableName] = make(map[string]string)
				}
				changes.RenamedColumns[tableName][cf.RenamedFrom] = cf.ColumnName
				changes.Descriptions = append(changes.Descriptions, fmt.Sprintf("Rename field %s.%s to %s.%s", current.Name, cf.RenamedFrom, current.Name, cf.ColumnName))
				delete(storedFieldMap, cf.RenamedFrom)
			} else {
				candidateAdded = append(candidateAdded, cf)
			}
		}

		for _, sf := range stored.Fields {
			if _, ok := currentFieldMap[sf.ColumnName]; !ok {
				candidateDeleted = append(candidateDeleted, sf)
			}
		}

		// Prompt for field renames
		for _, added := range candidateAdded {
			matched := false
			for i, deleted := range candidateDeleted {
				q := fmt.Sprintf("Did you rename %s.%s to %s.%s?", current.Name, deleted.ColumnName, current.Name, added.ColumnName)
				if prompter.PromptBool(q, false) {
					if changes.RenamedColumns[tableName] == nil {
						changes.RenamedColumns[tableName] = make(map[string]string)
					}
					changes.RenamedColumns[tableName][deleted.ColumnName] = added.ColumnName
					changes.Descriptions = append(changes.Descriptions, fmt.Sprintf("Rename field %s.%s to %s.%s", current.Name, deleted.ColumnName, current.Name, added.ColumnName))
					// Remove from candidate deleted
					candidateDeleted = append(candidateDeleted[:i], candidateDeleted[i+1:]...)
					matched = true
					break
				}
			}

			if !matched {
				// Check if non-nullable without default
				if added.IsNotNull && !added.HasDefault && !added.IsPrimaryKey {
					defVal := prompter.PromptString(fmt.Sprintf("Adding non-nullable field '%s' without default to '%s'. Provide default value", added.ColumnName, tableName), "''")
					if defVal != "" {
						added.DefaultVal = defVal
						added.HasDefault = true
					}
				}
				changes.AddedColumns[tableName] = append(changes.AddedColumns[tableName], added)
				changes.Descriptions = append(changes.Descriptions, fmt.Sprintf("Add field %s to %s", added.Name, current.Name))
			}
		}

		for _, deleted := range candidateDeleted {
			changes.DeletedColumns[tableName] = append(changes.DeletedColumns[tableName], deleted)
			changes.Descriptions = append(changes.Descriptions, fmt.Sprintf("Remove field %s from %s", deleted.Name, current.Name))
		}
	}

	return changes, nil
}

// GenerateSQLFromChanges renders dialect-aware Up and Down SQL strings.
func GenerateSQLFromChanges(changes *ModelChanges, d dialect.Dialect) (upSQL string, downSQL string) {
	if d == nil {
		d, _ = dialect.Get("sqlite")
	}

	var upSB strings.Builder
	var downSB strings.Builder

	// 1. Renamed Models / Tables
	for oldTable, newTable := range changes.RenamedModels {
		upSB.WriteString(fmt.Sprintf("ALTER TABLE %s RENAME TO %s;\n", d.QuoteIdentifier(oldTable), d.QuoteIdentifier(newTable)))
		downSB.WriteString(fmt.Sprintf("ALTER TABLE %s RENAME TO %s;\n", d.QuoteIdentifier(newTable), d.QuoteIdentifier(oldTable)))
	}

	// 2. Created Models
	for _, m := range changes.CreatedModels {
		qTable := d.QuoteIdentifier(m.TableName)
		upSB.WriteString(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n", qTable))

		var colDefs []string
		for _, f := range m.Fields {
			parts := []string{fmt.Sprintf("    %s %s", d.QuoteIdentifier(f.ColumnName), f.SQLType)}
			if f.IsPrimaryKey {
				parts = append(parts, "PRIMARY KEY")
			}
			if f.IsNotNull && !f.IsPrimaryKey {
				parts = append(parts, "NOT NULL")
			}
			if f.IsUnique && !f.IsPrimaryKey {
				parts = append(parts, "UNIQUE")
			}
			if f.HasDefault {
				parts = append(parts, fmt.Sprintf("DEFAULT %s", f.DefaultVal))
			}
			colDefs = append(colDefs, strings.Join(parts, " "))
		}
		upSB.WriteString(strings.Join(colDefs, ",\n"))
		upSB.WriteString("\n);\n")

		downSB.WriteString(fmt.Sprintf("DROP TABLE IF EXISTS %s;\n", qTable))
	}

	// 3. Renamed Columns
	for tableName, renames := range changes.RenamedColumns {
		qTable := d.QuoteIdentifier(tableName)
		for oldCol, newCol := range renames {
			upSB.WriteString(fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s;\n", qTable, d.QuoteIdentifier(oldCol), d.QuoteIdentifier(newCol)))
			downSB.WriteString(fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s;\n", qTable, d.QuoteIdentifier(newCol), d.QuoteIdentifier(oldCol)))
		}
	}

	// 4. Added Columns
	for tableName, fields := range changes.AddedColumns {
		qTable := d.QuoteIdentifier(tableName)
		for _, f := range fields {
			colDef := fmt.Sprintf("%s %s", d.QuoteIdentifier(f.ColumnName), f.SQLType)
			if f.IsNotNull {
				colDef += " NOT NULL"
			}
			if f.HasDefault {
				colDef += fmt.Sprintf(" DEFAULT %s", f.DefaultVal)
			}
			upSB.WriteString(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s;\n", qTable, colDef))
			downSB.WriteString(fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s;\n", qTable, d.QuoteIdentifier(f.ColumnName)))
		}
	}

	// 5. Modified Columns
	for tableName, fields := range changes.ModifiedColumns {
		qTable := d.QuoteIdentifier(tableName)
		for _, f := range fields {
			if d.Name() == "mysql" {
				upSB.WriteString(fmt.Sprintf("ALTER TABLE %s MODIFY COLUMN %s %s;\n", qTable, d.QuoteIdentifier(f.ColumnName), f.SQLType))
			} else {
				upSB.WriteString(fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE %s;\n", qTable, d.QuoteIdentifier(f.ColumnName), f.SQLType))
			}
		}
	}

	// 6. Deleted Columns
	for tableName, fields := range changes.DeletedColumns {
		qTable := d.QuoteIdentifier(tableName)
		for _, f := range fields {
			upSB.WriteString(fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s;\n", qTable, d.QuoteIdentifier(f.ColumnName)))
			// Down rollback re-adds the column
			colDef := fmt.Sprintf("%s %s", d.QuoteIdentifier(f.ColumnName), f.SQLType)
			if f.IsNotNull {
				colDef += " NOT NULL"
			}
			if f.HasDefault {
				colDef += fmt.Sprintf(" DEFAULT %s", f.DefaultVal)
			}
			downSB.WriteString(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s;\n", qTable, colDef))
		}
	}

	// 7. Deleted Models
	for _, tableName := range changes.DeletedModels {
		upSB.WriteString(fmt.Sprintf("DROP TABLE IF EXISTS %s;\n", d.QuoteIdentifier(tableName)))
	}

	return strings.TrimSpace(upSB.String()), strings.TrimSpace(downSB.String())
}
