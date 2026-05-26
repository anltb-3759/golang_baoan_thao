package dtos

import (
	"testing"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/models"
)

func TestNewCitizenProfileResponse(t *testing.T) {
	dob := time.Date(1995, 1, 15, 0, 0, 0, 0, time.UTC)
	user := &models.User{
		ID:      "u1",
		Name:    "An",
		Email:   "an@example.com",
		Phone:   "0901234567",
		Address: "Ha Noi",
	}
	profile := &models.CitizenProfile{
		CitizenIDNumber:          "123456789012",
		DateOfBirth:              &dob,
		Gender:                   "male",
		PermanentAddress:         "HN",
		EmailNotificationEnabled: true,
		CreatedAt:                time.Now(),
		UpdatedAt:                time.Now(),
	}

	resp := NewCitizenProfileResponse(user, profile)

	if resp.UserID != "u1" {
		t.Fatalf("expected UserID u1, got %q", resp.UserID)
	}
	if resp.Name != "An" {
		t.Fatalf("expected Name An, got %q", resp.Name)
	}
	if resp.Email != "an@example.com" {
		t.Fatalf("expected email, got %q", resp.Email)
	}
	if resp.CitizenIDNumber != "123456789012" {
		t.Fatalf("expected CitizenIDNumber, got %q", resp.CitizenIDNumber)
	}
	if !resp.EmailNotificationEnabled {
		t.Fatal("expected EmailNotificationEnabled true")
	}
}

func TestNewNotificationResponse(t *testing.T) {
	appID := "app-1"
	readAt := time.Now()
	n := &models.Notification{
		ID:            "n1",
		ApplicationID: &appID,
		Title:         "Test",
		Message:       "Hello",
		Type:          models.NotificationTypeSystem,
		IsRead:        true,
		ReadAt:        &readAt,
		CreatedAt:     time.Now(),
	}

	resp := NewNotificationResponse(n)

	if resp.ID != "n1" {
		t.Fatalf("expected ID n1, got %q", resp.ID)
	}
	if resp.ApplicationID == nil || *resp.ApplicationID != "app-1" {
		t.Fatalf("expected ApplicationID app-1")
	}
	if resp.Title != "Test" {
		t.Fatalf("expected Title Test, got %q", resp.Title)
	}
	if resp.Type != "system" {
		t.Fatalf("expected type system, got %q", resp.Type)
	}
	if !resp.IsRead {
		t.Fatal("expected IsRead true")
	}
	if resp.ReadAt == nil {
		t.Fatal("expected ReadAt non-nil")
	}
}
