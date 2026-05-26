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

func newMockUserRepo(t *testing.T) (UserRepository, sqlmock.Sqlmock, func()) {
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

	cleanup := func() {
		_ = sqlDB.Close()
	}

	return NewUserRepo(db), mock, cleanup
}

func TestUserRepoFindByEmailReturnsUser(t *testing.T) {
	repo, mock, cleanup := newMockUserRepo(t)
	defer cleanup()

	email := "user@example.com"
	rows := sqlmock.NewRows([]string{"id", "name", "email", "password_hash", "role", "status"}).
		AddRow("user-id", "User", email, "hashed-password", models.UserRoleCitizen, models.UserStatusActive)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE email = $1 AND deleted_at IS NULL ORDER BY "users"."id" LIMIT $2`)).
		WithArgs(email, 1).
		WillReturnRows(rows)

	user, err := repo.FindByEmail(email)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if user == nil {
		t.Fatal("expected user, got nil")
	}
	if user.Email != email {
		t.Fatalf("expected email %q, got %q", email, user.Email)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestUserRepoFindByEmailReturnsNilWhenNotFound(t *testing.T) {
	repo, mock, cleanup := newMockUserRepo(t)
	defer cleanup()

	email := "missing@example.com"
	rows := sqlmock.NewRows([]string{"id", "name", "email", "password_hash", "role", "status"})

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE email = $1 AND deleted_at IS NULL ORDER BY "users"."id" LIMIT $2`)).
		WithArgs(email, 1).
		WillReturnRows(rows)

	user, err := repo.FindByEmail(email)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if user != nil {
		t.Fatalf("expected nil user, got %#v", user)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestUserRepoFindByEmailReturnsDatabaseError(t *testing.T) {
	repo, mock, cleanup := newMockUserRepo(t)
	defer cleanup()

	email := "user@example.com"
	dbErr := errors.New("database error")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE email = $1 AND deleted_at IS NULL ORDER BY "users"."id" LIMIT $2`)).
		WithArgs(email, 1).
		WillReturnError(dbErr)

	user, err := repo.FindByEmail(email)
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected database error, got %v", err)
	}
	if user != nil {
		t.Fatalf("expected nil user, got %#v", user)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestUserRepoCreateReturnsCreatedUser(t *testing.T) {
	repo, mock, cleanup := newMockUserRepo(t)
	defer cleanup()

	user := &models.User{
		Name:         "User",
		Email:        "user@example.com",
		PasswordHash: "hashed-password",
		Role:         models.UserRoleCitizen,
		Status:       models.UserStatusActive,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("created-user-id"))
	mock.ExpectCommit()

	createdUser, err := repo.Create(user)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if createdUser.ID != "created-user-id" {
		t.Fatalf("expected generated id, got %q", createdUser.ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestUserRepoCreateReturnsDatabaseError(t *testing.T) {
	repo, mock, cleanup := newMockUserRepo(t)
	defer cleanup()

	user := &models.User{
		Name:         "User",
		Email:        "user@example.com",
		PasswordHash: "hashed-password",
		Role:         models.UserRoleCitizen,
		Status:       models.UserStatusActive,
	}
	dbErr := errors.New("database error")

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "users"`).WillReturnError(dbErr)
	mock.ExpectRollback()

	createdUser, err := repo.Create(user)
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected database error, got %v", err)
	}
	if createdUser != nil {
		t.Fatalf("expected nil user, got %#v", createdUser)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestUserRepoFindByIDReturnsUser(t *testing.T) {
	repo, mock, cleanup := newMockUserRepo(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "name", "email"}).AddRow("u1", "An", "an@example.com")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE id = $1 AND deleted_at IS NULL ORDER BY "users"."id" LIMIT $2`)).
		WithArgs("u1", 1).
		WillReturnRows(rows)

	user, err := repo.FindByID("u1")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if user == nil || user.ID != "u1" {
		t.Fatalf("expected user u1, got %#v", user)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUserRepoFindByIDReturnsNilWhenNotFound(t *testing.T) {
	repo, mock, cleanup := newMockUserRepo(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "name", "email"})
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE id = $1 AND deleted_at IS NULL ORDER BY "users"."id" LIMIT $2`)).
		WithArgs("missing", 1).
		WillReturnRows(rows)

	user, err := repo.FindByID("missing")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if user != nil {
		t.Fatalf("expected nil user, got %#v", user)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUserRepoFindByIDReturnsDatabaseError(t *testing.T) {
	repo, mock, cleanup := newMockUserRepo(t)
	defer cleanup()

	dbErr := errors.New("db error")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE id = $1 AND deleted_at IS NULL ORDER BY "users"."id" LIMIT $2`)).
		WithArgs("u1", 1).
		WillReturnError(dbErr)

	user, err := repo.FindByID("u1")
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected db error, got %v", err)
	}
	if user != nil {
		t.Fatalf("expected nil user, got %#v", user)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUserRepoCreateInTxReturnsSuccess(t *testing.T) {
	sqlDB2, innerMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create inner sqlmock: %v", err)
	}
	defer sqlDB2.Close()
	innerDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn:                 sqlDB2,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("open inner gorm db: %v", err)
	}

	repo := NewUserRepo(innerDB)
	user := &models.User{Name: "An", Email: "an@example.com"}
	innerMock.ExpectBegin()
	innerMock.ExpectQuery(`INSERT INTO "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("new-id"))
	innerMock.ExpectCommit()

	tx := innerDB.Begin()
	if err := repo.CreateInTx(tx, user); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	_ = tx.Commit()
	if err := innerMock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUserRepoCreateInTxReturnsDatabaseError(t *testing.T) {
	sqlDB2, innerMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create inner sqlmock: %v", err)
	}
	defer sqlDB2.Close()
	innerDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn:                 sqlDB2,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("open inner gorm db: %v", err)
	}

	repo := NewUserRepo(innerDB)
	dbErr := errors.New("insert error")
	user := &models.User{Name: "An", Email: "an@example.com"}
	innerMock.ExpectBegin()
	innerMock.ExpectQuery(`INSERT INTO "users"`).WillReturnError(dbErr)
	innerMock.ExpectRollback()

	tx := innerDB.Begin()
	err = repo.CreateInTx(tx, user)
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected insert error, got %v", err)
	}
}

