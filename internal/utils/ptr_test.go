package utils

import (
	"os"
	"testing"
)

func TestSetIfNotNil_SetsWhenSrcNotNil(t *testing.T) {
	dst := 1
	src := 42
	SetIfNotNil(&dst, &src)
	if dst != 42 {
		t.Fatalf("expected 42, got %d", dst)
	}
}

func TestSetIfNotNil_NoOpWhenSrcNil(t *testing.T) {
	dst := 1
	SetIfNotNil(&dst, (*int)(nil))
	if dst != 1 {
		t.Fatalf("expected 1, got %d", dst)
	}
}

func TestSetIfNotNil_WorksWithString(t *testing.T) {
	dst := "original"
	src := "updated"
	SetIfNotNil(&dst, &src)
	if dst != "updated" {
		t.Fatalf("expected 'updated', got %q", dst)
	}
}

func TestEnvOr_ReturnsFallbackWhenUnset(t *testing.T) {
	os.Unsetenv("TEST_ENVOR_KEY_UNUSED")
	got := EnvOr("TEST_ENVOR_KEY_UNUSED", "fallback")
	if got != "fallback" {
		t.Fatalf("expected 'fallback', got %q", got)
	}
}

func TestEnvOr_ReturnsEnvVarWhenSet(t *testing.T) {
	os.Setenv("TEST_ENVOR_KEY", "myvalue")
	defer os.Unsetenv("TEST_ENVOR_KEY")
	got := EnvOr("TEST_ENVOR_KEY", "fallback")
	if got != "myvalue" {
		t.Fatalf("expected 'myvalue', got %q", got)
	}
}

func TestEnvOr_ReturnsFallbackWhenEmpty(t *testing.T) {
	os.Setenv("TEST_ENVOR_EMPTY", "")
	defer os.Unsetenv("TEST_ENVOR_EMPTY")
	got := EnvOr("TEST_ENVOR_EMPTY", "fallback")
	if got != "fallback" {
		t.Fatalf("expected 'fallback', got %q", got)
	}
}
