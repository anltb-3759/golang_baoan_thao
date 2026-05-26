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

func newMockApplicationRepo(t *testing.T) (ApplicationRepository, sqlmock.Sqlmock, func()) {
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

func TestApplicationRepoGetByIDPreloadsCitizenUser(t *testing.T) {
	repo, mock, cleanup := newMockApplicationRepo(t)
	defer cleanup()

	appRows := sqlmock.NewRows([]string{"id", "application_code", "citizen_user_id", "service_type_id", "assigned_staff_user_id"}).
		AddRow("app1", "APP-1", "citizen-1", "service-1", nil)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "applications" WHERE id = $1 AND deleted_at IS NULL ORDER BY "applications"."id" LIMIT $2`)).
		WithArgs("app1", 1).
		WillReturnRows(appRows)

	attachRows := sqlmock.NewRows([]string{"id", "application_id"})
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "application_attachments" WHERE "application_attachments"."application_id" = $1`)).
		WithArgs("app1").
		WillReturnRows(attachRows)

	serviceRows := sqlmock.NewRows([]string{"id", "name"}).AddRow("service-1", "Service A")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "service_types" WHERE "service_types"."id" = $1`)).
		WithArgs("service-1").
		WillReturnRows(serviceRows)

	citizenRows := sqlmock.NewRows([]string{"id", "name", "email"}).AddRow("citizen-1", "Citizen A", "citizen@example.com")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1`)).
		WithArgs("citizen-1").
		WillReturnRows(citizenRows)

	app, err := repo.GetByID("app1")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if app == nil {
		t.Fatal("expected app, got nil")
	}
	if app.CitizenUser.Name != "Citizen A" {
		t.Fatalf("expected citizen name %q, got %q", "Citizen A", app.CitizenUser.Name)
	}
}

func TestApplicationRepoListStatusLogsByCitizenReturnsItems(t *testing.T) {
	repo, mock, cleanup := newMockApplicationRepo(t)
	defer cleanup()

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(`SELECT count\(\*\)`).WillReturnRows(countRows)

	logRows := sqlmock.NewRows([]string{"id", "application_id", "new_status", "created_at"}).
		AddRow("log1", "app1", "processing", time.Now())
	mock.ExpectQuery(`SELECT`).WillReturnRows(logRows)

	userRows := sqlmock.NewRows([]string{"id", "name"}).AddRow("staff1", "Staff A")
	mock.ExpectQuery(`SELECT`).WillReturnRows(userRows)

	items, total, err := repo.ListStatusLogsByCitizen("app1", "u1", 1, 10, nil)
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

func TestApplicationRepoCreateAttachmentsSuccess(t *testing.T) {
	repo, mock, cleanup := newMockApplicationRepo(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "application_attachments"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("att1"))
	mock.ExpectCommit()

	err := repo.CreateAttachments("app1", []models.ApplicationAttachment{{FileName: "a.pdf", AttachmentType: models.AttachmentTypeSupplement}})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestApplicationRepoAdminList_WithAssignedStaffFilter(t *testing.T) {
	repo, mock, cleanup := newMockApplicationRepo(t)
	defer cleanup()

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(`SELECT count\(\*\).*assigned_staff_user_id`).
		WithArgs("staff-1").
		WillReturnRows(countRows)

	appRows := sqlmock.NewRows([]string{"id", "application_code", "citizen_user_id", "service_type_id", "assigned_staff_user_id"}).
		AddRow("app1", "APP-1", "citizen-1", "service-1", "staff-1")
	mock.ExpectQuery(`SELECT .*assigned_staff_user_id`).
		WithArgs("staff-1", 10, 0).
		WillReturnRows(appRows)

	serviceRows := sqlmock.NewRows([]string{"id", "name"}).AddRow("service-1", "Service A")
	mock.ExpectQuery(`SELECT`).WillReturnRows(serviceRows)
	citizenRows := sqlmock.NewRows([]string{"id", "name"}).AddRow("citizen-1", "Citizen A")
	mock.ExpectQuery(`SELECT`).WillReturnRows(citizenRows)
	staffRows := sqlmock.NewRows([]string{"id", "name"}).AddRow("staff-1", "Staff A")
	mock.ExpectQuery(`SELECT`).WillReturnRows(staffRows)

	items, total, err := repo.AdminList(ApplicationFilter{AssignedStaffUserID: "staff-1"}, 1, 10)
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
