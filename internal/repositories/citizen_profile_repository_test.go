package repositories

import (
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func newMockCitizenProfileRepo(t *testing.T) (CitizenProfileRepository, sqlmock.Sqlmock, func()) {
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
	return NewCitizenProfileRepository(db), mock, func() { _ = sqlDB.Close() }
}

func TestCitizenProfileRepoGetByUserIDReturnsProfile(t *testing.T) {
	repo, mock, cleanup := newMockCitizenProfileRepo(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "user_id"}).
		AddRow("p1", "u1")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "citizen_profiles" WHERE user_id = $1 ORDER BY "citizen_profiles"."id" LIMIT $2`)).
		WithArgs("u1", 1).
		WillReturnRows(rows)

	p, err := repo.GetByUserID("u1")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if p == nil || p.UserID != "u1" {
		t.Fatalf("expected profile for u1, got %#v", p)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCitizenProfileRepoGetByUserIDReturnsNilWhenNotFound(t *testing.T) {
	repo, mock, cleanup := newMockCitizenProfileRepo(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "user_id"})
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "citizen_profiles" WHERE user_id = $1 ORDER BY "citizen_profiles"."id" LIMIT $2`)).
		WithArgs("missing", 1).
		WillReturnRows(rows)

	p, err := repo.GetByUserID("missing")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if p != nil {
		t.Fatalf("expected nil profile, got %#v", p)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCitizenProfileRepoGetByUserIDReturnsError(t *testing.T) {
	repo, mock, cleanup := newMockCitizenProfileRepo(t)
	defer cleanup()

	dbErr := errors.New("db error")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "citizen_profiles" WHERE user_id = $1 ORDER BY "citizen_profiles"."id" LIMIT $2`)).
		WithArgs("u1", 1).
		WillReturnError(dbErr)

	_, err := repo.GetByUserID("u1")
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected db error, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCitizenProfileRepoUpdateReturnsSuccess(t *testing.T) {
	repo, mock, cleanup := newMockCitizenProfileRepo(t)
	defer cleanup()

	profile := &models.CitizenProfile{ID: "p1", UserID: "u1"}
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "citizen_profiles"`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := repo.Update(profile); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCitizenProfileRepoUpdateReturnsError(t *testing.T) {
	repo, mock, cleanup := newMockCitizenProfileRepo(t)
	defer cleanup()

	dbErr := errors.New("db error")
	profile := &models.CitizenProfile{ID: "p1", UserID: "u1"}
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "citizen_profiles"`).WillReturnError(dbErr)
	mock.ExpectRollback()

	err := repo.Update(profile)
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected db error, got %v", err)
	}
}

func TestCitizenProfileRepoCreateInTxReturnsSuccess(t *testing.T) {
	repo, mock, cleanup := newMockCitizenProfileRepo(t)
	defer cleanup()

	sqlDB, innerMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create inner sqlmock: %v", err)
	}
	defer sqlDB.Close()
	innerDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn:                 sqlDB,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("open inner gorm db: %v", err)
	}
	_ = mock // outer mock not used for CreateInTx

	profile := &models.CitizenProfile{UserID: "u1"}
	innerMock.ExpectBegin()
	innerMock.ExpectQuery(`INSERT INTO "citizen_profiles"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("p-id"))
	innerMock.ExpectCommit()

	tx := innerDB.Begin()
	if err := repo.CreateInTx(tx, profile); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	_ = tx.Commit()
	if err := innerMock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCitizenProfileRepoFindByCitizenIDNumberFound(t *testing.T) {
	repo, mock, cleanup := newMockCitizenProfileRepo(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "citizen_id_number"}).AddRow("p1", "123456789012")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "citizen_profiles" WHERE citizen_id_number = $1 AND deleted_at IS NULL ORDER BY "citizen_profiles"."id" LIMIT $2`)).
		WithArgs("123456789012", 1).
		WillReturnRows(rows)

	p, err := repo.FindByCitizenIDNumber("123456789012")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p == nil || p.ID != "p1" {
		t.Fatalf("expected profile p1, got %#v", p)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCitizenProfileRepoFindByCitizenIDNumberNotFound(t *testing.T) {
	repo, mock, cleanup := newMockCitizenProfileRepo(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "citizen_profiles" WHERE citizen_id_number = $1 AND deleted_at IS NULL ORDER BY "citizen_profiles"."id" LIMIT $2`)).
		WithArgs("000000000000", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	p, err := repo.FindByCitizenIDNumber("000000000000")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p != nil {
		t.Fatalf("expected nil, got %#v", p)
	}
}

func TestCitizenProfileRepoFindByCitizenIDNumberError(t *testing.T) {
	repo, mock, cleanup := newMockCitizenProfileRepo(t)
	defer cleanup()

	dbErr := errors.New("db error")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "citizen_profiles" WHERE citizen_id_number = $1 AND deleted_at IS NULL ORDER BY "citizen_profiles"."id" LIMIT $2`)).
		WithArgs("bad", 1).
		WillReturnError(dbErr)

	_, err := repo.FindByCitizenIDNumber("bad")
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected db error, got %v", err)
	}
}

func TestCitizenProfileRepoCreateInTxReturnsError(t *testing.T) {
	repo, mock, cleanup := newMockCitizenProfileRepo(t)
	defer cleanup()
	_ = mock

	sqlDB, innerMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create inner sqlmock: %v", err)
	}
	defer sqlDB.Close()
	innerDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn:                 sqlDB,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("open inner gorm db: %v", err)
	}

	dbErr := errors.New("insert error")
	profile := &models.CitizenProfile{UserID: "u1"}
	innerMock.ExpectBegin()
	innerMock.ExpectQuery(`INSERT INTO "citizen_profiles"`).WillReturnError(dbErr)
	innerMock.ExpectRollback()

	tx := innerDB.Begin()
	err = repo.CreateInTx(tx, profile)
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected insert error, got %v", err)
	}
}

func TestCitizenProfileRepoListAllForExport_CountError(t *testing.T) {
	repo, mock, cleanup := newMockCitizenProfileRepo(t)
	defer cleanup()

	dbErr := errors.New("count error")
	mock.ExpectQuery(`SELECT count\(\*\) FROM "citizen_profiles"`).
		WillReturnError(dbErr)

	_, _, err := repo.ListAllForExport(0, 20)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCitizenProfileRepoListAllForExport_ScanError(t *testing.T) {
	repo, mock, cleanup := newMockCitizenProfileRepo(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "citizen_profiles"`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	dbErr := errors.New("scan error")
	mock.ExpectQuery(`citizen_profiles cp`).WillReturnError(dbErr)

	_, _, err := repo.ListAllForExport(0, 20)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCitizenProfileRepoListAllForExport_Success(t *testing.T) {
	repo, mock, cleanup := newMockCitizenProfileRepo(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "citizen_profiles"`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	rows := sqlmock.NewRows([]string{"so_cccd", "ho_ten", "email", "so_dien_thoai", "dia_chi", "ngay_sinh", "tong_ho_so"}).
		AddRow("123456789012", "Nguyen Van A", "a@test.com", "0901234567", "Ha Noi", "01/01/1990", 2)
	mock.ExpectQuery(`citizen_profiles cp`).WillReturnRows(rows)

	result, total, err := repo.ListAllForExport(0, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected total 1, got %d", total)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 row, got %d", len(result))
	}
	if result[0].SoCCCD != "123456789012" {
		t.Fatalf("expected SoCCCD 123456789012, got %q", result[0].SoCCCD)
	}
}
