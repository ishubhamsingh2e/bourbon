// Package config defines application configuration models, environment types, and zero-dependency environment loading.
package config

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Option configures the configuration Loader behavior.
type Option func(*Loader)

// WithDefaults sets default key-value pairs for configuration.
func WithDefaults(defaults map[string]any) Option {
	return func(l *Loader) {
		for k, v := range defaults {
			l.defaults[k] = v
		}
	}
}

// WithConfigFile specifies a custom configuration file path.
func WithConfigFile(file, fileType string) Option {
	return func(l *Loader) {
		l.configFile = file
	}
}

// WithEnvPrefix sets an environment variable prefix for automatic lookup.
func WithEnvPrefix(prefix string) Option {
	return func(l *Loader) {
		l.prefix = prefix
	}
}

// Loader encapsulates configuration loading, .env parsing, and typed struct binding.
type Loader struct {
	configFile string
	prefix     string
	defaults   map[string]any
	fileValues map[string]string
}

// NewLoader constructs a new configuration Loader with optional configuration functions.
func NewLoader(opts ...Option) *Loader {
	l := &Loader{
		configFile: ".env",
		defaults:   make(map[string]any),
		fileValues: make(map[string]string),
	}
	for _, opt := range opts {
		opt(l)
	}

	if l.configFile != "" {
		l.fileValues = parseEnvFile(l.configFile)
	}
	return l
}

// Bind unmarshals the loaded configuration into the target struct.
func (l *Loader) Bind(target any) error {
	val := reflect.ValueOf(target)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("config: target must be a pointer to a struct, got %T", target)
	}

	return l.bindStruct(val.Elem())
}

func (l *Loader) bindStruct(structVal reflect.Value) error {
	structType := structVal.Type()

	for i := 0; i < structVal.NumField(); i++ {
		field := structType.Field(i)
		fieldVal := structVal.Field(i)

		if !fieldVal.CanSet() {
			continue
		}

		if field.Anonymous && fieldVal.Kind() == reflect.Struct {
			if err := l.bindStruct(fieldVal); err != nil {
				return err
			}
			continue
		}

		if fieldVal.Kind() == reflect.Struct && field.Type != reflect.TypeOf(time.Time{}) {
			if err := l.bindStruct(fieldVal); err != nil {
				return err
			}
			continue
		}

		tag := field.Tag.Get("env")
		if tag == "" {
			tag = field.Tag.Get("mapstructure")
		}

		if tag == "-" {
			continue
		}

		key, defaultVal := parseTag(tag, field.Name)
		if l.prefix != "" && key != "" {
			key = l.prefix + "_" + key
		}

		rawVal := l.lookupValue(key, defaultVal)
		if rawVal == "" && defaultVal == "" {
			continue
		}

		if err := setFieldValue(fieldVal, rawVal); err != nil {
			return fmt.Errorf("config: error setting field %s (%s): %w", field.Name, key, err)
		}
	}

	return nil
}

func (l *Loader) lookupValue(key, defaultVal string) string {
	// 1. System environment variable takes precedence
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}

	// 2. Value from .env file
	if v, ok := l.fileValues[key]; ok && v != "" {
		return v
	}

	// 3. Provided defaults map
	if v, ok := l.defaults[key]; ok {
		return fmt.Sprintf("%v", v)
	}

	// 4. Default value from struct tag
	return defaultVal
}

func parseTag(tag, fallbackFieldName string) (key string, defaultVal string) {
	if tag == "" {
		return toEnvKey(fallbackFieldName), ""
	}

	parts := strings.Split(tag, ",")
	key = strings.TrimSpace(parts[0])
	if key == "" {
		key = toEnvKey(fallbackFieldName)
	}

	for _, opt := range parts[1:] {
		opt = strings.TrimSpace(opt)
		if strings.HasPrefix(opt, "default=") {
			defaultVal = strings.TrimPrefix(opt, "default=")
		}
	}

	return key, defaultVal
}

func toEnvKey(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToUpper(result.String())
}

