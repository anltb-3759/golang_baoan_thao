package utils

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSanitizeFileName_Valid(t *testing.T) {
	cases := []struct {
		input    string
		wantPart string
	}{
		{"report.pdf", "report.pdf"},
		{"my-file_2024.png", "my-file_2024.png"},
		{"path/to/file.jpg", "file.jpg"},
	}
	for _, tc := range cases {
		got, err := sanitizeFileName(tc.input)
		if err != nil {
			t.Fatalf("sanitizeFileName(%q): unexpected error: %v", tc.input, err)
		}
		if got != tc.wantPart {
			t.Fatalf("sanitizeFileName(%q) = %q, want %q", tc.input, got, tc.wantPart)
		}
	}
}

func TestSanitizeFileName_EmptyName(t *testing.T) {
	_, err := sanitizeFileName("")
	if err != ErrEmptyFileName {
		t.Fatalf("expected ErrEmptyFileName, got %v", err)
	}
}

func TestSanitizeFileName_DotFile(t *testing.T) {
	_, err := sanitizeFileName(".hidden")
	if err != ErrUnsafeFileName {
		t.Fatalf("expected ErrUnsafeFileName for dotfile, got %v", err)
	}
}

func TestSanitizeFileName_DotDot(t *testing.T) {
	// base name itself contains ..
	_, err := sanitizeFileName("..evil.txt")
	if err != ErrUnsafeFileName {
		t.Fatalf("expected ErrUnsafeFileName for .. in name, got %v", err)
	}
}

func TestSanitizeFileName_SpecialCharsReplaced(t *testing.T) {
	got, err := sanitizeFileName("my file (1).pdf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(got, " ") || strings.Contains(got, "(") {
		t.Fatalf("expected special chars replaced, got %q", got)
	}
}

func TestUniquifyFileName_Format(t *testing.T) {
	name := uniquifyFileName("document.pdf")
	if !strings.HasSuffix(name, ".pdf") {
		t.Fatalf("expected .pdf suffix, got %q", name)
	}
	if !strings.HasPrefix(name, "document-") {
		t.Fatalf("expected document- prefix, got %q", name)
	}
}

func TestUniquifyFileName_Unique(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		n := uniquifyFileName("file.jpg")
		if seen[n] {
			t.Fatalf("duplicate filename: %q", n)
		}
		seen[n] = true
	}
}

func TestUniquifyFileName_NoExtension(t *testing.T) {
	name := uniquifyFileName("myfile")
	if !strings.HasPrefix(name, "myfile-") {
		t.Fatalf("expected myfile- prefix, got %q", name)
	}
}

func TestNewLocalDiskStorage(t *testing.T) {
	s := NewLocalDiskStorage("/tmp/uploads", "http://localhost/files/")
	if s.BaseDir != "/tmp/uploads" {
		t.Fatalf("expected BaseDir /tmp/uploads, got %q", s.BaseDir)
	}
	if s.PublicPrefix != "http://localhost/files" {
		t.Fatalf("expected PublicPrefix without trailing slash, got %q", s.PublicPrefix)
	}
}

func TestRemoveApplicationDir_Empty(t *testing.T) {
	s := NewLocalDiskStorage("/tmp", "/files")
	err := s.RemoveApplicationDir("")
	if err != nil {
		t.Fatalf("expected nil for empty applicationID, got %v", err)
	}
}

func TestRemoveApplicationDir_Existing(t *testing.T) {
	dir := t.TempDir()
	s := NewLocalDiskStorage(dir, "/files")

	appDir := filepath.Join(dir, "applications", "app-test")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		t.Fatal(err)
	}

	err := s.RemoveApplicationDir("app-test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(appDir); !os.IsNotExist(err) {
		t.Fatal("expected directory to be removed")
	}
}

func TestRemoveFile_EmptyURL(t *testing.T) {
	s := NewLocalDiskStorage("/tmp", "/files")
	err := s.RemoveFile("")
	if err != nil {
		t.Fatalf("expected nil for empty URL, got %v", err)
	}
}

func TestRemoveFile_WithPrefix(t *testing.T) {
	dir := t.TempDir()
	s := NewLocalDiskStorage(dir, "/files")

	// Create a file to remove
	appDir := filepath.Join(dir, "applications", "app-1")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		t.Fatal(err)
	}
	filePath := filepath.Join(appDir, "test.pdf")
	if err := os.WriteFile(filePath, []byte("content"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := s.RemoveFile("/files/applications/app-1/test.pdf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Fatal("expected file to be removed")
	}
}

func TestRemoveFile_WithFallbackPath(t *testing.T) {
	dir := t.TempDir()
	s := NewLocalDiskStorage(dir, "/files")

	appDir := filepath.Join(dir, "applications", "app-2")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		t.Fatal(err)
	}
	filePath := filepath.Join(appDir, "doc.pdf")
	if err := os.WriteFile(filePath, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}

	// URL without the expected prefix — triggers fallback
	err := s.RemoveFile("http://other-host/applications/app-2/doc.pdf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRemoveFile_NoApplicationsPath(t *testing.T) {
	s := NewLocalDiskStorage("/tmp", "/files")
	// URL with no /applications/ segment — should be a no-op
	err := s.RemoveFile("http://other-host/static/image.png")
	if err != nil {
		t.Fatalf("expected nil for unrelated URL, got %v", err)
	}
}

func TestSaveApplicationFile_DisallowedMime(t *testing.T) {
	dir := t.TempDir()
	s := NewLocalDiskStorage(dir, "/files")

	// Create a multipart form with a text file (mime: text/plain — not allowed)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.txt")
	_, _ = io.WriteString(part, "hello world plain text")
	writer.Close()

	req, _ := http.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if err := req.ParseMultipartForm(10 << 20); err != nil {
		t.Fatal(err)
	}

	fh := req.MultipartForm.File["file"][0]
	_, _, _, err := s.SaveApplicationFile("app-1", fh)
	if err != ErrDisallowedMime {
		t.Fatalf("expected ErrDisallowedMime, got %v", err)
	}
}

func TestSaveApplicationFile_UnsafeFileName(t *testing.T) {
	dir := t.TempDir()
	s := NewLocalDiskStorage(dir, "/files")

	// dotfile name which sanitizeFileName rejects
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", ".hidden")
	_, _ = io.WriteString(part, "content")
	writer.Close()

	req, _ := http.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	_ = req.ParseMultipartForm(10 << 20)

	fh := req.MultipartForm.File["file"][0]
	_, _, _, err := s.SaveApplicationFile("app-1", fh)
	if err != ErrUnsafeFileName {
		t.Fatalf("expected ErrUnsafeFileName, got %v", err)
	}
}
