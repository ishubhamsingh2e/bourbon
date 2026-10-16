package model

import (
	"testing"
)

type TestUser struct {
	BaseModel
	Email string `db:"user_email"`
	Age   *int   `db:"age"`
	Notes string `db:"-"`
}

func (t TestUser) TableName() string {
	return "auth_users"
}

type AutoTable struct {
	ID   int
	Name string
}

func TestModelInspectCustomTable(t *testing.T) {
	meta := Inspect(TestUser{})
	if meta == nil {
		t.Fatalf("Inspect returned nil")
	}

	if meta.TableName != "auth_users" {
		t.Errorf("meta.TableName = %v, want auth_users", meta.TableName)
	}

	if meta.PrimaryCol != "id" {
		t.Errorf("meta.PrimaryCol = %v, want id", meta.PrimaryCol)
	}

	// Verify fields
	fieldMap := make(map[string]FieldMeta)
	for _, f := range meta.Fields {
		fieldMap[f.Name] = f
	}

	if _, exists := fieldMap["Notes"]; exists {
		t.Errorf("ignored field 'Notes' with db:'-' was included")
	}

	emailField, exists := fieldMap["Email"]
	if !exists || emailField.Column != "user_email" {
		t.Errorf("Email field metadata incorrect: %+v", emailField)
	}

	ageField, exists := fieldMap["Age"]
	if !exists || !ageField.Nullable {
		t.Errorf("Age pointer field should be marked nullable")
	}
}

func TestModelInspectAutoPluralTable(t *testing.T) {
	meta := Inspect(&AutoTable{})
	if meta == nil {
		t.Fatalf("Inspect returned nil")
	}

	if meta.TableName != "auto_tables" {
		t.Errorf("meta.TableName = %v, want auto_tables", meta.TableName)
	}
	if meta.PrimaryCol != "id" {
		t.Errorf("meta.PrimaryCol = %v, want id", meta.PrimaryCol)
	}
}
