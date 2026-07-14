package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"vnti/internal"
	"vnti/pkg/entities"

	"github.com/hibiken/asynq"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type mockRepo struct {
	users        map[uint]*entities.User
	nextID       uint
	createErr    error
	findByIDFunc func(id uint) (*entities.User, error)
	updateErr    error
	deleteErr    error
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		users:  make(map[uint]*entities.User),
		nextID: 1,
	}
}

func (m *mockRepo) Create(_ context.Context, u *entities.User) (*entities.User, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	u.ID = m.nextID
	m.nextID++
	m.users[u.ID] = u
	return u, nil
}

func (m *mockRepo) FindByEmail(_ context.Context, email string) (*entities.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockRepo) FindByID(_ context.Context, id uint) (*entities.User, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(id)
	}
	u, ok := m.users[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return u, nil
}

func (m *mockRepo) FindByUsername(_ context.Context, username string) (*entities.User, error) {
	for _, u := range m.users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockRepo) Update(_ context.Context, u *entities.User) (*entities.User, error) {
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	m.users[u.ID] = u
	return u, nil
}

func (m *mockRepo) UpdateByID(_ context.Context, userId uint, data map[string]any, uBy uint) error {
	if m.updateErr != nil {
		return m.updateErr
	}

	u, ok := m.users[userId]
	if !ok {
		return gorm.ErrRecordNotFound
	}

	if v, ok := data["name"]; ok {
		u.Name, _ = v.(string)
	}
	if v, ok := data["username"]; ok {
		u.Username, _ = v.(string)
	}
	if v, ok := data["avatar"]; ok {
		u.Avatar, _ = v.(string)
	}
	if v, ok := data["email"]; ok {
		u.Email, _ = v.(string)
	}
	if v, ok := data["password"]; ok {
		s := v.(string)
		u.Password = &s
	}
	u.UpdatedAt = time.Now()
	u.UpdatedBy = uBy

	return nil
}

func (m *mockRepo) Delete(_ context.Context, u *entities.User) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.users, u.ID)
	return nil
}

func TestServiceCreate(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(&internal.Config{}, repo, nil, &asynq.Client{})
	svc.SetSkipTaskQueue(true)
	ctx := context.Background()

	user := &entities.User{Email: "test@test.com", Username: "test", Password: p("secret")}
	created, err := svc.Create(ctx, user)
	if err != nil {
		t.Fatal(err)
	}
	if created.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if created.Email != "test@test.com" {
		t.Fatalf("expected 'test@test.com', got '%s'", created.Email)
	}
}

func TestServiceCreateLowercasesEmailAndUsername(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(&internal.Config{}, repo, nil, &asynq.Client{})
	svc.SetSkipTaskQueue(true)

	ctx := context.Background()

	created, err := svc.Create(ctx, &entities.User{Email: "UPPER@TEST.COM", Username: "UserName", Password: p("secret")})
	if err != nil {
		t.Fatal(err)
	}
	if created.Email != "upper@test.com" {
		t.Fatalf("expected 'upper@test.com', got '%s'", created.Email)
	}
	if created.Username != "username" {
		t.Fatalf("expected 'username', got '%s'", created.Username)
	}
}

func TestServiceCreateMissingEmail(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(&internal.Config{}, repo, nil, &asynq.Client{})
	svc.SetSkipTaskQueue(true)
	ctx := context.Background()

	_, err := svc.Create(ctx, &entities.User{Username: "test", Password: p("secret")})
	if err == nil {
		t.Fatal("expected error for missing email")
	}
}

func TestServiceCreateMissingUsername(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(&internal.Config{}, repo, nil, &asynq.Client{})
	svc.SetSkipTaskQueue(true)
	ctx := context.Background()

	_, err := svc.Create(ctx, &entities.User{Email: "test@test.com", Password: p("secret")})
	if err == nil {
		t.Fatal("expected error for missing username")
	}
}

func TestServiceCreateMissingPassword(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(&internal.Config{}, repo, nil, &asynq.Client{})
	svc.SetSkipTaskQueue(true)
	ctx := context.Background()

	_, err := svc.Create(ctx, &entities.User{Email: "test@test.com", Username: "test"})
	if err == nil {
		t.Fatal("expected error for missing password")
	}
}

