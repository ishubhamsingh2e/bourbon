package validation

import (
	"context"
	"encoding/json"
	"reflect"
)

// Rule defines a single field validation constraint.
type Rule interface {
	Validate(ctx context.Context, field string, value any) error
}

// Validator validates input fields and domain models.
type Validator interface {
	Validate(v any) error
}

// SchemaValidator executes declarative validation rules on structured inputs.
type SchemaValidator struct {
	rules map[string][]Rule
}

// New constructs an empty SchemaValidator.
func New() *SchemaValidator {
	return &SchemaValidator{
		rules: make(map[string][]Rule),
	}
}

// Field registers validation rules for a specific field name.
func (v *SchemaValidator) Field(name string, rules ...Rule) *SchemaValidator {
	if _, exists := v.rules[name]; !exists {
		v.rules[name] = make([]Rule, 0, len(rules))
	}
	v.rules[name] = append(v.rules[name], rules...)
	return v
}

// ValidateMap applies all schema rules against a key-value data map.
func (v *SchemaValidator) ValidateMap(ctx context.Context, data map[string]any) error {
	var errs Errors

	for field, rules := range v.rules {
		val := data[field]
		for _, r := range rules {
			if err := r.Validate(ctx, field, val); err != nil {
				if fe, ok := err.(FieldError); ok {
					errs = append(errs, fe)
				} else {
					errs = append(errs, FieldError{
						Field:   field,
						Tag:     "custom",
						Message: err.Error(),
					})
				}
			}
		}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

// Validate converts structs or maps and executes the schema rules.
func (v *SchemaValidator) Validate(target any) error {
	if target == nil {
		return v.ValidateMap(context.Background(), nil)
	}

	// If already a map
	if m, ok := target.(map[string]any); ok {
		return v.ValidateMap(context.Background(), m)
	}

	// Convert struct or pointer to map via JSON or reflection
	data := structToMap(target)
	return v.ValidateMap(context.Background(), data)
}

func structToMap(target any) map[string]any {
	res := make(map[string]any)
	raw, err := json.Marshal(target)
	if err == nil {
		_ = json.Unmarshal(raw, &res)
		return res
	}

	val := reflect.ValueOf(target)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return res
	}

	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.IsExported() {
			name := field.Tag.Get("json")
			if name == "" || name == "-" {
				name = field.Name
			}
			res[name] = val.Field(i).Interface()
		}
	}
	return res
}
