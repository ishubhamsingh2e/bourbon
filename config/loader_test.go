package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoaderDefaults(t *testing.T) {
	os.Unsetenv("TEST_STRING")
	os.Unsetenv("TEST_INT")
	os.Unsetenv("TEST_BOOL")
	os.Unsetenv("TEST_DURATION")

	if got := GetString("TEST_STRING", "default"); got != "default" {
		t.Errorf("GetString() = %v, want %v", got, "default")
	}

	if got := GetInt("TEST_INT", 42); got != 42 {
		t.Errorf("GetInt() = %v, want %v", got, 42)
	}

	if got := GetBool("TEST_BOOL", true); got != true {
		t.Errorf("GetBool() = %v, want %v", got, true)
	}

	if got := GetDuration("TEST_DURATION", 10*time.Second); got != 10*time.Second {
		t.Errorf("GetDuration() = %v, want %v", got, 10*time.Second)
	}
}

func TestLoaderEnvOverrides(t *testing.T) {
	t.Setenv("TEST_STR_SET", "custom_val")
	t.Setenv("TEST_INT_SET", "8081")
	t.Setenv("TEST_BOOL_SET", "true")
	t.Setenv("TEST_DUR_SET", "5m")

	if got := GetString("TEST_STR_SET", "fallback"); got != "custom_val" {
		t.Errorf("GetString() = %v, want %v", got, "custom_val")
	}

	if got := MustGetString("TEST_STR_SET"); got != "custom_val" {
		t.Errorf("MustGetString() = %v, want %v", got, "custom_val")
	}

	if got := GetInt("TEST_INT_SET", 80); got != 8081 {
		t.Errorf("GetInt() = %v, want %v", got, 8081)
	}

	if got := MustGetInt("TEST_INT_SET"); got != 8081 {
		t.Errorf("MustGetInt() = %v, want %v", got, 8081)
	}

	if got := GetBool("TEST_BOOL_SET", false); got != true {
		t.Errorf("GetBool() = %v, want %v", got, true)
	}

	if got := GetDuration("TEST_DUR_SET", time.Second); got != 5*time.Minute {
		t.Errorf("GetDuration() = %v, want %v", got, 5*time.Minute)
	}
}

func TestMustGetStringPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("MustGetString did not panic on missing env variable")
		}
	}()
	os.Unsetenv("DEFINITELY_NOT_SET_KEY")
	MustGetString("DEFINITELY_NOT_SET_KEY")
}

func TestMustGetIntPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("MustGetInt did not panic on invalid int")
		}
	}()
	t.Setenv("INVALID_INT_KEY", "not_a_number")
	MustGetInt("INVALID_INT_KEY")
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("APP_NAME", "Test Bourbon")
	t.Setenv("APP_ENV", "production")
	t.Setenv("HTTP_PORT", "9000")
	t.Setenv("HTTP_READ_TIMEOUT", "30s")

	cfg := LoadFromEnv()
	if cfg.App.Name != "Test Bourbon" {
		t.Errorf("cfg.App.Name = %v, want %v", cfg.App.Name, "Test Bourbon")
	}
	if cfg.Env != EnvProduction {
		t.Errorf("cfg.Env = %v, want %v", cfg.Env, EnvProduction)
	}
	if cfg.HTTP.Port != 9000 {
		t.Errorf("cfg.HTTP.Port = %v, want %v", cfg.HTTP.Port, 9000)
	}
	if cfg.HTTP.ReadTimeout != 30*time.Second {
		t.Errorf("cfg.HTTP.ReadTimeout = %v, want %v", cfg.HTTP.ReadTimeout, 30*time.Second)
	}
}

func TestLoadEnvStructBinding(t *testing.T) {
	type CustomSettings struct {
		AppName string        `env:"APP_NAME,default=default_app"`
		Port    int           `env:"PORT,default=8080"`
		Debug   bool          `env:"DEBUG,default=false"`
		Timeout time.Duration `env:"TIMEOUT,default=15s"`
	}

	t.Setenv("PORT", "9090")
	t.Setenv("DEBUG", "true")

	var s CustomSettings
	err := Load(&s)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if s.AppName != "default_app" {
		t.Errorf("expected AppName 'default_app', got: %s", s.AppName)
	}
	if s.Port != 9090 {
		t.Errorf("expected Port 9090, got: %d", s.Port)
	}
	if !s.Debug {
		t.Errorf("expected Debug true, got false")
	}
	if s.Timeout != 15*time.Second {
		t.Errorf("expected Timeout 15s, got %v", s.Timeout)
	}
}

func TestLoadEnvFileParsing(t *testing.T) {
	tempDir := t.TempDir()
	envPath := filepath.Join(tempDir, ".env")
	envContent := `
# Sample env file
APP_NAME="Todo App"
PORT=3000
IS_ACTIVE=true
`
	if err := os.WriteFile(envPath, []byte(envContent), 0644); err != nil {
		t.Fatalf("failed writing .env: %v", err)
	}

	type AppSettings struct {
		AppName  string `env:"APP_NAME"`
		Port     int    `env:"PORT"`
		IsActive bool   `env:"IS_ACTIVE"`
	}

	var s AppSettings
	err := Load(&s, WithConfigFile(envPath, "env"))
	if err != nil {
		t.Fatalf("Load with config file failed: %v", err)
	}

	if s.AppName != "Todo App" {
		t.Errorf("expected AppName 'Todo App', got: %s", s.AppName)
	}
	if s.Port != 3000 {
		t.Errorf("expected Port 3000, got: %d", s.Port)
	}
	if !s.IsActive {
		t.Errorf("expected IsActive true, got false")
	}
}
