package dialect_test

import (
	"testing"

	"github.com/ishubhamsingh2e/bourbon/database/dialect"
)

func TestDialectsRegistration(t *testing.T) {
	cases := []struct {
		name       string
		expectName string
		quote      string
	}{
		{"sqlite", "sqlite", `"users"`},
		{"sqlite3", "sqlite", `"users"`},
		{"postgres", "postgres", `"users"`},
		{"postgresql", "postgres", `"users"`},
		{"mysql", "mysql", "`users`"},
		{"mariadb", "mysql", "`users`"},
	}

	for _, tc := range cases {
		d, err := dialect.Get(tc.name)
		if err != nil {
			t.Fatalf("expected dialect for %s, got err: %v", tc.name, err)
		}
		if d.Name() != tc.expectName {
			t.Errorf("expected dialect name %s, got %s", tc.expectName, d.Name())
		}
		if quoted := d.QuoteIdentifier("users"); quoted != tc.quote {
			t.Errorf("expected %s, got %s", tc.quote, quoted)
		}
	}
}

func TestPostgresRebind(t *testing.T) {
	d, err := dialect.Get("postgres")
	if err != nil {
		t.Fatal(err)
	}

	query := "SELECT * FROM users WHERE email = ? AND status = ? AND note = 'is this ?' AND age > ?"
	expected := "SELECT * FROM users WHERE email = $1 AND status = $2 AND note = 'is this ?' AND age > $3"

	result := d.Rebind(query)
	if result != expected {
		t.Errorf("\nExpected: %s\nGot:      %s", expected, result)
	}
}
