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

func newMockStaffProfileRepo(t *testing.T) (StaffProfileRepository, sqlmock.Sqlmock, func()) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	mock.MatchExpectationsInOrder(false)
	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn:                 sqlDB,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("open gorm db: %v", err)
	}
	return NewStaffProfileRepo(db), mock, func() { _ = sqlDB.Close() }
}

func TestStaffProfileRepo_FindByUserID_Found(t *testing.T) {
	repo, mock, cleanup := newMockStaffProfileRepo(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "user_id", "position", "created_at", "updated_at"}).
		AddRow("sp-1", "user-1", "Engineer", time.Now(), time.Now())
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "staff_profiles" WHERE user_id = $1 AND deleted_at IS NULL ORDER BY "staff_profiles"."id" LIMIT $2`)).
		WithArgs("user-1", 1).
		WillReturnRows(rows)

	// Preload queries for User and Department (order depends on GORM internals)
	mock.ExpectQuery(`SELECT \* FROM "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery(`SELECT \* FROM "departments"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	profile, err := repo.FindByUserID("user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile == nil || profile.ID != "sp-1" {
		t.Fatalf("expected profile sp-1, got %v", profile)
	}
}

func TestStaffProfileRepo_FindByUserID_NotFound(t *testing.T) {
	repo, mock, cleanup := newMockStaffProfileRepo(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "staff_profiles" WHERE user_id = $1 AND deleted_at IS NULL ORDER BY "staff_profiles"."id" LIMIT $2`)).
		WithArgs("missing", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	profile, err := repo.FindByUserID("missing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile != nil {
		t.Fatal("expected nil for not found")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestStaffProfileRepo_FindByUserID_Error(t *testing.T) {
	repo, mock, cleanup := newMockStaffProfileRepo(t)
	defer cleanup()

	dbErr := errors.New("db error")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "staff_profiles" WHERE user_id = $1 AND deleted_at IS NULL ORDER BY "staff_profiles"."id" LIMIT $2`)).
		WithArgs("user-err", 1).
		WillReturnError(dbErr)

	_, err := repo.FindByUserID("user-err")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestStaffProfileRepo_ListByDepartment_WithDeptID(t *testing.T) {
	repo, mock, cleanup := newMockStaffProfileRepo(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "staff_profiles" WHERE deleted_at IS NULL AND department_id = $1`)).
		WithArgs("dept-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	rows := sqlmock.NewRows([]string{"id", "user_id", "department_id", "created_at", "updated_at"}).
		AddRow("sp-1", "user-1", "dept-1", time.Now(), time.Now())
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "staff_profiles" WHERE deleted_at IS NULL AND department_id = $1 ORDER BY created_at DESC LIMIT $2`)).
		WithArgs("dept-1", 10).
		WillReturnRows(rows)

	// Preload queries (departments before users for ListByDepartment)
	mock.ExpectQuery(`SELECT \* FROM "departments"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery(`SELECT \* FROM "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	items, total, err := repo.ListByDepartment("dept-1", 0, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected total 1, got %d", total)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
}

func TestStaffProfileRepo_ListByDepartment_NoDeptID(t *testing.T) {
	repo, mock, cleanup := newMockStaffProfileRepo(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "staff_profiles" WHERE deleted_at IS NULL AND department_id IS NULL`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	rows := sqlmock.NewRows([]string{"id", "user_id", "created_at", "updated_at"}).
		AddRow("sp-1", "user-1", time.Now(), time.Now()).
		AddRow("sp-2", "user-2", time.Now(), time.Now())
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "staff_profiles" WHERE deleted_at IS NULL AND department_id IS NULL ORDER BY created_at DESC LIMIT $1`)).
		WithArgs(10).
		WillReturnRows(rows)

	mock.ExpectQuery(`SELECT \* FROM "departments"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery(`SELECT \* FROM "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	items, total, err := repo.ListByDepartment("", 0, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected total 2, got %d", total)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
}

func TestStaffProfileRepo_UpdateDepartment(t *testing.T) {
	repo, mock, cleanup := newMockStaffProfileRepo(t)
	defer cleanup()

	deptID := "dept-1"
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "staff_profiles" SET`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "user-1").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.UpdateDepartment("user-1", &deptID, "admin-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestStaffProfileRepo_UpdateDepartment_Nil(t *testing.T) {
	repo, mock, cleanup := newMockStaffProfileRepo(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "staff_profiles" SET`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "user-1").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.UpdateDepartment("user-1", nil, "admin-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestStaffProfileRepo_Create(t *testing.T) {
	repo, mock, cleanup := newMockStaffProfileRepo(t)
	defer cleanup()

	profile := &models.StaffProfile{
		ID:        "sp-new",
		UserID:    "user-1",
		Position:  "Staff",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "staff_profiles"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("sp-new"))
	mock.ExpectCommit()

	result, err := repo.Create(profile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || result.ID != "sp-new" {
		t.Fatalf("expected profile sp-new, got %v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestStaffProfileRepo_Create_Error(t *testing.T) {
	repo, mock, cleanup := newMockStaffProfileRepo(t)
	defer cleanup()

	profile := &models.StaffProfile{
		ID:        "sp-fail",
		UserID:    "user-fail",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "staff_profiles"`)).
		WillReturnError(errors.New("insert error"))
	mock.ExpectRollback()

	result, err := repo.Create(profile)
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

func TestStaffProfileRepo_CreateInTx(t *testing.T) {
	_, mock, cleanup := newMockStaffProfileRepo(t)
	defer cleanup()

	sqlDB2, mock2, _ := sqlmock.New()
	db2, _ := gorm.Open(postgres.New(postgres.Config{
		Conn:                 sqlDB2,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	repo2 := NewStaffProfileRepo(db2)

	profile := &models.StaffProfile{
		ID:        "sp-tx",
		UserID:    "user-tx",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	mock2.ExpectBegin()
	mock2.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "staff_profiles"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("sp-tx"))
	mock2.ExpectCommit()

	tx := db2.Begin()
	err := repo2.(*staffProfileRepo).CreateInTx(tx, profile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tx.Commit()

	_ = mock
	if err := mock2.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestStaffProfileRepo_ListByDepartment_CountError(t *testing.T) {
	repo, mock, cleanup := newMockStaffProfileRepo(t)
	defer cleanup()

	dbErr := errors.New("count error")
	mock.ExpectQuery(`SELECT count\(\*\) FROM "staff_profiles"`).WillReturnError(dbErr)

	_, _, err := repo.ListByDepartment("dept-1", 0, 10)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestStaffProfileRepo_ListByDepartment_FindError(t *testing.T) {
	repo, mock, cleanup := newMockStaffProfileRepo(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT count\(\*\) FROM "staff_profiles"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	dbErr := errors.New("find error")
	mock.ExpectQuery(`SELECT \* FROM "staff_profiles"`).WillReturnError(dbErr)

	_, _, err := repo.ListByDepartment("dept-1", 0, 10)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
