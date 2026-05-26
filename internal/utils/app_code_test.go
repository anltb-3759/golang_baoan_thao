package utils

import (
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestGenerateUUID_Format(t *testing.T) {
	uuid := GenerateUUID()
	pattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	if !pattern.MatchString(uuid) {
		t.Fatalf("invalid UUID v4 format: %q", uuid)
	}
}

func TestGenerateUUID_Unique(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		u := GenerateUUID()
		if seen[u] {
			t.Fatalf("duplicate UUID generated: %q", u)
		}
		seen[u] = true
	}
}

func TestGenerateApplicationCodeAt_Format(t *testing.T) {
	ts := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	code := GenerateApplicationCodeAt(ts)

	if !strings.HasPrefix(code, "APP-20240315-") {
		t.Fatalf("expected prefix APP-20240315-, got %q", code)
	}
	parts := strings.Split(code, "-")
	// APP-YYYYMMDD-XXXXXX => 3 parts
	if len(parts) != 3 {
		t.Fatalf("expected 3 parts, got %d: %q", len(parts), code)
	}
	suffix := parts[2]
	if len(suffix) != appCodeSuffixLen {
		t.Fatalf("expected suffix length %d, got %d: %q", appCodeSuffixLen, len(suffix), suffix)
	}
	for _, ch := range suffix {
		if !strings.ContainsRune(appCodeCharset, ch) {
			t.Fatalf("suffix contains invalid char %q in %q", ch, suffix)
		}
	}
}

func TestGenerateApplicationCode_ContainsToday(t *testing.T) {
	code := GenerateApplicationCode()
	today := time.Now().Format("20060102")
	if !strings.Contains(code, today) {
		t.Fatalf("expected code to contain today %q, got %q", today, code)
	}
}

func TestGenerateApplicationCodeAt_Unique(t *testing.T) {
	ts := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	seen := make(map[string]bool)
	for i := 0; i < 200; i++ {
		c := GenerateApplicationCodeAt(ts)
		if seen[c] {
			t.Fatalf("duplicate code generated: %q", c)
		}
		seen[c] = true
	}
}
