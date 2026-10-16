package schema_test

import (
	"strings"
	"testing"
	"time"

	"github.com/ishubhamsingh2e/bourbon/database/dialect"
	"github.com/ishubhamsingh2e/bourbon/database/schema"
	"github.com/ishubhamsingh2e/bourbon/model"
)

type Author struct {
	model.BaseModel
	Name  string `db:"name" bourbon:",notnull"`
	Email string `db:"email" bourbon:",unique,notnull,index"`
}

func (Author) TableName() string {
	return "authors"
}

type Article struct {
	model.BaseModel
	AuthorID    string     `db:"author_id,references:authors(id),on_delete:cascade" bourbon:",notnull,index"`
	Title       string     `db:"title" bourbon:",notnull"`
	Body        string     `db:"body" bourbon:"type:text"`
	PublishedAt *time.Time `db:"published_at"`
	Author      *Author    `bourbon:"rel:belongs-to"`
}

func (Article) TableName() string {
	return "articles"
}

func TestGenerateCreateTableSQL(t *testing.T) {
	sqliteDialect, _ := dialect.Get("sqlite")
	pgDialect, _ := dialect.Get("postgres")
	mysqlDialect, _ := dialect.Get("mysql")

	table, err := schema.InspectModel(sqliteDialect, Article{})
	if err != nil {
		t.Fatalf("InspectModel failed: %v", err)
	}

	// 1. Test SQLite DDL
	sqliteSQL := schema.GenerateCreateTableSQL(sqliteDialect, table)
	if !strings.Contains(sqliteSQL, `CREATE TABLE IF NOT EXISTS "articles"`) {
		t.Errorf("expected articles table in SQLite SQL, got:\n%s", sqliteSQL)
	}
	if !strings.Contains(sqliteSQL, `REFERENCES "authors"("id") ON DELETE CASCADE`) {
		t.Errorf("expected foreign key in SQLite SQL, got:\n%s", sqliteSQL)
	}
	if !strings.Contains(sqliteSQL, `CREATE INDEX IF NOT EXISTS "idx_articles_author_id"`) {
		t.Errorf("expected index for author_id, got:\n%s", sqliteSQL)
	}

	// 2. Test PostgreSQL DDL
	pgTable, _ := schema.InspectModel(pgDialect, Article{})
	pgSQL := schema.GenerateCreateTableSQL(pgDialect, pgTable)
	if !strings.Contains(pgSQL, `CREATE TABLE IF NOT EXISTS "articles"`) {
		t.Errorf("expected articles table in Postgres SQL, got:\n%s", pgSQL)
	}

	// 3. Test MySQL DDL
	mysqlTable, _ := schema.InspectModel(mysqlDialect, Article{})
	mysqlSQL := schema.GenerateCreateTableSQL(mysqlDialect, mysqlTable)
	if !strings.Contains(mysqlSQL, "CREATE TABLE IF NOT EXISTS `articles`") {
		t.Errorf("expected backticks in MySQL SQL, got:\n%s", mysqlSQL)
	}
}

func TestGenerateDropTableSQL(t *testing.T) {
	sqliteDialect, _ := dialect.Get("sqlite")
	table, _ := schema.InspectModel(sqliteDialect, Article{})

	dropSQL := schema.GenerateDropTableSQL(sqliteDialect, table)
	if !strings.Contains(dropSQL, `DROP TABLE IF EXISTS "articles";`) {
		t.Errorf("expected DROP TABLE in SQL, got: %s", dropSQL)
	}
	if !strings.Contains(dropSQL, `DROP INDEX IF EXISTS "idx_articles_author_id";`) {
		t.Errorf("expected DROP INDEX in SQL, got: %s", dropSQL)
	}
}
