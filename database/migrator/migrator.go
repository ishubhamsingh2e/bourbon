// Package migrator provides embedded database migration management, schema versioning, and migration file scaffolding for Bourbon.
package migrator

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ishubhamsingh2e/bourbon/database/dialect"
	"github.com/pressly/goose/v3"
)

func init() {
	// Silence third-party logger output to maintain clean Bourbon formatting
	goose.SetLogger(goose.NopLogger())
}

// MakeMigrationsResult contains metadata about generated migrations.
type MakeMigrationsResult struct {
	FilePath     string
	AppName      string
	MigrationID  string
	Descriptions []string
}

// MigrationStatus describes the execution state of an individual migration.
type MigrationStatus struct {
	Version   int64
	Name      string
	Applied   bool
	AppliedAt *time.Time
}

// SQLDBProvider is implemented by database adapters that provide access to *sql.DB and driver identification.
type SQLDBProvider interface {
	SQLDB() *sql.DB
	DriverName() string
}

// Migrator manages migration execution against a database connection.
type Migrator struct {
	db      *sql.DB
	dir     string
	dialect string
}

// ResolveMigrator resolves a *Migrator from any supported database instance (*sql.DB or SQLDBProvider).
func ResolveMigrator(db any, dir string) (*Migrator, error) {
	if db == nil {
		return nil, fmt.Errorf("migrator: db cannot be nil")
	}

	var sqlDB *sql.DB
	dialectName := "sqlite3"

	if provider, ok := db.(SQLDBProvider); ok {
		sqlDB = provider.SQLDB()
		dialectName = provider.DriverName()
	} else if rawSQL, ok := db.(*sql.DB); ok {
		sqlDB = rawSQL
	} else if getter, ok := db.(interface{ DB() *sql.DB }); ok {
		sqlDB = getter.DB()
	}

	if sqlDB == nil {
		return nil, fmt.Errorf("migrator: unsupported database type %T, must implement SQLDBProvider or *sql.DB", db)
	}

	return New(sqlDB, dir, dialectName)
}

// Migrate applies all pending database migrations in dir using the provided database connection.
func Migrate(db any, dir string) error {
	m, err := ResolveMigrator(db, dir)
	if err != nil {
		return err
	}
	return m.Migrate()
}

// Rollback rolls back the last database migration in dir using the provided database connection.
func Rollback(db any, dir string) error {
	m, err := ResolveMigrator(db, dir)
	if err != nil {
		return err
	}
	return m.Rollback()
}

// Status returns the migration status in dir using the provided database connection.
func Status(db any, dir string) ([]MigrationStatus, error) {
	m, err := ResolveMigrator(db, dir)
	if err != nil {
		return nil, err
	}
	return m.Status()
}

// Version returns the current applied schema version in dir using the provided database connection.
func Version(db any, dir string) (int64, error) {
	m, err := ResolveMigrator(db, dir)
	if err != nil {
		return 0, err
	}
	return m.Version()
}

// New constructs a new Migrator instance.
func New(db *sql.DB, dir string, dialectName string) (*Migrator, error) {
	if db == nil {
		return nil, fmt.Errorf("migrator: db connection cannot be nil")
	}
	if dir == "" {
		dir = filepath.Join("db", "migrations")
	}

	normDialect := strings.ToLower(strings.TrimSpace(dialectName))
	switch normDialect {
	case "sqlite", "sqlite3":
		normDialect = "sqlite3"
	case "postgres", "postgresql", "pgx":
		normDialect = "postgres"
	case "mysql", "mariadb":
		normDialect = "mysql"
	default:
		normDialect = "sqlite3"
	}

	return &Migrator{
		db:      db,
		dir:     dir,
		dialect: normDialect,
	}, nil
}

// Migrate applies all pending migrations.
func (m *Migrator) Migrate() error {
	if err := goose.SetDialect(m.dialect); err != nil {
		return fmt.Errorf("migrator: failed setting dialect %s: %w", m.dialect, err)
	}
	return goose.Up(m.db, m.dir)
}

