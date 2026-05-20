package repositories

import (
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func newMockApplicationRepo(t *testing.T) (ApplicationRepository, sqlmock.Sqlmock, func()) {
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
	return NewApplicationRepository(db), mock, func() { _ = sqlDB.Close() }
}

func TestApplicationRepoListByCitizenReturnsItems(t *testing.T) {
	repo, mock, cleanup := newMockApplicationRepo(t)
	defer cleanup()

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "applications" WHERE citizen_user_id = $1 AND deleted_at IS NULL`)).
		WithArgs("u1").
		WillReturnRows(countRows)

	appRows := sqlmock.NewRows([]string{"id", "application_code", "citizen_user_id"}).
		AddRow("app1", "APP-20240101-ABCDEF", "u1")
	mock.ExpectQuery(`SELECT`).WillReturnRows(appRows)

	// Preload ServiceType
	stRows := sqlmock.NewRows([]string{"id", "name"})
	mock.ExpectQuery(`SELECT`).WillReturnRows(stRows)

	items, total, err := repo.ListByCitizen("u1", 1, 10)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if total != 1 {
		t.Fatalf("expected total=1, got %d", total)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
}

func TestApplicationRepoListByCitizenCountError(t *testing.T) {
	repo, mock, cleanup := newMockApplicationRepo(t)
	defer cleanup()

	dbErr := errors.New("count error")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "applications" WHERE citizen_user_id = $1 AND deleted_at IS NULL`)).
		WithArgs("u1").
		WillReturnError(dbErr)

	_, _, err := repo.ListByCitizen("u1", 1, 10)
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected count error, got %v", err)
	}
}

func TestApplicationRepoListByCitizenFindError(t *testing.T) {
	repo, mock, cleanup := newMockApplicationRepo(t)
	defer cleanup()

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "applications" WHERE citizen_user_id = $1 AND deleted_at IS NULL`)).
		WithArgs("u1").
		WillReturnRows(countRows)

	dbErr := errors.New("find error")
	mock.ExpectQuery(`SELECT`).WillReturnError(dbErr)

	_, _, err := repo.ListByCitizen("u1", 1, 10)
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected find error, got %v", err)
	}
}

func TestApplicationRepoGetByIDForCitizenReturnsApp(t *testing.T) {
	repo, mock, cleanup := newMockApplicationRepo(t)
	defer cleanup()

	appRows := sqlmock.NewRows([]string{"id", "citizen_user_id"}).
		AddRow("app1", "u1")
	mock.ExpectQuery(`SELECT`).WillReturnRows(appRows)

	stRows := sqlmock.NewRows([]string{"id", "name"})
	mock.ExpectQuery(`SELECT`).WillReturnRows(stRows)

	app, err := repo.GetByIDForCitizen("app1", "u1")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if app == nil || app.ID != "app1" {
		t.Fatalf("expected app1, got %#v", app)
	}
}

func TestApplicationRepoGetByIDForCitizenNotFound(t *testing.T) {
	repo, mock, cleanup := newMockApplicationRepo(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT`).WillReturnError(gorm.ErrRecordNotFound)

	_, err := repo.GetByIDForCitizen("bad-id", "u1")
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}
