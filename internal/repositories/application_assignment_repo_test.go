package repositories

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func newMockApplicationAssignmentRepo(t *testing.T) (ApplicationAssignmentRepository, sqlmock.Sqlmock, func()) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn:                 sqlDB,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("open gorm db: %v", err)
	}
	return NewApplicationAssignmentRepo(db), mock, func() { _ = sqlDB.Close() }
}

func TestApplicationAssignmentRepo_Create(t *testing.T) {
	repo, mock, cleanup := newMockApplicationAssignmentRepo(t)
	defer cleanup()

	assignment := &models.ApplicationAssignment{
		ID:            "assign-1",
		ApplicationID: "app-1",
		Action:        models.AssignmentActionAssigned,
		CreatedAt:     time.Now(),
	}
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "application_assignments"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("assign-1"))
	mock.ExpectCommit()

	result, err := repo.Create(assignment)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || result.ID != "assign-1" {
		t.Fatalf("expected assignment assign-1, got %v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestApplicationAssignmentRepo_Create_Error(t *testing.T) {
	repo, mock, cleanup := newMockApplicationAssignmentRepo(t)
	defer cleanup()

	assignment := &models.ApplicationAssignment{
		ID:            "assign-fail",
		ApplicationID: "app-1",
		Action:        models.AssignmentActionAssigned,
		CreatedAt:     time.Now(),
	}
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "application_assignments"`)).
		WillReturnError(errors.New("insert error"))
	mock.ExpectRollback()

	result, err := repo.Create(assignment)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if result != nil {
		t.Fatal("expected nil result on error")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestApplicationAssignmentRepo_ListByApplication(t *testing.T) {
	repo, mock, cleanup := newMockApplicationAssignmentRepo(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "application_id", "action", "created_at"}).
		AddRow("assign-1", "app-1", "assigned", time.Now()).
		AddRow("assign-2", "app-1", "transferred", time.Now())

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "application_assignments" WHERE application_id = $1 ORDER BY created_at DESC`)).
		WithArgs("app-1").
		WillReturnRows(rows)

	items, err := repo.ListByApplication("app-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestApplicationAssignmentRepo_ListByApplication_Empty(t *testing.T) {
	repo, mock, cleanup := newMockApplicationAssignmentRepo(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "application_assignments" WHERE application_id = $1 ORDER BY created_at DESC`)).
		WithArgs("app-empty").
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	items, err := repo.ListByApplication("app-empty")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(items))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestApplicationAssignmentRepo_ListByApplication_Error(t *testing.T) {
	repo, mock, cleanup := newMockApplicationAssignmentRepo(t)
	defer cleanup()

	dbErr := errors.New("db error")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "application_assignments" WHERE application_id = $1 ORDER BY created_at DESC`)).
		WithArgs("app-err").
		WillReturnError(dbErr)

	_, err := repo.ListByApplication("app-err")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestApplicationAssignmentRepo_ListByApplication_RecordNotFound(t *testing.T) {
	repo, mock, cleanup := newMockApplicationAssignmentRepo(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "application_assignments" WHERE application_id = $1 ORDER BY created_at DESC`)).
		WithArgs("app-nf").
		WillReturnError(gorm.ErrRecordNotFound)

	items, err := repo.ListByApplication("app-nf")
	if err != nil {
		t.Fatalf("expected nil error for RecordNotFound, got %v", err)
	}
	if items != nil {
		t.Fatalf("expected nil items, got %v", items)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
