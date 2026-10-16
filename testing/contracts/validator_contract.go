package contracts

import (
	"testing"

	"github.com/ishubhamsingh2e/bourbon/validation"
)

// VerifyValidatorContract executes a standard compliance test suite against any validation.Validator implementation.
func VerifyValidatorContract(t *testing.T, validator validation.Validator, validTarget, invalidTarget any) {
	t.Run("Valid_Target_Success", func(t *testing.T) {
		if err := validator.Validate(validTarget); err != nil {
			t.Fatalf("expected validTarget to pass validation, got: %v", err)
		}
	})

	t.Run("Invalid_Target_Failure", func(t *testing.T) {
		err := validator.Validate(invalidTarget)
		if err == nil {
			t.Fatalf("expected invalidTarget to fail validation, got nil")
		}
	})
}