func TestServiceCreateRepoError(t *testing.T) {
	repo := newMockRepo()
	repo.createErr = errors.New("db down")
	svc := NewService(&internal.Config{}, repo, nil, &asynq.Client{})
	svc.SetSkipTaskQueue(true)
	ctx := context.Background()

	_, err := svc.Create(ctx, &entities.User{Email: "test@test.com", Username: "test", Password: p("secret")})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestServiceFindByID(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(&internal.Config{}, repo, nil, &asynq.Client{})
	svc.SetSkipTaskQueue(true)
	ctx := context.Background()

	user, _ := svc.Create(ctx, &entities.User{Email: "test@test.com", Username: "test", Password: p("secret")})
	found, err := svc.FindByID(ctx, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if found.ID != user.ID {
		t.Fatalf("expected ID %d, got %d", user.ID, found.ID)
	}
}

func TestServiceFindByIDNotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(&internal.Config{}, repo, nil, &asynq.Client{})
	svc.SetSkipTaskQueue(true)
	ctx := context.Background()

	_, err := svc.FindByID(ctx, 999)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestServiceFindByEmail(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(&internal.Config{}, repo, nil, &asynq.Client{})
	svc.SetSkipTaskQueue(true)
	ctx := context.Background()

	svc.Create(ctx, &entities.User{Email: "find@test.com", Username: "find", Password: p("secret")})
	found, err := svc.FindByEmail(ctx, "find@test.com")
	if err != nil {
		t.Fatal(err)
	}
	if found.Email != "find@test.com" {
		t.Fatalf("expected 'find@test.com', got '%s'", found.Email)
	}
}

func TestServiceFindByEmailNotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(&internal.Config{}, repo, nil, &asynq.Client{})
	svc.SetSkipTaskQueue(true)
	ctx := context.Background()

	_, err := svc.FindByEmail(ctx, "nonexistent@test.com")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestServiceFindByUsername(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(&internal.Config{}, repo, nil, &asynq.Client{})
	svc.SetSkipTaskQueue(true)
	ctx := context.Background()

	svc.Create(ctx, &entities.User{Email: "user@test.com", Username: "findme", Password: p("secret")})
	found, err := svc.FindByUsername(ctx, "findme")
	if err != nil {
		t.Fatal(err)
	}
	if found.Username != "findme" {
		t.Fatalf("expected 'findme', got '%s'", found.Username)
	}
}

func TestServiceFindByUsernameNotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(&internal.Config{}, repo, nil, &asynq.Client{})
	svc.SetSkipTaskQueue(true)
	ctx := context.Background()

	_, err := svc.FindByUsername(ctx, "nobody")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestServiceUpdate(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(&internal.Config{}, repo, nil, &asynq.Client{})
	svc.SetSkipTaskQueue(true)
	ctx := context.Background()

	user, _ := svc.Create(ctx, &entities.User{Email: "test@test.com", Username: "test", Password: p("secret")})
	user.Name = "Updated"
	updated, err := svc.Update(ctx, user)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "Updated" {
		t.Fatalf("expected 'Updated', got '%s'", updated.Name)
	}
}

func TestServiceUpdateNotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(&internal.Config{}, repo, nil, &asynq.Client{})
	svc.SetSkipTaskQueue(true)
	ctx := context.Background()

	_, err := svc.Update(ctx, &entities.User{ID: 999})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestServiceUpdateRepoError(t *testing.T) {
	repo := newMockRepo()
	repo.updateErr = errors.New("update failed")
	svc := NewService(&internal.Config{}, repo, nil, &asynq.Client{})
	svc.SetSkipTaskQueue(true)
	ctx := context.Background()

	user, _ := svc.Create(ctx, &entities.User{Email: "test@test.com", Username: "test", Password: p("secret")})
	_, err := svc.Update(ctx, user)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrUpdateFailed) {
		t.Fatalf("expected ErrUpdateFailed, got %v", err)
	}
}

func TestServiceDelete(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(&internal.Config{}, repo, nil, &asynq.Client{})
	svc.SetSkipTaskQueue(true)
	ctx := context.Background()

	user, _ := svc.Create(ctx, &entities.User{Email: "test@test.com", Username: "test", Password: p("secret")})
	err := svc.Delete(ctx, user)
	if err != nil {
		t.Fatal(err)
	}
}

func TestServiceDeleteNotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(&internal.Config{}, repo, nil, &asynq.Client{})
	svc.SetSkipTaskQueue(true)
	ctx := context.Background()

	err := svc.Delete(ctx, &entities.User{ID: 999})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestServiceDeleteRepoError(t *testing.T) {
	repo := newMockRepo()
	repo.deleteErr = errors.New("delete failed")
	svc := NewService(&internal.Config{}, repo, nil, &asynq.Client{})
	svc.SetSkipTaskQueue(true)
	ctx := context.Background()

	user, _ := svc.Create(ctx, &entities.User{Email: "test@test.com", Username: "test", Password: p("secret")})
	err := svc.Delete(ctx, user)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrDeleteFailed) {
		t.Fatalf("expected ErrDeleteFailed, got %v", err)
	}
}

