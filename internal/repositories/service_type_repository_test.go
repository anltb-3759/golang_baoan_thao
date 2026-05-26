package repositories

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func newMockServiceTypeRepo(t *testing.T) (ServiceTypeRepository, sqlmock.Sqlmock, func()) {
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
	return NewServiceTypeRepository(db), mock, func() { _ = sqlDB.Close() }
}

func TestServiceTypeRepoListReturnsItems(t *testing.T) {
	repo, mock, cleanup := newMockServiceTypeRepo(t)
	defer cleanup()

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "service_types" WHERE is_active = $1 AND deleted_at IS NULL`)).
		WithArgs(true).
		WillReturnRows(countRows)

	stRows := sqlmock.NewRows([]string{"id", "name", "code"}).
		AddRow("st1", "Cấp CCCD", "CCCD_NEW")
	mock.ExpectQuery(`SELECT`).WillReturnRows(stRows)

	// Preload ResponsibleDepartment
	deptRows := sqlmock.NewRows([]string{"id", "name"})
	mock.ExpectQuery(`SELECT`).WillReturnRows(deptRows)

	result, err := repo.List(context.Background(), ListFilter{Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if result.Total != 1 {
		t.Fatalf("expected total=1, got %d", result.Total)
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result.Items))
	}
}

func TestServiceTypeRepoListWithCategoryAndSearch(t *testing.T) {
	repo, mock, cleanup := newMockServiceTypeRepo(t)
	defer cleanup()

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery(`SELECT count\(\*\) FROM "service_types"`).
		WillReturnRows(countRows)

	stRows := sqlmock.NewRows([]string{"id", "name", "code"})
	mock.ExpectQuery(`SELECT`).WillReturnRows(stRows)

	deptRows := sqlmock.NewRows([]string{"id", "name"})
	mock.ExpectQuery(`SELECT`).WillReturnRows(deptRows)

	result, err := repo.List(context.Background(), ListFilter{Category: "hanh_chinh", Search: "CCCD", Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if result.Total != 0 {
		t.Fatalf("expected total=0, got %d", result.Total)
	}
}

func TestServiceTypeRepoListCountError(t *testing.T) {
	repo, mock, cleanup := newMockServiceTypeRepo(t)
	defer cleanup()

	dbErr := errors.New("count error")
	mock.ExpectQuery(`SELECT count`).WillReturnError(dbErr)

	_, err := repo.List(context.Background(), ListFilter{Page: 1, Limit: 10})
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected count error, got %v", err)
	}
}

func TestServiceTypeRepoListFindError(t *testing.T) {
	repo, mock, cleanup := newMockServiceTypeRepo(t)
	defer cleanup()

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery(`SELECT count`).WillReturnRows(countRows)

	dbErr := errors.New("find error")
	mock.ExpectQuery(`SELECT`).WillReturnError(dbErr)

	_, err := repo.List(context.Background(), ListFilter{Page: 1, Limit: 10})
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected find error, got %v", err)
	}
}

func TestServiceTypeRepoGetByIDReturnsServiceType(t *testing.T) {
	repo, mock, cleanup := newMockServiceTypeRepo(t)
	defer cleanup()

	stRows := sqlmock.NewRows([]string{"id", "name", "is_active"}).
		AddRow("st1", "Cấp CCCD", true)
	mock.ExpectQuery(`SELECT`).WillReturnRows(stRows)

	deptRows := sqlmock.NewRows([]string{"id", "name"})
	mock.ExpectQuery(`SELECT`).WillReturnRows(deptRows)

	st, err := repo.GetByID(context.Background(), "st1")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if st == nil {
		t.Fatal("expected service type, got nil")
	}
}

func TestServiceTypeRepoGetByIDReturnsNotFound(t *testing.T) {
	repo, mock, cleanup := newMockServiceTypeRepo(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT`).WillReturnError(gorm.ErrRecordNotFound)

	_, err := repo.GetByID(context.Background(), "bad-id")
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}

func TestServiceTypeRepoGetByIDForAdminReturnsInactiveServiceType(t *testing.T) {
	repo, mock, cleanup := newMockServiceTypeRepo(t)
	defer cleanup()

	stRows := sqlmock.NewRows([]string{"id", "name", "is_active"}).AddRow("st1", "Cấp CCCD", false)
	mock.ExpectQuery(`SELECT`).WillReturnRows(stRows)

	deptRows := sqlmock.NewRows([]string{"id", "name"})
	mock.ExpectQuery(`SELECT`).WillReturnRows(deptRows)

	st, err := repo.GetByIDForAdmin(context.Background(), "st1")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if st == nil {
		t.Fatal("expected service type, got nil")
	}
	if st.IsActive {
		t.Fatal("expected inactive service type")
	}
}

