package templates

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
)

func newTestEcho() *echo.Echo {
	return echo.New()
}

func TestNewRenderer_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	r, err := NewRenderer(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r == nil {
		t.Fatal("expected non-nil renderer")
	}
}

func TestNewRenderer_WithHTML(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "hello.html"), []byte(`Hello {{.Name}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	r, err := NewRenderer(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r == nil {
		t.Fatal("expected non-nil renderer")
	}
}

func TestNewRenderer_InvalidDir(t *testing.T) {
	_, err := NewRenderer("/nonexistent/path/xyz")
	if err == nil {
		t.Fatal("expected error for nonexistent directory")
	}
}

func TestRenderer_Render(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "page.html"), []byte(`Hello {{.Name}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	r, err := NewRenderer(dir)
	if err != nil {
		t.Fatalf("setup error: %v", err)
	}

	e := newTestEcho()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var buf bytes.Buffer
	if err := r.Render(c, &buf, "page.html", map[string]string{"Name": "An"}); err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if got := buf.String(); got != "Hello An" {
		t.Fatalf("expected 'Hello An', got %q", got)
	}
}

func TestRenderer_Render_UnknownTemplate(t *testing.T) {
	dir := t.TempDir()
	r, err := NewRenderer(dir)
	if err != nil {
		t.Fatalf("setup error: %v", err)
	}

	e := newTestEcho()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var buf bytes.Buffer
	err = r.Render(c, &buf, "nonexistent.html", nil)
	if err == nil {
		t.Fatal("expected error for unknown template")
	}
}

func TestRenderer_Render_WithFmtTime(t *testing.T) {
	dir := t.TempDir()
	tmplContent := `{{fmtTime .T "2006-01-02"}}`
	if err := os.WriteFile(filepath.Join(dir, "time.html"), []byte(tmplContent), 0o644); err != nil {
		t.Fatal(err)
	}
	r, err := NewRenderer(dir)
	if err != nil {
		t.Fatalf("setup error: %v", err)
	}

	e := newTestEcho()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	ts, _ := time.Parse(time.RFC3339, "2024-03-15T10:00:00Z")

	cases := []struct {
		name string
		data any
		want string
	}{
		{"time.Time", map[string]any{"T": ts}, "2024-03-15"},
		{"*time.Time", map[string]any{"T": &ts}, "2024-03-15"},
		{"string RFC3339", map[string]any{"T": "2024-03-15T10:00:00Z"}, "2024-03-15"},
		{"string YYYY-MM-DD HH:MM", map[string]any{"T": "2024-03-15 10:00"}, "2024-03-15 10:00"},
		{"nil", map[string]any{"T": nil}, ""},
		{"zero time", map[string]any{"T": time.Time{}}, ""},
		{"empty string", map[string]any{"T": ""}, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := r.Render(c, &buf, "time.html", tc.data); err != nil {
				t.Fatalf("Render error: %v", err)
			}
			if got := buf.String(); got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}
}

func TestNewRenderer_InvalidTemplate(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "bad.html"), []byte(`{{invalid template`), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := NewRenderer(dir)
	if err == nil {
		t.Fatal("expected error for invalid template syntax")
	}
}

func TestRenderer_Render_WithDefault(t *testing.T) {
	dir := t.TempDir()
	tmplContent := `{{default .V "fallback"}}`
	if err := os.WriteFile(filepath.Join(dir, "def.html"), []byte(tmplContent), 0o644); err != nil {
		t.Fatal(err)
	}
	r, err := NewRenderer(dir)
	if err != nil {
		t.Fatalf("setup error: %v", err)
	}

	e := newTestEcho()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	cases := []struct {
		name string
		data any
		want string
	}{
		{"non-empty string", map[string]any{"V": "hello"}, "hello"},
		{"empty string", map[string]any{"V": ""}, "fallback"},
		{"nil", map[string]any{"V": nil}, "fallback"},
		{"non-string int", map[string]any{"V": 42}, "fallback"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := r.Render(c, &buf, "def.html", tc.data); err != nil {
				t.Fatalf("Render error: %v", err)
			}
			if got := buf.String(); got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}
}
