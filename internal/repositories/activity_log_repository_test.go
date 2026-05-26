package repositories

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
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

func TestActivityLogRepo_Create_OK(t *testing.T) {
	repo, mock, cleanup := newMockActivityLogRepo(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "activity_logs"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("log-new"))
	mock.ExpectCommit()

	log := &models.ActivityLog{Action: "auth.login", Result: "success"}
	if err := repo.Create(log); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestActivityLogRepo_Create_Error(t *testing.T) {
	repo, mock, cleanup := newMockActivityLogRepo(t)
	defer cleanup()

	dbErr := fmt.Errorf("insert error")
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "activity_logs"`)).WillReturnError(dbErr)
	mock.ExpectRollback()

	if err := repo.Create(&models.ActivityLog{Action: "test"}); err == nil {
		t.Fatal("expected error, got nil")
	}
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

func TestActivityLogRepo_List_NegativeOffset(t *testing.T) {
	repo, mock, cleanup := newMockActivityLogRepo(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "activity_logs"`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "activity_logs" ORDER BY created_at DESC LIMIT $1`)).
		WithArgs(10).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	_, _, err := repo.List(ActivityLogFilter{}, -1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestActivityLogRepo_List_ZeroLimit(t *testing.T) {
	repo, mock, cleanup := newMockActivityLogRepo(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "activity_logs"`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "activity_logs" ORDER BY created_at DESC LIMIT $1`)).
		WithArgs(defaultActivityLogListLimit).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	_, _, err := repo.List(ActivityLogFilter{}, 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestActivityLogRepo_List_MultipleFilters(t *testing.T) {
	repo, mock, cleanup := newMockActivityLogRepo(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "activity_logs" WHERE action = $1 AND entity_type = $2 AND result = $3 AND actor_user_id = $4`)).
		WithArgs("login", "user", "success", "u1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	rows := sqlmock.NewRows([]string{"id", "action"}).AddRow("log-1", "login")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "activity_logs" WHERE action = $1 AND entity_type = $2 AND result = $3 AND actor_user_id = $4 ORDER BY created_at DESC LIMIT $5`)).
		WithArgs("login", "user", "success", "u1", 10).
		WillReturnRows(rows)

	filter := ActivityLogFilter{Action: "login", EntityType: "user", Result: "success", ActorUserID: "u1"}
	logs, total, err := repo.List(filter, 0, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || len(logs) != 1 {
		t.Fatalf("expected 1 log, got total=%d len=%d", total, len(logs))
	}
}

func TestActivityLogRepo_List_CountError(t *testing.T) {
	repo, mock, cleanup := newMockActivityLogRepo(t)
	defer cleanup()

	dbErr := fmt.Errorf("count error")
	mock.ExpectQuery(`SELECT count\(\*\) FROM "activity_logs"`).WillReturnError(dbErr)

	_, _, err := repo.List(ActivityLogFilter{}, 0, 10)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestActivityLogRepo_List_FindError(t *testing.T) {
	repo, mock, cleanup := newMockActivityLogRepo(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT count\(\*\) FROM "activity_logs"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	dbErr := fmt.Errorf("find error")
	mock.ExpectQuery(`SELECT \* FROM "activity_logs"`).WillReturnError(dbErr)

	_, _, err := repo.List(ActivityLogFilter{}, 0, 10)
	if err == nil {
		t.Fatal("expected error, got nil")
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

func TestActivityLogRepo_DeleteBefore_Error(t *testing.T) {
	repo, mock, cleanup := newMockActivityLogRepo(t)
	defer cleanup()

	before := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	dbErr := fmt.Errorf("delete failed")

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "activity_logs" WHERE created_at < $1`)).
		WithArgs(before).
		WillReturnError(dbErr)
	mock.ExpectRollback()

	_, err := repo.DeleteBefore(before)
	if err == nil {
		t.Fatal("expected error, got nil")
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