// Rollback rolls back the last applied migration.
func (m *Migrator) Rollback() error {
	if err := goose.SetDialect(m.dialect); err != nil {
		return fmt.Errorf("migrator: failed setting dialect %s: %w", m.dialect, err)
	}
	return goose.Down(m.db, m.dir)
}

// Reset rolls back all applied migrations.
func (m *Migrator) Reset() error {
	if err := goose.SetDialect(m.dialect); err != nil {
		return fmt.Errorf("migrator: failed setting dialect %s: %w", m.dialect, err)
	}
	return goose.Reset(m.db, m.dir)
}

// Version returns the current database schema version number.
func (m *Migrator) Version() (int64, error) {
	if err := goose.SetDialect(m.dialect); err != nil {
		return 0, fmt.Errorf("migrator: failed setting dialect: %w", err)
	}
	return goose.GetDBVersion(m.db)
}

// Status returns a list of all migrations and their applied status.
func (m *Migrator) Status() ([]MigrationStatus, error) {
	if err := goose.SetDialect(m.dialect); err != nil {
		return nil, fmt.Errorf("migrator: failed setting dialect: %w", err)
	}

	migrations, err := goose.CollectMigrations(m.dir, 0, goose.MaxVersion)
	if err != nil {
		return nil, fmt.Errorf("migrator: failed collecting migrations from %s: %w", m.dir, err)
	}

	// Read version records
	appliedMap := make(map[int64]time.Time)
	rows, err := m.db.Query(fmt.Sprintf("SELECT version_id, is_applied, tstamp FROM %s", goose.TableName()))
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var v int64
			var applied bool
			var tstamp time.Time
			if err := rows.Scan(&v, &applied, &tstamp); err == nil && applied {
				appliedMap[v] = tstamp
			}
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
	}

	statuses := make([]MigrationStatus, 0, len(migrations))
	for _, mg := range migrations {
		st := MigrationStatus{
			Version: mg.Version,
			Name:    filepath.Base(mg.Source),
		}
		if tstamp, ok := appliedMap[mg.Version]; ok {
			st.Applied = true
			st.AppliedAt = &tstamp
		}
		statuses = append(statuses, st)
	}

	return statuses, nil
}

// Create generates a new timestamped migration file with Up and Down sections.
func Create(dir string, name string, upSQL, downSQL string) (string, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("migrator: failed creating migrations dir: %w", err)
	}

	if strings.TrimSpace(name) == "" {
		name = "migration"
	}

	now := time.Now().UTC()
	timestampStr := now.Format("20060102150405")
	var currentTs int64
	_, _ = fmt.Sscanf(timestampStr, "%d", &currentTs)

	files, _ := os.ReadDir(dir)
	var maxTs int64 = 0
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".sql") {
			parts := strings.SplitN(f.Name(), "_", 2)
			var fileTs int64
			if _, err := fmt.Sscanf(parts[0], "%d", &fileTs); err == nil {
				if fileTs > maxTs {
					maxTs = fileTs
				}
			}
		}
	}

	if maxTs >= currentTs {
		tParsed, err := time.Parse("20060102150405", fmt.Sprintf("%d", maxTs))
		if err == nil {
			timestampStr = tParsed.Add(time.Second).Format("20060102150405")
		} else {
			timestampStr = fmt.Sprintf("%d", maxTs+1)
		}
	}

	snakeName := toSnakeCase(name)
	filename := fmt.Sprintf("%s_%s.sql", timestampStr, snakeName)
	filePath := filepath.Join(dir, filename)

	if upSQL == "" {
		tblName := deriveTableName(snakeName)
		upSQL = fmt.Sprintf("-- +bourbon StatementBegin\nCREATE TABLE IF NOT EXISTS %s (\n    id VARCHAR(36) PRIMARY KEY,\n    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,\n    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP\n);\n-- +bourbon StatementEnd", tblName)
	}
	if downSQL == "" {
		tblName := deriveTableName(snakeName)
		downSQL = fmt.Sprintf("-- +bourbon StatementBegin\nDROP TABLE IF EXISTS %s;\n-- +bourbon StatementEnd", tblName)
	}

	// Normalize Bourbon markers to goose-compatible markers
	content := fmt.Sprintf(`-- +goose Up
%s

-- +goose Down
%s
`, normalizeToEngineMarkers(upSQL), normalizeToEngineMarkers(downSQL))

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("migrator: failed writing migration file: %w", err)
	}

	return filePath, nil
}