func setFieldValue(v reflect.Value, raw string) error {
	// Special case: time.Duration
	if v.Type() == reflect.TypeOf(time.Duration(0)) {
		d, err := time.ParseDuration(raw)
		if err != nil {
			return err
		}
		v.SetInt(int64(d))
		return nil
	}

	switch v.Kind() {
	case reflect.String:
		v.SetString(raw)
	case reflect.Bool:
		b, err := strconv.ParseBool(raw)
		if err != nil {
			return err
		}
		v.SetBool(b)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return err
		}
		v.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			return err
		}
		v.SetUint(u)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return err
		}
		v.SetFloat(f)
	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.String {
			items := strings.Split(raw, ",")
			slice := reflect.MakeSlice(v.Type(), 0, len(items))
			for _, item := range items {
				trimmed := strings.TrimSpace(item)
				if trimmed != "" {
					slice = reflect.Append(slice, reflect.ValueOf(trimmed))
				}
			}
			v.Set(slice)
		}
	}

	return nil
}

func parseEnvFile(path string) map[string]string {
	result := make(map[string]string)
	f, err := os.Open(path)
	if err != nil {
		return result
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			k := strings.TrimSpace(parts[0])
			v := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
			result[k] = v
		}
	}
	return result
}

// Load loads environment variables into the target struct.
func Load(target any, opts ...Option) error {
	loader := NewLoader(opts...)
	return loader.Bind(target)
}

// LoadEnv loads environment variables and .env file into the target struct with optional defaults.
func LoadEnv(target any, opts ...Option) error {
	return Load(target, opts...)
}

// GetString returns the environment variable value for the given key, or fallback if unset or empty.
func GetString(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}

// MustGetString returns the environment variable value for key, or panics if unset or empty.
func MustGetString(key string) string {
	val, ok := os.LookupEnv(key)
	if !ok || val == "" {
		panic(fmt.Sprintf("config: required environment variable %q is not set", key))
	}
	return val
}

// GetInt returns the parsed integer from the environment variable for key, or fallback on error/missing.
func GetInt(key string, fallback int) int {
	val, ok := os.LookupEnv(key)
	if !ok || val == "" {
		return fallback
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}
	return i
}

// MustGetInt returns the parsed integer from the environment variable for key, or panics on error/missing.
func MustGetInt(key string) int {
	val := MustGetString(key)
	i, err := strconv.Atoi(val)
	if err != nil {
		panic(fmt.Sprintf("config: environment variable %q is not a valid integer: %v", key, err))
	}
	return i
}

// GetBool returns the parsed boolean from the environment variable for key, or fallback on error/missing.
func GetBool(key string, fallback bool) bool {
	val, ok := os.LookupEnv(key)
	if !ok || val == "" {
		return fallback
	}
	b, err := strconv.ParseBool(val)
	if err != nil {
		return fallback
	}
	return b
}

// GetDuration returns the parsed time.Duration from the environment variable for key, or fallback on error/missing.
func GetDuration(key string, fallback time.Duration) time.Duration {
	val, ok := os.LookupEnv(key)
	if !ok || val == "" {
		return fallback
	}
	d, err := time.ParseDuration(val)
	if err != nil {
		return fallback
	}
	return d
}

// LoadFromEnv loads configuration from standard environment variables with sensible defaults.
func LoadFromEnv() *Config {
	envStr := GetString("APP_ENV", string(EnvDevelopment))
	env := Environment(envStr)

	return &Config{
		App: AppConfig{
			Name:     GetString("APP_NAME", "Bourbon App"),
			Key:      GetString("APP_KEY", ""),
			Debug:    GetBool("APP_DEBUG", env == EnvDevelopment),
			URL:      GetString("APP_URL", "http://localhost:8080"),
			Timezone: GetString("APP_TIMEZONE", "UTC"),
		},
		HTTP: HTTPConfig{
			Host:         GetString("HTTP_HOST", "0.0.0.0"),
			Port:         GetInt("HTTP_PORT", 8080),
			ReadTimeout:  GetDuration("HTTP_READ_TIMEOUT", 15*time.Second),
			WriteTimeout: GetDuration("HTTP_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:  GetDuration("HTTP_IDLE_TIMEOUT", 60*time.Second),
		},
		Env: env,
	}
}
