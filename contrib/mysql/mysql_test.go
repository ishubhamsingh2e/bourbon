package mysql_test

import (
	"testing"

	"github.com/ishubhamsingh2e/bourbon/contrib/mysql"
)

func TestOpenEmptyDSN(t *testing.T) {
	_, err := mysql.Open(mysql.Config{DSN: ""})
	if err == nil {
		t.Fatal("expected error when opening mysql with empty DSN")
	}
}
