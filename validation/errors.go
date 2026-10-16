// Package validation provides declarative input validation and RFC 7807 problem details error integration.
package validation

import (
	"fmt"
	"strings"
)

// FieldError represents a single validation constraint failure on a field.
type FieldError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Message string `json:"message"`
}

func (e FieldError) Error() string {
	return fmt.Sprintf("validation failed on '%s': %s", e.Field, e.Message)
}

// Errors is a collection of FieldErrors that implements error and RFC 7807 formatting.
type Errors []FieldError

func (e Errors) Error() string {
	if len(e) == 0 {
		return ""
	}
	var sb strings.Builder
	for i, err := range e {
		if i > 0 {
			sb.WriteString("; ")
		}
		sb.WriteString(err.Error())
	}
	return sb.String()
}

// Has checks if any validation error was registered for the field.
func (e Errors) Has(field string) bool {
	for _, err := range e {
		if err.Field == field {
			return true
		}
	}
	return false
}

// Get returns all error messages associated with the specified field.
func (e Errors) Get(field string) []string {
	var msgs []string
	for _, err := range e {
		if err.Field == field {
			msgs = append(msgs, err.Message)
		}
	}
	return msgs
}
