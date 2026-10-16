package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/ishubhamsingh2e/bourbon/router"
)

var (
	// ErrUnsupportedBindingTarget is returned when Bind is given a non-pointer target.
	ErrUnsupportedBindingTarget = errors.New("core: binding target must be a non-nil pointer to struct")
)

// Validator defines an interface for self-validating request models.
type Validator interface {
	Validate() error
}

// Param returns the URL path parameter matching key.
func (c *Context) Param(key string) string {
	c.mu.RLock()
	if val, ok := c.params[key]; ok {
		c.mu.RUnlock()
		return val
	}
	c.mu.RUnlock()

	if c.Request != nil {
		if val := router.URLParam(c.Request, key); val != "" {
			return val
		}
		return c.Request.PathValue(key)
	}
	return ""
}

// SetParam manually sets a path parameter value.
func (c *Context) SetParam(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.params[key] = value
}

// Query returns the URL query parameter matching key.
func (c *Context) Query(key string) string {
	if c.Request == nil || c.Request.URL == nil {
		return ""
	}
	return c.Request.URL.Query().Get(key)
}

// QueryDefault returns the URL query parameter matching key, or fallback if empty.
func (c *Context) QueryDefault(key, fallback string) string {
	val := c.Query(key)
	if val == "" {
		return fallback
	}
	return val
}

// Header returns the incoming request HTTP header for key.
func (c *Context) Header(key string) string {
	if c.Request == nil {
		return ""
	}
	return c.Request.Header.Get(key)
}

// SetHeader sets an outgoing HTTP response header.
func (c *Context) SetHeader(key, value string) {
	if c.Response != nil {
		c.Response.Header().Set(key, value)
	}
}

// BindJSON decodes the JSON request body into the target pointer.
func (c *Context) BindJSON(dst any) error {
	if c.Request == nil || c.Request.Body == nil {
		return errors.New("core: request body is empty")
	}
	dec := json.NewDecoder(c.Request.Body)
	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("core: failed decoding json body: %w", err)
	}
	return c.Validate(dst)
}

// BindForm parses the URL-encoded or multipart form data into the destination struct.
func (c *Context) BindForm(dst any) error {
	if c.Request == nil {
		return errors.New("core: nil request")
	}
	if err := c.Request.ParseForm(); err != nil {
		return fmt.Errorf("core: failed parsing form data: %w", err)
	}
	if err := mapValuesToStruct(c.Request.Form, dst, "form"); err != nil {
		return err
	}
	return c.Validate(dst)
}

// BindQuery parses URL query parameters into the destination struct.
func (c *Context) BindQuery(dst any) error {
	if c.Request == nil || c.Request.URL == nil {
		return errors.New("core: nil request URL")
	}
	if err := mapValuesToStruct(c.Request.URL.Query(), dst, "query"); err != nil {
		return err
	}
	return c.Validate(dst)
}

// Validate executes validation on the destination if it implements Validator or SchemaValidator.
func (c *Context) Validate(v any) error {
	if val, ok := v.(Validator); ok {
		return val.Validate()
	}
	if val, ok := v.(interface{ Validate(v any) error }); ok {
		return val.Validate(v)
	}
	return nil
}

func mapValuesToStruct(values url.Values, dst any, tagKey string) error {
	val := reflect.ValueOf(dst)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return ErrUnsupportedBindingTarget
	}

	elem := val.Elem()
	if elem.Kind() != reflect.Struct {
		return ErrUnsupportedBindingTarget
	}

	typ := elem.Type()
	for i := 0; i < elem.NumField(); i++ {
		field := elem.Field(i)
		if !field.CanSet() {
			continue
		}

		structField := typ.Field(i)
		tag := structField.Tag.Get(tagKey)
		if tag == "-" {
			continue
		}
		if tag == "" {
			tag = strings.ToLower(structField.Name)
		}

		tagParts := strings.Split(tag, ",")
		fieldName := tagParts[0]

		if vals, exists := values[fieldName]; exists && len(vals) > 0 {
			rawVal := vals[0]
			if err := setFieldValue(field, rawVal); err != nil {
				return fmt.Errorf("core: field %s: %w", structField.Name, err)
			}
		}
	}
	return nil
}

func setFieldValue(field reflect.Value, raw string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(raw)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(u)
	case reflect.Bool:
		b, err := strconv.ParseBool(raw)
		if err != nil {
			return err
		}
		field.SetBool(b)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return err
		}
		field.SetFloat(f)
	}
	return nil
}
