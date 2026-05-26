package models

import (
	"testing"
	"time"
)

func TestSubmittedAtFormatted_NonZero(t *testing.T) {
	ts := time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC)
	a := Application{SubmittedAt: ts}
	got := a.SubmittedAtFormatted()
	want := "2024-03-15 10:30"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestSubmittedAtFormatted_Zero(t *testing.T) {
	a := Application{}
	got := a.SubmittedAtFormatted()
	if got != "" {
		t.Fatalf("expected empty string for zero time, got %q", got)
	}
}
