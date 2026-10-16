package validation

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type requiredRule struct{}

func (r *requiredRule) Validate(ctx context.Context, field string, value any) error {
	if value == nil {
		return FieldError{Field: field, Tag: "required", Message: "field is required"}
	}
	switch v := value.(type) {
	case string:
		if strings.TrimSpace(v) == "" {
			return FieldError{Field: field, Tag: "required", Message: "field cannot be empty"}
		}
	case []any:
		if len(v) == 0 {
			return FieldError{Field: field, Tag: "required", Message: "field cannot be empty"}
		}
	}
	return nil
}

// Required ensures the field is present and non-empty.
func Required() Rule {
	return &requiredRule{}
}

type emailRule struct{}

func (r *emailRule) Validate(ctx context.Context, field string, value any) error {
	if value == nil {
		return nil
	}
	str, ok := value.(string)
	if !ok || str == "" {
		return nil
	}
	if !emailRegex.MatchString(str) {
		return FieldError{Field: field, Tag: "email", Message: "must be a valid email address"}
	}
	return nil
}

// Email ensures the string field matches standard email syntax.
func Email() Rule {
	return &emailRule{}
}

type minLengthRule struct {
	min int
}

func (r *minLengthRule) Validate(ctx context.Context, field string, value any) error {
	if value == nil {
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return nil
	}
	if len([]rune(str)) < r.min {
		return FieldError{
			Field:   field,
			Tag:     "min_length",
			Message: fmt.Sprintf("must have at least %d characters", r.min),
		}
	}
	return nil
}

// MinLength enforces minimum string length.
func MinLength(min int) Rule {
	return &minLengthRule{min: min}
}

type maxLengthRule struct {
	max int
}

func (r *maxLengthRule) Validate(ctx context.Context, field string, value any) error {
	if value == nil {
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return nil
	}
	if len([]rune(str)) > r.max {
		return FieldError{
			Field:   field,
			Tag:     "max_length",
			Message: fmt.Sprintf("must not exceed %d characters", r.max),
		}
	}
	return nil
}

// MaxLength enforces maximum string length.
func MaxLength(max int) Rule {
	return &maxLengthRule{max: max}
}

type lengthRule struct {
	exact int
}

func (r *lengthRule) Validate(ctx context.Context, field string, value any) error {
	if value == nil {
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return nil
	}
	if len([]rune(str)) != r.exact {
		return FieldError{
			Field:   field,
			Tag:     "length",
			Message: fmt.Sprintf("must have exactly %d characters", r.exact),
		}
	}
	return nil
}

// Length enforces an exact string length.
func Length(exact int) Rule {
	return &lengthRule{exact: exact}
}

type minRule struct {
	min float64
}

func (r *minRule) Validate(ctx context.Context, field string, value any) error {
	if value == nil {
		return nil
	}
	num, ok := toFloat(value)
	if !ok {
		return nil
	}
	if num < r.min {
		return FieldError{
			Field:   field,
			Tag:     "min",
			Message: fmt.Sprintf("must be at least %v", r.min),
		}
	}
	return nil
}

// Min enforces a minimum numeric value.
func Min(min float64) Rule {
	return &minRule{min: min}
}

type maxRule struct {
	max float64
}

func (r *maxRule) Validate(ctx context.Context, field string, value any) error {
	if value == nil {
		return nil
	}
	num, ok := toFloat(value)
	if !ok {
		return nil
	}
	if num > r.max {
		return FieldError{
			Field:   field,
			Tag:     "max",
			Message: fmt.Sprintf("must not exceed %v", r.max),
		}
	}
	return nil
}

// Max enforces a maximum numeric value.
func Max(max float64) Rule {
	return &maxRule{max: max}
}

type inRule struct {
	allowed map[string]struct{}
}

func (r *inRule) Validate(ctx context.Context, field string, value any) error {
	if value == nil {
		return nil
	}
	str := fmt.Sprintf("%v", value)
	if _, ok := r.allowed[str]; !ok {
		return FieldError{
			Field:   field,
			Tag:     "in",
			Message: "is not an allowed value",
		}
	}
	return nil
}

// In enforces that the value belongs to a predefined list of allowed values.
func In(allowed ...string) Rule {
	set := make(map[string]struct{}, len(allowed))
	for _, a := range allowed {
		set[a] = struct{}{}
	}
	return &inRule{allowed: set}
}

type regexRule struct {
	pattern *regexp.Regexp
}

func (r *regexRule) Validate(ctx context.Context, field string, value any) error {
	if value == nil {
		return nil
	}
	str, ok := value.(string)
	if !ok || str == "" {
		return nil
	}
	if !r.pattern.MatchString(str) {
		return FieldError{
			Field:   field,
			Tag:     "regex",
			Message: "does not match required format",
		}
	}
	return nil
}

// Regex enforces string matching against a regular expression.
func Regex(pattern string) Rule {
	return &regexRule{pattern: regexp.MustCompile(pattern)}
}

func toFloat(val any) (float64, bool) {
	switch v := val.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case string:
		f, err := strconv.ParseFloat(v, 64)
		return f, err == nil
	}
	return 0, false
}
