# Validation & Error Handling in Bourbon

Bourbon includes a high-performance, zero-reflection validation engine capable of declarative schema validation, tag-based struct inspection, custom domain rules, and automatic RFC 7807 Problem Details serialization.

---

## 1. Declarative Schema Validation

Use `validation.New()` to construct a schema validator with explicit rule chaining:

```go
package users

import (
	"context"

	"github.com/ishubhamsingh2e/bourbon/validation"
)

var registerSchema = validation.New().
	Field("email", validation.Required(), validation.Email()).
	Field("password", validation.Required(), validation.MinLength(8)).
	Field("role", validation.Required(), validation.In("admin", "editor", "viewer"))

func validateRegistration(ctx context.Context, payload map[string]any) error {
	return registerSchema.ValidateMap(ctx, payload)
}
```

---

## 2. Built-in Validation Rules

Bourbon provides standard library-backed rules with no external dependencies:

| Rule | Description | Example |
| :--- | :--- | :--- |
| `validation.Required()` | Rejects `nil`, empty strings, and empty slices. | `validation.Required()` |
| `validation.Email()` | Verifies standard RFC email formatting via regex. | `validation.Email()` |
| `validation.MinLength(n)` | Validates string minimum character count. | `validation.MinLength(8)` |
| `validation.MaxLength(n)` | Validates string maximum character count. | `validation.MaxLength(64)` |
| `validation.Min(n)` | Checks minimum integer or float numeric bounds. | `validation.Min(18)` |
| `validation.Max(n)` | Checks maximum integer or float numeric bounds. | `validation.Max(120)` |
| `validation.In(vals...)` | Ensures value belongs to an allowed set of options. | `validation.In("active", "pending")` |
| `validation.Pattern(regex)`| Matches a precompiled `*regexp.Regexp` pattern. | `validation.Pattern(slugRegex)` |
| `validation.Custom(fn)` | Custom validation function for business logic. | `validation.Custom(checkUniqueEmail)` |

---

## 3. Struct Tag Validation

For request DTOs, Bourbon supports tag-based validation using the `validate` struct tag:

```go
package dtos

import (
	"github.com/ishubhamsingh2e/bourbon/core"
	"github.com/ishubhamsingh2e/bourbon/validation"
)

type CreateArticleRequest struct {
	Title   string `json:"title" validate:"required,min=5,max=120"`
	Content string `json:"content" validate:"required,min=20"`
	Status  string `json:"status" validate:"required,in=draft,published"`
}

func HandleCreateArticle(c *core.Context) error {
	var req CreateArticleRequest
	if err := c.BindJSON(&req); err != nil {
		return err
	}

	validator := validation.NewTagValidator()
	if err := validator.Validate(req); err != nil {
		// Automatically returns 422 Unprocessable Entity with RFC 7807 payload
		return c.ValidationError(err)
	}

	return c.JSON(201, map[string]string{"message": "article created"})
}
```

---

## 4. RFC 7807 Problem Details Output

When validation fails, Bourbon serializes errors conforming to the RFC 7807 Problem Details specification:

```json
{
  "type": "https://bourbon.dev/errors/validation-failed",
  "title": "Validation Failed",
  "status": 422,
  "detail": "One or more fields failed validation checks",
  "instance": "/api/v1/articles",
  "invalid_params": [
    {
      "field": "title",
      "tag": "min",
      "message": "value length must be at least 5"
    },
    {
      "field": "status",
      "tag": "in",
      "message": "value must be one of [draft published]"
    }
  ]
}
```

---

## 5. Custom Domain Validation Rules

You can implement custom rules by satisfying the `validation.Rule` interface:

```go
package rules

import (
	"context"
	"strings"

	"github.com/ishubhamsingh2e/bourbon/validation"
)

type NoProfanityRule struct{}

func (r *NoProfanityRule) Validate(ctx context.Context, field string, value any) error {
	str, ok := value.(string)
	if !ok {
		return nil
	}
	if strings.Contains(strings.ToLower(str), "forbidden_word") {
		return validation.FieldError{
			Field:   field,
			Tag:     "profanity",
			Message: "content contains prohibited vocabulary",
		}
	}
	return nil
}
```
