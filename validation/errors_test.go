package validation

import (
	"strings"
	"testing"
)

func TestValidationErrorsFormatting(t *testing.T) {
	errs := Errors{
		FieldError{Field: "username", Tag: "required", Message: "field is required"},
		FieldError{Field: "email", Tag: "email", Message: "must be valid email"},
	}

	errStr := errs.Error()
	if !strings.Contains(errStr, "username") || !strings.Contains(errStr, "email") {
		t.Errorf("Error() output did not contain field names: %s", errStr)
	}

	if !errs.Has("username") {
		t.Errorf("Has(username) should be true")
	}
	if errs.Has("age") {
		t.Errorf("Has(age) should be false")
	}

	msgs := errs.Get("username")
	if len(msgs) != 1 || msgs[0] != "field is required" {
		t.Errorf("Get(username) mismatch: %v", msgs)
	}
}
