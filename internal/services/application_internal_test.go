package services

import (
	"errors"
	"os"
	"testing"

	"gorm.io/gorm"
)

func TestMustJSON_Marshallable(t *testing.T) {
	result := mustJSON(map[string]any{"key": "value"})
	if string(result) == "{}" {
		t.Error("expected non-empty JSON, got fallback {}")
	}
}

func TestMustJSON_Unmarshallable(t *testing.T) {
	// channels cannot be JSON-marshalled — hits the error path
	result := mustJSON(make(chan int))
	if string(result) != "{}" {
		t.Errorf("expected fallback {}, got %s", result)
	}
}

func TestIsApplicationRecordNotFound_Nil(t *testing.T) {
	if isApplicationRecordNotFound(nil) {
		t.Error("expected false for nil error")
	}
}

func TestIsApplicationRecordNotFound_GormErr(t *testing.T) {
	if !isApplicationRecordNotFound(gorm.ErrRecordNotFound) {
		t.Error("expected true for gorm.ErrRecordNotFound")
	}
}

func TestIsApplicationRecordNotFound_StringMatch(t *testing.T) {
	err := errors.New("some record not found in database")
	if !isApplicationRecordNotFound(err) {
		t.Error("expected true for error message containing 'record not found'")
	}
}

func TestIsApplicationRecordNotFound_Unrelated(t *testing.T) {
	err := errors.New("connection refused")
	if isApplicationRecordNotFound(err) {
		t.Error("expected false for unrelated error")
	}
}

func TestDefaultAdminPassword_EnvVar(t *testing.T) {
	t.Setenv("DEFAULT_ADMIN_PASSWORD", "my-secure-password")
	pwd := defaultAdminPassword()
	if pwd != "my-secure-password" {
		t.Errorf("expected env var password, got %q", pwd)
	}
}

func TestDefaultAdminPassword_Default(t *testing.T) {
	os.Unsetenv("DEFAULT_ADMIN_PASSWORD")
	pwd := defaultAdminPassword()
	if pwd == "" {
		t.Error("expected non-empty default password")
	}
}
