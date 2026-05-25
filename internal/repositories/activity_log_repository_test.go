package repositories

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func newMockActivityLogRepo(t *testing.T) (ActivityLogRepository, sqlmock.Sqlmock, func()) {
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
	return NewActivityLogRepository(db), mock, func() { _ = sqlDB.Close() }
}

func TestActivityLogRepo_List_FilterByAction(t *testing.T) {
	repo, mock, cleanup := newMockActivityLogRepo(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "activity_logs" WHERE action = $1`)).
		WithArgs("auth.login").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	rows := sqlmock.NewRows([]string{"id", "action", "created_at"}).
		AddRow("log-1", "auth.login", time.Now())
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "activity_logs" WHERE action = $1 ORDER BY created_at DESC LIMIT $2`)).
		WithArgs("auth.login", 10).
		WillReturnRows(rows)

	logs, total, err := repo.List(ActivityLogFilter{Action: "auth.login"}, 0, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected total 1, got %d", total)
	}
	if len(logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(logs))
	}
	if logs[0].Action != "auth.login" {
		t.Fatalf("expected action auth.login, got %q", logs[0].Action)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestActivityLogRepo_DeleteBefore(t *testing.T) {
	repo, mock, cleanup := newMockActivityLogRepo(t)
	defer cleanup()

	before := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "activity_logs" WHERE created_at < $1`)).
		WithArgs(before).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectCommit()

	deleted, err := repo.DeleteBefore(before)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deleted != 2 {
		t.Fatalf("expected deleted 2, got %d", deleted)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestActivityLogRepo_DeleteRetainDays(t *testing.T) {
	repo, mock, cleanup := newMockActivityLogRepo(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "activity_logs" WHERE created_at < $1`)).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 3))
	mock.ExpectCommit()

	deleted, err := repo.DeleteRetainDays(7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deleted != 3 {
		t.Fatalf("expected deleted 3, got %d", deleted)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestActivityLogRepo_DeleteRetainDays_InvalidDays(t *testing.T) {
	repo, _, cleanup := newMockActivityLogRepo(t)
	defer cleanup()

	deleted, err := repo.DeleteRetainDays(0)
	if err == nil {
		t.Fatal("expected error for non-positive retain days")
	}
	if deleted != 0 {
		t.Fatalf("expected deleted 0, got %d", deleted)
	}
}
