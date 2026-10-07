package repository

import (
	"context"
	"testing"

	"vnti/pkg/entities"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func userStringPtr(s string) *string { return &s }

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	db.AutoMigrate(&entities.User{})
	return db
}

func TestRepositoryCreate(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &entities.User{
		Email:    "test@example.com",
		Username: "testuser",
		Name:     "Test",
		Password: userStringPtr("hashed"),
		Active:   false,
	}

	created, err := repo.Create(ctx, user)
	if err != nil {
		t.Fatal(err)
	}
	if created.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if created.Email != "test@example.com" {
		t.Fatalf("expected test@example.com, got %s", created.Email)
	}
}

func TestRepositoryCreateDuplicateEmail(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	repo.Create(ctx, &entities.User{Email: "dup@test.com", Username: "first", Password: userStringPtr("hash")})

	_, err := repo.Create(ctx, &entities.User{Email: "dup@test.com", Username: "second", Password: userStringPtr("hash")})
	if err == nil {
		t.Fatal("expected error for duplicate email")
	}
}

func TestRepositoryCreateDuplicateUsername(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	repo.Create(ctx, &entities.User{Email: "first@test.com", Username: "dupuser", Password: userStringPtr("hash")})

	_, err := repo.Create(ctx, &entities.User{Email: "second@test.com", Username: "dupuser", Password: userStringPtr("hash")})
	if err == nil {
		t.Fatal("expected error for duplicate username")
	}
}

func TestRepositoryFindByEmail(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	repo.Create(ctx, &entities.User{Email: "findme@example.com", Username: "findme", Password: userStringPtr("hashed")})

	found, err := repo.FindByEmail(ctx, "findme@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if found.Email != "findme@example.com" {
		t.Fatalf("expected findme@example.com, got %s", found.Email)
	}
}

func TestRepositoryFindByEmailUpperLowerCase(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	repo.Create(ctx, &entities.User{Email: "findme@example.com", Username: "findme", Password: userStringPtr("hashed")})

	found, err := repo.FindByEmail(ctx, "FindMe@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if found.Email != "findme@example.com" {
		t.Fatalf("expected findme@example.com, got %s", found.Email)
	}
}

func TestRepositoryFindByEmailNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	_, err := repo.FindByEmail(ctx, "nobody@example.com")
	if err == nil {
		t.Fatal("expected error for non-existent email")
	}
}

func TestRepositoryFindByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	created, _ := repo.Create(ctx, &entities.User{Email: "id@test.com", Username: "id", Password: userStringPtr("hashed")})

	found, err := repo.FindByID(ctx, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if found.ID != created.ID {
		t.Fatalf("expected ID %d, got %d", created.ID, found.ID)
	}
}

func TestRepositoryFindByIDNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	_, err := repo.FindByID(ctx, 999)
	if err == nil {
		t.Fatal("expected error for non-existent ID")
	}
}

func TestRepositoryFindByUsername(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	repo.Create(ctx, &entities.User{Email: "user@test.com", Username: "uniqueuser", Password: userStringPtr("hash")})

	found, err := repo.FindByUsername(ctx, "uniqueuser")
	if err != nil {
		t.Fatal(err)
	}
	if found.Username != "uniqueuser" {
		t.Fatalf("expected 'uniqueuser', got '%s'", found.Username)
	}
}

func TestRepositoryFindByUsernameNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	_, err := repo.FindByUsername(ctx, "nobody")
	if err == nil {
		t.Fatal("expected error for non-existent username")
	}
}

func TestRepositoryUpdate(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	created, _ := repo.Create(ctx, &entities.User{Email: "update@test.com", Username: "update", Name: "Old", Password: userStringPtr("userStringPtr")})

	created.Name = "New"
	updated, err := repo.Update(ctx, created)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "New" {
		t.Fatalf("expected 'New', got '%s'", updated.Name)
	}

	fresh, _ := repo.FindByID(ctx, created.ID)
	if fresh.Name != "New" {
		t.Fatalf("expected 'New' from DB, got '%s'", fresh.Name)
	}
}

func TestRepositoryDelete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	created, _ := repo.Create(ctx, &entities.User{Email: "delete@test.com", Username: "del", Password: userStringPtr("userStringPtr")})

	err := repo.Delete(ctx, created)
	if err != nil {
		t.Fatal(err)
	}

	_, err = repo.FindByID(ctx, created.ID)
	if err == nil {
		t.Fatal("expected error after delete")
	}
}
