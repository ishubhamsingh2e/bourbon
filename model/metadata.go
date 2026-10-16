package model

import (
	"reflect"
	"strings"
	"unicode"
)

// FieldMeta describes metadata about a single model struct field.
type FieldMeta struct {
	Name       string
	Type       reflect.Type
	Column     string
	PrimaryKey bool
	Nullable   bool
}

// ModelMeta provides structural and mapping metadata for a domain model.
type ModelMeta struct {
	Name       string
	TableName  string
	Fields     []FieldMeta
	PrimaryCol string
}

// Inspect extracts metadata from a model struct or pointer using reflection.
func Inspect(v any) *ModelMeta {
	if v == nil {
		return nil
	}

	typ := reflect.TypeOf(v)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return nil
	}

	meta := &ModelMeta{
		Name:   typ.Name(),
		Fields: make([]FieldMeta, 0),
	}

	// Check if v implements Model for custom TableName
	if m, ok := v.(Model); ok && m.TableName() != "" {
		meta.TableName = m.TableName()
	} else {
		// Fallback: pluralize lowercase struct name
		meta.TableName = toSnakeCase(typ.Name()) + "s"
	}

	inspectStructFields(typ, meta)

	for _, f := range meta.Fields {
		if f.PrimaryKey {
			meta.PrimaryCol = f.Column
			break
		}
	}
	if meta.PrimaryCol == "" && len(meta.Fields) > 0 {
		meta.PrimaryCol = meta.Fields[0].Column
	}

	return meta
}

func inspectStructFields(typ reflect.Type, meta *ModelMeta) {
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if !field.IsExported() {
			continue
		}

		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			inspectStructFields(field.Type, meta)
			continue
		}

		dbTag := field.Tag.Get("db")
		if dbTag == "-" {
			continue
		}

		colName := toSnakeCase(field.Name)
		isPrimary := false
		isNullable := field.Type.Kind() == reflect.Ptr

		if dbTag != "" {
			parts := strings.Split(dbTag, ",")
			if parts[0] != "" {
				colName = parts[0]
			}
			for _, opt := range parts[1:] {
				if opt == "primary" || opt == "pk" {
					isPrimary = true
				}
			}
		}

		if strings.EqualFold(field.Name, "id") || strings.EqualFold(colName, "id") {
			isPrimary = true
		}

		meta.Fields = append(meta.Fields, FieldMeta{
			Name:       field.Name,
			Type:       field.Type,
			Column:     colName,
			PrimaryKey: isPrimary,
			Nullable:   isNullable,
		})
	}
}

func toSnakeCase(s string) string {
	if s == "ID" {
		return "id"
	}
	var res []rune
	runes := []rune(s)
	for i, r := range runes {
		if unicode.IsUpper(r) {
			if i > 0 {
				prev := runes[i-1]
				nextIsLower := i+1 < len(runes) && unicode.IsLower(runes[i+1])
				if unicode.IsLower(prev) || (unicode.IsUpper(prev) && nextIsLower) {
					res = append(res, '_')
				}
			}
			res = append(res, unicode.ToLower(r))
		} else {
			res = append(res, r)
		}
	}
	return string(res)
}