func TestUserRepoUpdateReturnsSuccess(t *testing.T) {
	repo, mock, cleanup := newMockUserRepo(t)
	defer cleanup()

	user := &models.User{ID: "u1", Name: "Updated"}
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "users"`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := repo.Update(user); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUserRepoUpdateReturnsDatabaseError(t *testing.T) {
	repo, mock, cleanup := newMockUserRepo(t)
	defer cleanup()

	dbErr := errors.New("db error")
	user := &models.User{ID: "u1", Name: "Updated"}
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "users"`).WillReturnError(dbErr)
	mock.ExpectRollback()

	err := repo.Update(user)
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected db error, got %v", err)
	}
}

func TestUserRepoList_NoFilter(t *testing.T) {
	repo, mock, cleanup := newMockUserRepo(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "users" WHERE deleted_at IS NULL`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	rows := sqlmock.NewRows([]string{"id", "name", "email", "role", "status"}).
		AddRow("u1", "User 1", "u1@example.com", models.UserRoleCitizen, models.UserStatusActive).
		AddRow("u2", "User 2", "u2@example.com", models.UserRoleStaff, models.UserStatusActive)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT $1`)).
		WithArgs(10).
		WillReturnRows(rows)

	users, total, err := repo.List(UserFilter{}, 0, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected total 2, got %d", total)
	}
	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUserRepoList_WithSearch(t *testing.T) {
	repo, mock, cleanup := newMockUserRepo(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "users" WHERE deleted_at IS NULL AND (LOWER(name) LIKE $1 OR LOWER(email) LIKE $2)`)).
		WithArgs("%alice%", "%alice%").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	rows := sqlmock.NewRows([]string{"id", "name", "email"}).
		AddRow("u1", "Alice", "alice@example.com")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE deleted_at IS NULL AND (LOWER(name) LIKE $1 OR LOWER(email) LIKE $2) ORDER BY created_at DESC LIMIT $3`)).
		WithArgs("%alice%", "%alice%", 10).
		WillReturnRows(rows)

	users, total, err := repo.List(UserFilter{Search: "alice"}, 0, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || len(users) != 1 {
		t.Fatalf("expected 1 user, got total=%d, len=%d", total, len(users))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUserRepoList_WithRole(t *testing.T) {
	repo, mock, cleanup := newMockUserRepo(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "users" WHERE deleted_at IS NULL AND role = $1`)).
		WithArgs(string(models.UserRoleStaff)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	rows := sqlmock.NewRows([]string{"id", "name", "email", "role"}).
		AddRow("u1", "Staff 1", "staff@example.com", models.UserRoleStaff)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE deleted_at IS NULL AND role = $1 ORDER BY created_at DESC LIMIT $2`)).
		WithArgs(string(models.UserRoleStaff), 10).
		WillReturnRows(rows)

	users, total, err := repo.List(UserFilter{Role: string(models.UserRoleStaff)}, 0, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || len(users) != 1 {
		t.Fatalf("expected 1 user, got total=%d", total)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUserRepoList_WithRoles(t *testing.T) {
	repo, mock, cleanup := newMockUserRepo(t)
	defer cleanup()

	roles := []string{string(models.UserRoleStaff), string(models.UserRoleManager)}
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "users" WHERE deleted_at IS NULL AND role IN ($1,$2)`)).
		WithArgs(roles[0], roles[1]).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	rows := sqlmock.NewRows([]string{"id", "name", "email", "role"}).
		AddRow("u1", "Staff", "staff@example.com", models.UserRoleStaff).
		AddRow("u2", "Manager", "mgr@example.com", models.UserRoleManager)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE deleted_at IS NULL AND role IN ($1,$2) ORDER BY created_at DESC LIMIT $3`)).
		WithArgs(roles[0], roles[1], 10).
		WillReturnRows(rows)

	users, total, err := repo.List(UserFilter{Roles: roles}, 0, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 || len(users) != 2 {
		t.Fatalf("expected 2 users, got total=%d", total)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUserRepoList_CountError(t *testing.T) {
	repo, mock, cleanup := newMockUserRepo(t)
	defer cleanup()

	dbErr := errors.New("count error")
	mock.ExpectQuery(`SELECT count\(\*\) FROM "users"`).WillReturnError(dbErr)

	_, _, err := repo.List(UserFilter{}, 0, 10)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUserRepoList_FindError(t *testing.T) {
	repo, mock, cleanup := newMockUserRepo(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT count\(\*\) FROM "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	dbErr := errors.New("find error")
	mock.ExpectQuery(`SELECT \* FROM "users"`).WillReturnError(dbErr)

	_, _, err := repo.List(UserFilter{}, 0, 10)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUserRepoUpdateStatus(t *testing.T) {
	repo, mock, cleanup := newMockUserRepo(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "users" SET`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "u1").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.UpdateStatus("u1", models.UserStatusBlocked, "admin-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUserRepoUpdateStatus_Error(t *testing.T) {
	repo, mock, cleanup := newMockUserRepo(t)
	defer cleanup()

	dbErr := errors.New("update error")
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "users" SET`)).
		WillReturnError(dbErr)
	mock.ExpectRollback()

	err := repo.UpdateStatus("u1", models.UserStatusBlocked, "admin-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUserRepoSoftDelete(t *testing.T) {
	repo, mock, cleanup := newMockUserRepo(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "users" SET`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "u1").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.SoftDelete("u1", "admin-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUserRepoSoftDelete_Error(t *testing.T) {
	repo, mock, cleanup := newMockUserRepo(t)
	defer cleanup()

	dbErr := errors.New("delete error")
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "users" SET`)).
		WillReturnError(dbErr)
	mock.ExpectRollback()

	err := repo.SoftDelete("u1", "admin-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
