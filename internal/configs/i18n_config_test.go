package configs

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/labstack/echo/v5"
)

func newI18nContext(t *testing.T) *echo.Context {
	t.Helper()
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec)
}

func TestLoadI18nMessages_Success(t *testing.T) {
	// Reset messages so we can reload
	messages = map[string]map[string]string{}
	if err := LoadI18nMessages("../../locales"); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if _, ok := messages["vi"]; !ok {
		t.Fatal("expected 'vi' locale to be loaded")
	}
}

func TestLoadI18nMessages_BadDir(t *testing.T) {
	messages = map[string]map[string]string{}
	if err := LoadI18nMessages("/nonexistent/path"); err == nil {
		t.Fatal("expected error for nonexistent directory")
	}
}

func TestLoadI18nMessages_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(dir+"/vi.json", []byte("not valid json"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	messages = map[string]map[string]string{}
	if err := LoadI18nMessages(dir); err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestLoadI18nMessages_MissingDefaultLocale(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(dir+"/en.json", []byte(`{"key":"value"}`), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	messages = map[string]map[string]string{}
	if err := LoadI18nMessages(dir); err == nil {
		t.Fatal("expected error when default locale 'vi' is missing")
	}
}

func TestNormalizeLocale_Empty(t *testing.T) {
	messages = map[string]map[string]string{"vi": {"key": "val"}}
	if got := NormalizeLocale(""); got != DefaultLocale {
		t.Fatalf("expected %q, got %q", DefaultLocale, got)
	}
}

func TestNormalizeLocale_Known(t *testing.T) {
	messages = map[string]map[string]string{"vi": {}, "en": {}}
	if got := NormalizeLocale("en"); got != "en" {
		t.Fatalf("expected en, got %q", got)
	}
}

func TestNormalizeLocale_Unknown(t *testing.T) {
	messages = map[string]map[string]string{"vi": {}}
	if got := NormalizeLocale("zh"); got != DefaultLocale {
		t.Fatalf("expected %q, got %q", DefaultLocale, got)
	}
}

func TestNormalizeLocale_WithQualityAndRegion(t *testing.T) {
	messages = map[string]map[string]string{"vi": {}, "en": {}}
	// "en-US,en;q=0.9" -> extracts "en-US" -> then "en"
	if got := NormalizeLocale("en-US,en;q=0.9"); got != "en" {
		t.Fatalf("expected en, got %q", got)
	}
}

func TestLocaleFromContext_WithLocale(t *testing.T) {
	messages = map[string]map[string]string{"vi": {}, "en": {}}
	c := newI18nContext(t)
	c.Set(LocaleKey, "en")

	if got := LocaleFromContext(c); got != "en" {
		t.Fatalf("expected en, got %q", got)
	}
}

func TestLocaleFromContext_WithoutLocale(t *testing.T) {
	messages = map[string]map[string]string{"vi": {}}
	c := newI18nContext(t)

	if got := LocaleFromContext(c); got != DefaultLocale {
		t.Fatalf("expected %q, got %q", DefaultLocale, got)
	}
}

func TestTLang_ReturnsTranslation(t *testing.T) {
	messages = map[string]map[string]string{
		"vi": {"greeting": "Xin chào {name}"},
		"en": {"greeting": "Hello {name}"},
	}
	got := TLang("en", "greeting", map[string]string{"name": "An"})
	if got != "Hello An" {
		t.Fatalf("expected 'Hello An', got %q", got)
	}
}

func TestTLang_FallsBackToDefault(t *testing.T) {
	messages = map[string]map[string]string{
		"vi": {"greeting": "Xin chào"},
	}
	got := TLang("en", "greeting", nil)
	if got != "Xin chào" {
		t.Fatalf("expected fallback 'Xin chào', got %q", got)
	}
}

func TestTLang_ReturnsKeyWhenMissing(t *testing.T) {
	messages = map[string]map[string]string{"vi": {}}
	got := TLang("vi", "missing.key", nil)
	if got != "missing.key" {
		t.Fatalf("expected key passthrough, got %q", got)
	}
}

func TestT_ReturnsTranslation(t *testing.T) {
	messages = map[string]map[string]string{
		"vi": {"hello": "Xin chào {name}"},
	}
	c := newI18nContext(t)
	c.Set(LocaleKey, "vi")

	got := T(c, "hello", map[string]string{"name": "Bảo"})
	if got != "Xin chào Bảo" {
		t.Fatalf("expected 'Xin chào Bảo', got %q", got)
	}
}
