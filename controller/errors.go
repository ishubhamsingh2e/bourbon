// Package controller provides base controller abstractions and RFC 7807 problem details error structures.
package controller

import (
	"fmt"
	"net/http"
)

// InvalidParam represents a single field-level validation failure in an RFC 7807 payload.
type InvalidParam struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Message string `json:"message"`
}

// ProblemDetail models an RFC 7807 application/problem+json response payload.
type ProblemDetail struct {
	Type          string         `json:"type"`
	Title         string         `json:"title"`
	Status        int            `json:"status"`
	Detail        string         `json:"detail,omitempty"`
	Instance      string         `json:"instance,omitempty"`
	InvalidParams []InvalidParam `json:"invalid_params,omitempty"`
}

// Error implements the standard error interface for ProblemDetail.
func (p *ProblemDetail) Error() string {
	if p.Detail != "" {
		return fmt.Sprintf("%s (%d): %s", p.Title, p.Status, p.Detail)
	}
	return fmt.Sprintf("%s (%d)", p.Title, p.Status)
}

// NewProblemDetail constructs a standard RFC 7807 ProblemDetail object.
func NewProblemDetail(status int, title, detail string) *ProblemDetail {
	if title == "" {
		title = http.StatusText(status)
	}
	return &ProblemDetail{
		Type:   "about:blank",
		Title:  title,
		Status: status,
		Detail: detail,
	}
}

// NewProblemDetails constructs an RFC 7807 ProblemDetail object with custom type and instance.
func NewProblemDetails(errType, title string, status int, detail, instance string) *ProblemDetail {
	if title == "" {
		title = http.StatusText(status)
	}
	if errType == "" {
		errType = "about:blank"
	}
	return &ProblemDetail{
		Type:     errType,
		Title:    title,
		Status:   status,
		Detail:   detail,
		Instance: instance,
	}
}

// NewValidationError constructs an unprocessable entity (422) problem detail with field errors.
func NewValidationError(instance string, params []InvalidParam) *ProblemDetail {
	return &ProblemDetail{
		Type:          "https://bourbon.dev/errors/validation-failed",
		Title:         "Validation Failed",
		Status:        http.StatusUnprocessableEntity,
		Detail:        "One or more request parameters failed validation.",
		Instance:      instance,
		InvalidParams: params,
	}
}