func TestServiceUpdateProfile(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(&internal.Config{}, repo, nil, &asynq.Client{})
	svc.SetSkipTaskQueue(true)
	ctx := context.Background()

	user, _ := svc.Create(ctx, &entities.User{Email: "test@test.com", Username: "test", Password: p("secret")})

	updated, err := svc.UpdateProfile(ctx, user.ID, "New Name", "newusername", "new-avatar-url")
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "New Name" {
		t.Fatalf("expected 'New Name', got '%s'", updated.Name)
	}
	if updated.Username != "newusername" {
		t.Fatalf("expected 'newusername', got '%s'", updated.Username)
	}
	if updated.Avatar != "new-avatar-url" {
		t.Fatalf("expected 'new-avatar-url', got '%s'", updated.Avatar)
	}
}

func TestServiceUpdateProfileEmptyFields(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(&internal.Config{}, repo, nil, &asynq.Client{})
	svc.SetSkipTaskQueue(true)
	ctx := context.Background()

	user, _ := svc.Create(ctx, &entities.User{Email: "test@test.com", Username: "test", Name: "Original", Avatar: "orig.jpg", Password: p("secret")})

	updated, err := svc.UpdateProfile(ctx, user.ID, "", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "Original" {
		t.Fatalf("name should remain 'Original', got '%s'", updated.Name)
	}
	if updated.Avatar != "orig.jpg" {
		t.Fatalf("avatar should remain 'orig.jpg', got '%s'", updated.Avatar)
	}
}

func TestServiceUpdateProfileNotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(&internal.Config{}, repo, nil, &asynq.Client{})
	svc.SetSkipTaskQueue(true)
	ctx := context.Background()

	_, err := svc.UpdateProfile(ctx, 999, "Name", "user", "")
	if err == nil {
		t.Fatal("expected error for non-existent user")
	}
	if !errors.Is(err, ErrUpdateFailed) {
		t.Fatalf("expected ErrUpdateFailed, got %v", err)
	}
}

func TestServiceChangePasswordSuccess(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(&internal.Config{}, repo, nil, &asynq.Client{})
	svc.SetSkipTaskQueue(true)
	ctx := context.Background()

	hashed, _ := bcrypt.GenerateFromPassword([]byte("oldpass"), bcrypt.DefaultCost)
	user, _ := svc.Create(ctx, &entities.User{Email: "cpw@test.com", Username: "cpw", Password: p(string(hashed))})

	err := svc.ChangePassword(ctx, user.ID, "oldpass", "newpass")
	if err != nil {
		t.Fatal(err)
	}

	updated, _ := repo.FindByID(ctx, user.ID)
	if err := bcrypt.CompareHashAndPassword([]byte(*updated.Password), []byte("newpass")); err != nil {
		t.Fatal("password should be updated")
	}
}

func TestServiceChangePasswordWrongCurrent(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(&internal.Config{}, repo, nil, &asynq.Client{})
	svc.SetSkipTaskQueue(true)
	ctx := context.Background()

	hashed, _ := bcrypt.GenerateFromPassword([]byte("oldpass"), bcrypt.DefaultCost)
	user, _ := svc.Create(ctx, &entities.User{Email: "cpw@test.com", Username: "cpw", Password: p(string(hashed))})

	err := svc.ChangePassword(ctx, user.ID, "wrongpass", "newpass")
	if err == nil {
		t.Fatal("expected error for wrong current password")
	}
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestServiceChangePasswordUserNotFound(t *testing.T) {
	svc := NewService(&internal.Config{}, newMockRepo(), nil, &asynq.Client{})
	svc.SetSkipTaskQueue(true)
	ctx := context.Background()

	err := svc.ChangePassword(ctx, 999, "oldpass", "newpass")
	if err == nil {
		t.Fatal("expected error for non-existent user")
	}
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}
func p(s string) *string { return &s }