func deriveTableName(snakeName string) string {
	name := snakeName
	if strings.HasPrefix(name, "create_") {
		name = strings.TrimPrefix(name, "create_")
	} else if strings.HasPrefix(name, "add_") {
		name = strings.TrimPrefix(name, "add_")
	}
	if strings.HasSuffix(name, "_table") {
		name = strings.TrimSuffix(name, "_table")
	} else if strings.HasSuffix(name, "_tbl") {
		name = strings.TrimSuffix(name, "_tbl")
	}

	if name == "" || name == "migration" {
		return "items"
	}

	if strings.HasSuffix(name, "y") && !strings.HasSuffix(name, "ay") && !strings.HasSuffix(name, "ey") && !strings.HasSuffix(name, "oy") && !strings.HasSuffix(name, "uy") {
		return name[:len(name)-1] + "ies"
	}
	if strings.HasSuffix(name, "s") {
		return name
	}
	if strings.HasSuffix(name, "x") || strings.HasSuffix(name, "ch") || strings.HasSuffix(name, "sh") {
		return name + "es"
	}
	return name + "s"
}

func normalizeToEngineMarkers(s string) string {
	s = strings.ReplaceAll(s, "-- +bourbon Up", "-- +goose Up")
	s = strings.ReplaceAll(s, "-- +bourbon Down", "-- +goose Down")
	s = strings.ReplaceAll(s, "-- +bourbon StatementBegin", "-- +goose StatementBegin")
	s = strings.ReplaceAll(s, "-- +bourbon StatementEnd", "-- +goose StatementEnd")
	return s
}

func toSnakeCase(s string) string {
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
		} else if r == ' ' || r == '-' {
			res = append(res, '_')
		} else {
			res = append(res, r)
		}
	}
	return string(res)
}