func TestServiceTypeRepoCreateUpdateDeleteCountApplications(t *testing.T) {
	repo, mock, cleanup := newMockServiceTypeRepo(t)
	defer cleanup()

	serviceType := &models.ServiceType{ID: "st1", Name: "Test", Code: "TEST", FormSchema: []byte("{}"), CreatedAt: time.Now(), UpdatedAt: time.Now()}
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "service_types"`).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("st1"))
	mock.ExpectCommit()

	if err := repo.Create(context.Background(), serviceType); err != nil {
		t.Fatalf("expected nil error on create, got %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "service_types"`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	if err := repo.Update(context.Background(), serviceType); err != nil {
		t.Fatalf("expected nil error on update, got %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "service_types"`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	if err := repo.Delete(context.Background(), "st1"); err != nil {
		t.Fatalf("expected nil error on delete, got %v", err)
	}

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(3)
	mock.ExpectQuery(`SELECT count\(\*\) FROM "applications"`).WillReturnRows(countRows)
	count, err := repo.CountApplications(context.Background(), "st1")
	if err != nil {
		t.Fatalf("expected nil error on count, got %v", err)
	}
	if count != 3 {
		t.Fatalf("expected count=3, got %d", count)
	}
}

func TestServiceTypeRepoListDepartments(t *testing.T) {
	repo, mock, cleanup := newMockServiceTypeRepo(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow("d1", "IT").AddRow("d2", "HR")
	mock.ExpectQuery(`SELECT \* FROM "departments"`).WillReturnRows(rows)

	depts, err := repo.ListDepartments(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(depts) != 2 {
		t.Fatalf("expected 2 departments, got %d", len(depts))
	}
}

func TestServiceTypeRepoListDepartmentsError(t *testing.T) {
	repo, mock, cleanup := newMockServiceTypeRepo(t)
	defer cleanup()

	dbErr := errors.New("db error")
	mock.ExpectQuery(`SELECT \* FROM "departments"`).WillReturnError(dbErr)

	_, err := repo.ListDepartments(context.Background())
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected db error, got %v", err)
	}
}

func TestServiceTypeRepoListCategories(t *testing.T) {
	repo, mock, cleanup := newMockServiceTypeRepo(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "name"}).AddRow("c1", "Hành chính")
	mock.ExpectQuery(`SELECT \* FROM "categories"`).WillReturnRows(rows)

	cats, err := repo.ListCategories(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cats) != 1 {
		t.Fatalf("expected 1 category, got %d", len(cats))
	}
}

func TestServiceTypeRepoListCategoriesError(t *testing.T) {
	repo, mock, cleanup := newMockServiceTypeRepo(t)
	defer cleanup()

	dbErr := errors.New("db error")
	mock.ExpectQuery(`SELECT \* FROM "categories"`).WillReturnError(dbErr)

	_, err := repo.ListCategories(context.Background())
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected db error, got %v", err)
	}
}

func TestServiceTypeRepoCountApplicationsError(t *testing.T) {
	repo, mock, cleanup := newMockServiceTypeRepo(t)
	defer cleanup()

	dbErr := errors.New("count error")
	mock.ExpectQuery(`SELECT count\(\*\) FROM "applications"`).WillReturnError(dbErr)

	_, err := repo.CountApplications(context.Background(), "st1")
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected count error, got %v", err)
	}
}

func TestServiceTypeRepoCreateInTx(t *testing.T) {
	_, _, cleanup := newMockServiceTypeRepo(t)
	defer cleanup()

	sqlDB2, innerMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create inner sqlmock: %v", err)
	}
	defer sqlDB2.Close()
	innerDB, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB2, PreferSimpleProtocol: true}), &gorm.Config{})
	if err != nil {
		t.Fatalf("open inner gorm: %v", err)
	}

	st := &models.ServiceType{ID: "st1", Name: "Test", Code: "TST", FormSchema: []byte("{}"), CreatedAt: time.Now(), UpdatedAt: time.Now()}
	innerMock.ExpectBegin()
	innerMock.ExpectQuery(`INSERT INTO "service_types"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("st1"))
	innerMock.ExpectCommit()

	repo := NewServiceTypeRepository(innerDB)
	tx := innerDB.Begin()
	if err := repo.CreateInTx(tx, st); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = tx.Commit()
	if err := innerMock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// keep unused import happy
var _ = models.ServiceType{}
