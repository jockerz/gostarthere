package repository

import (
	"context"
	"testing"
	"time"

	"vnti/pkg/entities"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func authStringPtr(s string) *string { return &s }

func setupAuthRepoDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	db.AutoMigrate(&entities.User{}, &entities.AuthToken{}, &entities.UserToken{})
	return db
}

func TestRepositoryCreateAuthToken(t *testing.T) {
	db := setupAuthRepoDB(t)
	repo := NewAuthRepository(db)
	ctx := context.Background()

	userRepo := NewUserRepository(db)
	user, _ := userRepo.Create(ctx, &entities.User{Email: "tok@test.com", Username: "tok", Password: authStringPtr("hash")})

	token := &entities.AuthToken{
		UserID:    user.ID,
		Prefix:    "auth_prefix",
		Secret:    "auth-secret-value",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	created, err := repo.CreateAuthToken(ctx, token)
	if err != nil {
		t.Fatal(err)
	}
	if created.ID == 0 {
		t.Fatal("expected non-zero token ID")
	}
	if created.Prefix != "auth_prefix" {
		t.Fatalf("expected 'auth_prefix', got '%s'", created.Prefix)
	}
	if created.Secret != "auth-secret-value" {
		t.Fatalf("expected 'auth-secret-value', got '%s'", created.Secret)
	}
}

func TestRepositoryFindAuthToken(t *testing.T) {
	db := setupAuthRepoDB(t)
	repo := NewAuthRepository(db)
	ctx := context.Background()

	userRepo := NewUserRepository(db)
	user, _ := userRepo.Create(ctx, &entities.User{Email: "findtok@test.com", Username: "findtok", Password: authStringPtr("hash")})
	repo.CreateAuthToken(ctx, &entities.AuthToken{
		UserID:    user.ID,
		Prefix:    "find-auth",
		Secret:    "secret",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})

	found, err := repo.FindAuthToken(ctx, "find-auth")
	if err != nil {
		t.Fatal(err)
	}
	if found.Prefix != "find-auth" {
		t.Fatalf("expected 'find-auth', got '%s'", found.Prefix)
	}
}

func TestRepositoryFindAuthTokenNotFound(t *testing.T) {
	db := setupAuthRepoDB(t)
	repo := NewAuthRepository(db)
	ctx := context.Background()

	_, err := repo.FindAuthToken(ctx, "nonexistent")
	if err == nil {
		t.Fatal("expected error for non-existent auth token")
	}
}

func TestRepositoryRefreshAuthToken(t *testing.T) {
	db := setupAuthRepoDB(t)
	repo := NewAuthRepository(db)
	ctx := context.Background()

	userRepo := NewUserRepository(db)
	user, _ := userRepo.Create(ctx, &entities.User{Email: "refreshtok@test.com", Username: "refreshtok", Password: authStringPtr("hash")})
	now := time.Now()
	original, _ := repo.CreateAuthToken(ctx, &entities.AuthToken{
		UserID:    user.ID,
		Prefix:    "refresh-auth",
		Secret:    "old-secret",
		ExpiresAt: now.Add(24 * time.Hour),
	})

	original.Secret = "new-secret"
	original.ExpiresAt = now.Add(48 * time.Hour)
	original.RefreshSecret = "new-refresh"
	original.RefreshedAt = now

	err := repo.RefreshAuthToken(ctx, original)
	if err != nil {
		t.Fatal(err)
	}

	fresh, _ := repo.FindAuthToken(ctx, "refresh-auth")
	if fresh.Secret != "new-secret" {
		t.Fatalf("expected 'new-secret', got '%s'", fresh.Secret)
	}
}

func TestRepositoryRevokeAuthToken(t *testing.T) {
	db := setupAuthRepoDB(t)
	repo := NewAuthRepository(db)
	ctx := context.Background()

	userRepo := NewUserRepository(db)
	user, _ := userRepo.Create(ctx, &entities.User{Email: "revoketok@test.com", Username: "revoketok", Password: authStringPtr("hash")})
	repo.CreateAuthToken(ctx, &entities.AuthToken{
		UserID:    user.ID,
		Prefix:    "revoke-auth",
		Secret:    "secret",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})

	err := repo.RevokeAuthToken(ctx, "revoke-auth")
	if err != nil {
		t.Fatal(err)
	}

	found, _ := repo.FindAuthToken(ctx, "revoke-auth")
	if !found.IsRevoked {
		t.Fatal("expected token to be revoked")
	}
	if found.RevokedAt == nil {
		t.Fatal("expected RevokedAt to be set")
	}
}

func TestRepositoryCreateUserToken(t *testing.T) {
	db := setupAuthRepoDB(t)
	repo := NewAuthRepository(db)
	ctx := context.Background()

	userRepo := NewUserRepository(db)
	user, _ := userRepo.Create(ctx, &entities.User{Email: "usertok@test.com", Username: "usertok", Password: authStringPtr("hash")})

	token := &entities.UserToken{
		UserID:    user.ID,
		Prefix:    "ut_prefix",
		Secret:    "ut-secret-value",
		Type:      entities.TokenActivation,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	created, err := repo.CreateUserToken(ctx, token)
	if err != nil {
		t.Fatal(err)
	}
	if created.ID == 0 {
		t.Fatal("expected non-zero token ID")
	}
	if created.Prefix != "ut_prefix" {
		t.Fatalf("expected 'ut_prefix', got '%s'", created.Prefix)
	}
	if created.Type != entities.TokenActivation {
		t.Fatalf("expected TokenActivation, got '%s'", created.Type)
	}
}

func TestRepositoryFindUserToken(t *testing.T) {
	db := setupAuthRepoDB(t)
	repo := NewAuthRepository(db)
	ctx := context.Background()

	userRepo := NewUserRepository(db)
	user, _ := userRepo.Create(ctx, &entities.User{Email: "findut@test.com", Username: "findut", Password: authStringPtr("hash")})
	repo.CreateUserToken(ctx, &entities.UserToken{
		UserID:    user.ID,
		Prefix:    "find-ut",
		Secret:    "secret",
		Type:      entities.TokenReset,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})

	found, err := repo.FindUserToken(ctx, "find-ut", entities.TokenReset)
	if err != nil {
		t.Fatal(err)
	}
	if found.Prefix != "find-ut" {
		t.Fatalf("expected 'find-ut', got '%s'", found.Prefix)
	}
}

func TestRepositoryFindUserTokenNotFound(t *testing.T) {
	db := setupAuthRepoDB(t)
	repo := NewAuthRepository(db)
	ctx := context.Background()

	_, err := repo.FindUserToken(ctx, "nonexistent", entities.TokenActivation)
	if err == nil {
		t.Fatal("expected error for non-existent user token")
	}
}

func TestRepositoryFindUserTokenWrongType(t *testing.T) {
	db := setupAuthRepoDB(t)
	repo := NewAuthRepository(db)
	ctx := context.Background()

	userRepo := NewUserRepository(db)
	user, _ := userRepo.Create(ctx, &entities.User{Email: "wtut@test.com", Username: "wtut", Password: authStringPtr("hash")})
	repo.CreateUserToken(ctx, &entities.UserToken{
		UserID:    user.ID,
		Prefix:    "wrong-type-ut",
		Secret:    "secret",
		Type:      entities.TokenReset,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})

	_, err := repo.FindUserToken(ctx, "wrong-type-ut", entities.TokenActivation)
	if err == nil {
		t.Fatal("expected error for wrong token type")
	}
}

func TestRepositoryRefreshUserToken(t *testing.T) {
	db := setupAuthRepoDB(t)
	repo := NewAuthRepository(db)
	ctx := context.Background()

	userRepo := NewUserRepository(db)
	user, _ := userRepo.Create(ctx, &entities.User{Email: "refreshut@test.com", Username: "refreshut", Password: authStringPtr("hash")})
	now := time.Now()
	original, _ := repo.CreateUserToken(ctx, &entities.UserToken{
		UserID:    user.ID,
		Prefix:    "refresh-ut",
		Secret:    "old-secret",
		Type:      entities.TokenActivation,
		ExpiresAt: now.Add(24 * time.Hour),
	})

	original.Secret = "new-secret"
	original.ExpiresAt = now.Add(48 * time.Hour)
	original.RefreshSecret = "new-refresh"
	original.RefreshedAt = now

	err := repo.RefreshUserToken(ctx, original)
	if err != nil {
		t.Fatal(err)
	}

	fresh, _ := repo.FindUserToken(ctx, "refresh-ut", entities.TokenActivation)
	if fresh.Secret != "new-secret" {
		t.Fatalf("expected 'new-secret', got '%s'", fresh.Secret)
	}
}

func TestRepositoryRevokeUserToken(t *testing.T) {
	db := setupAuthRepoDB(t)
	repo := NewAuthRepository(db)
	ctx := context.Background()

	userRepo := NewUserRepository(db)
	user, _ := userRepo.Create(ctx, &entities.User{Email: "revokeut@test.com", Username: "revokeut", Password: authStringPtr("hash")})
	repo.CreateUserToken(ctx, &entities.UserToken{
		UserID:    user.ID,
		Prefix:    "revoke-ut",
		Secret:    "secret",
		Type:      entities.TokenActivation,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})

	err := repo.RevokeUserToken(ctx, "revoke-ut")
	if err != nil {
		t.Fatal(err)
	}

	found, _ := repo.FindUserToken(ctx, "revoke-ut", entities.TokenActivation)
	if !found.IsRevoked {
		t.Fatal("expected token to be revoked")
	}
	if found.RevokedAt == nil {
		t.Fatal("expected RevokedAt to be set")
	}
}

func TestRepositoryMarkuAsUsedUserToken(t *testing.T) {
	db := setupAuthRepoDB(t)
	repo := NewAuthRepository(db)
	ctx := context.Background()

	userRepo := NewUserRepository(db)
	user, _ := userRepo.Create(ctx, &entities.User{Email: "markut@test.com", Username: "markut", Password: authStringPtr("hash")})
	token, _ := repo.CreateUserToken(ctx, &entities.UserToken{
		UserID:    user.ID,
		Prefix:    "mark-ut",
		Secret:    "secret",
		Type:      entities.TokenReset,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	})

	err := repo.MarkAsUsedUserToken(ctx, token.ID)
	if err != nil {
		t.Fatal(err)
	}

	found, _ := repo.FindUserToken(ctx, "mark-ut", entities.TokenReset)
	if !found.IsUsed {
		t.Fatal("expected token to be marked used")
	}
	if found.UsedAt == nil {
		t.Fatal("expected UsedAt to be set")
	}
}

func TextRepositoryUpdate(t *testing.T) {
	db := setupAuthRepoDB(t)
	ctx := context.Background()

	userRepo := NewUserRepository(db)
	user, _ := userRepo.Create(ctx, &entities.User{Email: "markut@test.com", Username: "markut", Password: authStringPtr("hash")})

	if *user.Password == "hash" {
		t.Fatalf("expected hashed password. got %s", *user.Password)
	}
}