// MakeMigrations scans the project models, diffs against previous state, prompts for renames, and generates migration files.
func MakeMigrations(rootDir string, migrationsDir string, targetApp string, name string, empty bool, prompter Prompter, dialectName string) (*MakeMigrationsResult, error) {
	if rootDir == "" {
		rootDir = "."
	}
	if migrationsDir == "" {
		migrationsDir = filepath.Join("db", "migrations")
	}
	if !filepath.IsAbs(migrationsDir) {
		migrationsDir = filepath.Join(rootDir, migrationsDir)
	}

	d, err := dialect.Get(dialectName)
	if err != nil {
		d, _ = dialect.Get("sqlite")
	}

	if empty {
		if name == "" {
			name = "custom_migration"
		}
		path, err := Create(migrationsDir, name, "-- Empty migration\n", "-- Empty rollback\n")
		if err != nil {
			return nil, err
		}
		return &MakeMigrationsResult{
			FilePath:     path,
			AppName:      targetApp,
			MigrationID:  filepath.Base(path),
			Descriptions: []string{"Empty migration skeleton created"},
		}, nil
	}

	models, err := ScanProjectModels(rootDir, targetApp, d)
	if err != nil {
		return nil, fmt.Errorf("failed scanning project models: %w", err)
	}

	state, err := LoadMigrationState(rootDir)
	if err != nil {
		return nil, fmt.Errorf("failed loading migration state: %w", err)
	}

	// Group models by app
	appModels := make(map[string][]ModelAST)
	for _, m := range models {
		appModels[m.App] = append(appModels[m.App], m)
	}

	// Check apps in previous state that might have had all models deleted
	for appName := range state.Apps {
		if _, exists := appModels[appName]; !exists && (targetApp == "" || targetApp == appName) {
			appModels[appName] = []ModelAST{}
		}
	}

	if len(appModels) == 0 {
		return nil, nil
	}

	var finalResult *MakeMigrationsResult

	for app, mList := range appModels {
		if targetApp != "" && app != targetApp {
			continue
		}

		prevAppState := state.Apps[app]
		changes, err := DetectSchemaChanges(app, mList, prevAppState, prompter, d)
		if err != nil {
			return nil, fmt.Errorf("failed detecting schema changes for app %s: %w", app, err)
		}

		if !changes.HasChanges() {
			continue
		}

		migName := name
		if migName == "" {
			if len(changes.CreatedModels) > 0 && len(changes.Descriptions) == 1 {
				migName = "create_" + toSnakeCase(changes.CreatedModels[0].Name) + "_table"
			} else if len(changes.RenamedModels) > 0 {
				migName = "rename_models"
			} else if len(changes.RenamedColumns) > 0 {
				migName = "rename_fields"
			} else if len(changes.AddedColumns) > 0 {
				migName = "add_fields"
			} else {
				migName = "migration"
			}
		}

		upSQL, downSQL := GenerateSQLFromChanges(changes, d)
		path, err := Create(migrationsDir, migName, upSQL, downSQL)
		if err != nil {
			return nil, fmt.Errorf("failed creating migration file: %w", err)
		}

		// Update state
		modelMap := make(map[string]*ModelAST)
		for i := range mList {
			modelMap[mList[i].Name] = &mList[i]
		}

		state.Apps[app] = &AppState{
			LastHash:      ComputeModelsHash(mList),
			LastMigration: filepath.Base(path),
			Models:        modelMap,
		}

		finalResult = &MakeMigrationsResult{
			FilePath:     path,
			AppName:      app,
			MigrationID:  filepath.Base(path),
			Descriptions: changes.Descriptions,
		}
	}

	if finalResult != nil {
		if err := SaveMigrationState(rootDir, state); err != nil {
			return nil, fmt.Errorf("failed saving migration state: %w", err)
		}
	}

	return finalResult, nil
}

// SQLMigrate extracts and returns the SQL statements for a given migration file.
func SQLMigrate(migrationsDir string, migrationQuery string, backwards bool) (string, error) {
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return "", fmt.Errorf("failed reading migrations directory %s: %w", migrationsDir, err)
	}

	var targetFile string
	normQuery := strings.ToLower(strings.TrimSpace(migrationQuery))
	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".sql") {
			continue
		}
		fName := strings.ToLower(f.Name())
		if strings.Contains(fName, normQuery) || strings.Contains(fName, strings.TrimSuffix(normQuery, "s")) {
			targetFile = filepath.Join(migrationsDir, f.Name())
			break
		}
	}

	if targetFile == "" {
		return "", fmt.Errorf("migration matching %q not found in %s", migrationQuery, migrationsDir)
	}

	content, err := os.ReadFile(targetFile)
	if err != nil {
		return "", fmt.Errorf("failed reading migration file %s: %w", targetFile, err)
	}

	raw := string(content)
	upIdx := strings.Index(raw, "-- +goose Up")
	downIdx := strings.Index(raw, "-- +goose Down")

	if backwards {
		if downIdx == -1 {
			return "-- No rollback SQL found in migration", nil
		}
		return strings.TrimSpace(raw[downIdx+len("-- +goose Down"):]), nil
	}

	if upIdx != -1 {
		if downIdx != -1 && downIdx > upIdx {
			return strings.TrimSpace(raw[upIdx+len("-- +goose Up") : downIdx]), nil
		}
		return strings.TrimSpace(raw[upIdx+len("-- +goose Up"):]), nil
	}

	return strings.TrimSpace(raw), nil
}
