package service

import (
	"SneakerFlash/internal/model"
	"SneakerFlash/internal/repository"
	"SneakerFlash/internal/testutil"
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"
)

func newUserServiceForTest(t *testing.T) (*UserService, *repository.UserRepo, *gorm.DB) {
	t.Helper()

	testutil.SetupTestConfig()
	db := testutil.NewSQLiteDB(t)
	repo := repository.NewUserRepo(db)
	return NewUserService(repo), repo, db
}

func TestUserService_RegisterAndLogin(t *testing.T) {
	svc, repo, _ := newUserServiceForTest(t)
	ctx := context.Background()

	if err := svc.Register(ctx, "alice", "password-123"); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	user, err := repo.GetByUsername(ctx, "alice")
	if err != nil {
		t.Fatalf("GetByUsername() error = %v", err)
	}
	if user.Password == "password-123" {
		t.Fatal("password stored in plaintext")
	}

	access, refresh, err := svc.Login(ctx, "alice", "password-123")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if access == "" || refresh == "" {
		t.Fatalf("Login() returned empty token, access=%q refresh=%q", access, refresh)
	}
}

func TestUserService_RegisterDuplicate(t *testing.T) {
	svc, _, _ := newUserServiceForTest(t)
	ctx := context.Background()

	if err := svc.Register(ctx, "alice", "password-123"); err != nil {
		t.Fatalf("first Register() error = %v", err)
	}

	err := svc.Register(ctx, "alice", "password-456")
	if !errors.Is(err, ErrUserExited) {
		t.Fatalf("duplicate Register() error = %v, want %v", err, ErrUserExited)
	}
}

func TestUserService_LoginWrongPassword(t *testing.T) {
	svc, _, _ := newUserServiceForTest(t)
	ctx := context.Background()

	if err := svc.Register(ctx, "alice", "password-123"); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	_, _, err := svc.Login(ctx, "alice", "wrong-password")
	if !errors.Is(err, ErrPasswordWrong) {
		t.Fatalf("Login() error = %v, want %v", err, ErrPasswordWrong)
	}
}

func TestUserService_RefreshRejectsAccessToken(t *testing.T) {
	svc, _, _ := newUserServiceForTest(t)
	ctx := context.Background()

	if err := svc.Register(ctx, "alice", "password-123"); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	access, _, err := svc.Login(ctx, "alice", "password-123")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	if _, err := svc.Refresh(ctx, access); !errors.Is(err, ErrTokenInvalid) {
		t.Fatalf("Refresh(access token) error = %v, want %v", err, ErrTokenInvalid)
	}
}

func TestUserService_UpdateProfileRejectsDuplicateName(t *testing.T) {
	svc, repo, _ := newUserServiceForTest(t)
	ctx := context.Background()

	for _, user := range []model.User{
		{Username: "alice", Password: "hashed"},
		{Username: "bob", Password: "hashed"},
	} {
		u := user
		if err := repo.Create(ctx, &u); err != nil {
			t.Fatalf("Create(%s) error = %v", u.Username, err)
		}
	}

	dup := "bob"
	if _, err := svc.UpdateProfile(ctx, 1, &dup, nil); !errors.Is(err, ErrUserExited) {
		t.Fatalf("UpdateProfile() error = %v, want %v", err, ErrUserExited)
	}
}
