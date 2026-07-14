package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"vnti/internal"
	"vnti/pkg/entities"

	"github.com/hibiken/asynq"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type mockAuthRepo struct {
	users           map[uint]*entities.User
	nextUserID      uint
	userTokens      map[uint]*entities.UserToken
	nextUserTokenID uint
	authTokens      map[uint]*entities.AuthToken
	nextAuthTokenID uint

	createUserTokenErr error
	createAuthTokenErr error
	findByIDFunc       func(ctx context.Context, id uint) (*entities.User, error)
}

func newMockAuthRepo() *mockAuthRepo {
	return &mockAuthRepo{
		users:           make(map[uint]*entities.User),
		nextUserID:      1,
		userTokens:      make(map[uint]*entities.UserToken),
		nextUserTokenID: 1,
		authTokens:      make(map[uint]*entities.AuthToken),
		nextAuthTokenID: 1,
	}
}

func (m *mockAuthRepo) Create(_ context.Context, user *entities.User) (*entities.User, error) {
	user.ID = m.nextUserID
	m.nextUserID++
	m.users[user.ID] = user
	return user, nil
}

func (m *mockAuthRepo) FindByEmail(_ context.Context, email string) (*entities.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockAuthRepo) FindByID(ctx context.Context, id uint) (*entities.User, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	u, ok := m.users[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return u, nil
}

func (m *mockAuthRepo) FindByUsername(_ context.Context, username string) (*entities.User, error) {
	for _, u := range m.users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockAuthRepo) Update(_ context.Context, user *entities.User) (*entities.User, error) {
	m.users[user.ID] = user
	return user, nil
}

func (m *mockAuthRepo) UpdateByID(_ context.Context, userId uint, data map[string]any, uBy uint) error {
	user := m.users[userId]
	if user == nil {
		return nil
	}

	if v, ok := data["password"]; ok {
		s := v.(string)
		user.Password = &s
	}

	return nil
}

func (m *mockAuthRepo) Delete(_ context.Context, user *entities.User) error {
	delete(m.users, user.ID)
	return nil
}

func (m *mockAuthRepo) CreateUserToken(_ context.Context, token *entities.UserToken) (*entities.UserToken, error) {
	if m.createUserTokenErr != nil {
		return nil, m.createUserTokenErr
	}
	token.ID = m.nextUserTokenID
	m.nextUserTokenID++
	m.userTokens[token.ID] = token
	return token, nil
}

func (m *mockAuthRepo) FindUserToken(_ context.Context, prefix string, tokenType entities.UserTokenType) (*entities.UserToken, error) {
	for _, t := range m.userTokens {
		if t.Prefix == prefix && t.Type == tokenType {
			return t, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockAuthRepo) FindUserTokenByUserId(_ context.Context, userId uint, tokenType entities.UserTokenType) (*[]entities.UserToken, error) {
	var tokens []entities.UserToken
	for _, t := range m.userTokens {
		if t.UserID == userId && t.Type == tokenType && !t.IsUsed {
			tokens = append(tokens, *t)
		}
	}
	if tokens == nil {
		tokens = []entities.UserToken{}
	}
	return &tokens, nil
}

func (m *mockAuthRepo) RefreshUserToken(_ context.Context, token *entities.UserToken) error {
	m.userTokens[token.ID] = token
	return nil
}

func (m *mockAuthRepo) RevokeUserToken(_ context.Context, prefix string) error {
	for _, t := range m.userTokens {
		if t.Prefix == prefix {
			now := time.Now()
			t.IsRevoked = true
			t.RevokedAt = &now
			return nil
		}
	}
	return nil
}

func (m *mockAuthRepo) MarkAsUsedUserToken(_ context.Context, tokenId uint) error {
	t, ok := m.userTokens[tokenId]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	now := time.Now()
	t.IsUsed = true
	t.UsedAt = &now
	return nil
}

func (m *mockAuthRepo) CreateAuthToken(_ context.Context, token *entities.AuthToken) (*entities.AuthToken, error) {
	if m.createAuthTokenErr != nil {
		return nil, m.createAuthTokenErr
	}
	token.ID = m.nextAuthTokenID
	m.nextAuthTokenID++
	m.authTokens[token.ID] = token
	return token, nil
}

func (m *mockAuthRepo) FindAuthToken(_ context.Context, prefix string) (*entities.AuthToken, error) {
	for _, t := range m.authTokens {
		if t.Prefix == prefix {
			return t, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockAuthRepo) RefreshAuthToken(_ context.Context, token *entities.AuthToken) error {
	m.authTokens[token.ID] = token
	return nil
}

func (m *mockAuthRepo) RevokeAuthToken(_ context.Context, prefix string) error {
	for _, t := range m.authTokens {
		if t.Prefix == prefix {
			now := time.Now()
			t.IsRevoked = true
			t.RevokedAt = &now
			return nil
		}
	}
	return nil
}

func setupAuthService(t *testing.T) (*mockAuthRepo, Service) {
	t.Helper()
	repo := newMockAuthRepo()
	svc := NewService(&internal.Config{}, repo, repo, "test-secret-key", &asynq.Client{})
	svc.SetSkipTaskQueue(true)
	return repo, svc
}

func TestGenerateAuthToken(t *testing.T) {
	repo, svc := setupAuthService(t)
	ctx := context.Background()

	user, _ := repo.Create(ctx, &entities.User{Email: "gentok@test.com", Username: "gentok", Password: p("hash")})
	t.Log(user)

	token, err := svc.GenerateAuthToken(ctx, user)
	if err != nil {
		t.Fatal(err)
	}
	if token.Prefix == "" {
		t.Fatal("expected non-empty prefix")
	}
	if token.UserID != user.ID {
		t.Fatalf("expected UserID %d, got %d", user.ID, token.UserID)
	}
	if strings.Contains(token.Secret, "$") {
		t.Fatalf("expected clean secret got %s", token.Secret)
	}

	var stored *entities.AuthToken
	for _, t := range repo.authTokens {
		stored = t
	}
	if stored == nil {
		t.Fatal("expected token to be stored")
	}
	if stored.Prefix != token.Prefix {
		t.Fatal("stored prefix should match returned prefix")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(stored.Secret), []byte(token.Secret)); err != nil {
		t.Fatal("stored secret should be hashed version of returned secret")
	}
}

func TestGenerateAuthTokenCreateError(t *testing.T) {
	repo, svc := setupAuthService(t)
	ctx := context.Background()

	repo.createAuthTokenErr = errors.New("db error")

	user, _ := repo.Create(ctx, &entities.User{Email: "err@test.com", Username: "err", Password: p("hash")})
	_, err := svc.GenerateAuthToken(ctx, user)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrCreateTokenFailed) {
		t.Fatalf("expected ErrCreateTokenFailed, got %v", err)
	}
}

func TestGenerateUserToken(t *testing.T) {
	repo, svc := setupAuthService(t)
	ctx := context.Background()

	user, _ := repo.Create(ctx, &entities.User{Email: "genutok@test.com", Username: "genutok", Password: p("hash")})

	token, err := svc.GenerateUserToken(ctx, user.ID, entities.TokenActivation)
	if err != nil {
		t.Fatal(err)
	}
	if token.Prefix == "" {
		t.Fatal("expected non-empty prefix")
	}

	var stored *entities.UserToken
	for _, t := range repo.userTokens {
		stored = t
	}
	if stored == nil {
		t.Fatal("expected token to be stored")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(stored.Secret), []byte(token.Secret)); err != nil {
		t.Fatal("stored secret should be hashed version of returned secret")
	}
}

func TestGenerateUserTokenCreateError(t *testing.T) {
	repo, svc := setupAuthService(t)
	ctx := context.Background()

	repo.createUserTokenErr = errors.New("db error")

	user, _ := repo.Create(ctx, &entities.User{Email: "utokerr@test.com", Username: "utokerr", Password: p("hash")})
	_, err := svc.GenerateUserToken(ctx, user.ID, entities.TokenActivation)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrCreateTokenFailed) {
		t.Fatalf("expected ErrCreateTokenFailed, got %v", err)
	}
}

func TestLoginWithEmail(t *testing.T) {
	repo, svc := setupAuthService(t)
	ctx := context.Background()

	hashed, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	repo.Create(ctx, &entities.User{
		Email:    "user@test.com",
		Username: "user",
		Password: p(string(hashed)),
		Active:   true,
	})

	user, bearer, err := svc.Login(ctx, "user@test.com", "password123")
	if err != nil {
		t.Fatal(err)
	}
	if user.Email != "user@test.com" {
		t.Fatalf("expected user@test.com, got %s", user.Email)
	}
	if bearer.AccessToken == "" {
		t.Fatal("expected non-empty access token")
	}
}

func TestLoginWithUsername(t *testing.T) {
	repo, svc := setupAuthService(t)
	ctx := context.Background()

	hashed, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	repo.Create(ctx, &entities.User{
		Email:    "loginuser@test.com",
		Username: "loginuser",
		Password: p(string(hashed)),
	})

	user, _, err := svc.Login(ctx, "loginuser", "password123")
	if err != nil {
		t.Fatal(err)
	}
	if user.Username != "loginuser" {
		t.Fatalf("expected 'loginuser', got '%s'", user.Username)
	}
}

func TestLoginWrongPassword(t *testing.T) {
	_, svc := setupAuthService(t)
	ctx := context.Background()

	_, _, err := svc.Login(ctx, "nonexistent@test.com", "wrong")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLoginUserNotFound(t *testing.T) {
	_, svc := setupAuthService(t)
	ctx := context.Background()

	_, _, err := svc.Login(ctx, "nobody@test.com", "password")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestRegisterSuccess(t *testing.T) {
	repo, svc := setupAuthService(t)
	ctx := context.Background()

	input := &entities.Register{
		Email:            "new@test.com",
		Username:         "newuser",
		Name:             "New User",
		Password:         "password123",
		PasswordValidate: "password123",
	}

	user, err := svc.Register(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	if user.Email != "new@test.com" {
		t.Fatalf("expected new@test.com, got %s", user.Email)
	}
	if user.Username != "newuser" {
		t.Fatalf("expected 'newuser', got '%s'", user.Username)
	}
	if user.Active {
		t.Fatal("new user should not be active")
	}
	if *user.Password == "password123" {
		t.Fatal("password should be hashed")
	}

	if len(repo.userTokens) != 1 {
		t.Fatal("expected one activation token to be created")
	}
}

func TestRegisterPasswordMismatch(t *testing.T) {
	_, svc := setupAuthService(t)
	ctx := context.Background()

	input := &entities.Register{
		Email:            "test@test.com",
		Username:         "test",
		Password:         "password123",
		PasswordValidate: "different",
	}

	_, err := svc.Register(ctx, input)
	if err == nil {
		t.Fatal("expected error for password mismatch")
	}
}

func TestRegisterDuplicateEmail(t *testing.T) {
	repo, svc := setupAuthService(t)
	ctx := context.Background()

	hashed, _ := bcrypt.GenerateFromPassword([]byte("p"), bcrypt.DefaultCost)
	repo.Create(ctx, &entities.User{Email: "dup@test.com", Username: "dup", Password: p(string(hashed))})

	input := &entities.Register{
		Email:            "dup@test.com",
		Username:         "dup2",
		Name:             "Dup",
		Password:         "password123",
		PasswordValidate: "password123",
	}

	_, err := svc.Register(ctx, input)
	if err == nil {
		t.Fatal("expected error for duplicate email")
	}
	if !errors.Is(err, ErrEmailAlreadyExists) {
		t.Fatalf("expected ErrEmailAlreadyExists, got %v", err)
	}
}

func TestRegisterDuplicateUsername(t *testing.T) {
	repo, svc := setupAuthService(t)
	ctx := context.Background()

	hashed, _ := bcrypt.GenerateFromPassword([]byte("p"), bcrypt.DefaultCost)
	repo.Create(ctx, &entities.User{Email: "first@test.com", Username: "dupuser", Password: p(string(hashed))})

	input := &entities.Register{
		Email:            "second@test.com",
		Username:         "dupuser",
		Name:             "Dup",
		Password:         "password123",
		PasswordValidate: "password123",
	}

	_, err := svc.Register(ctx, input)
	if err == nil {
		t.Fatal("expected error for duplicate username")
	}
	if !errors.Is(err, ErrUsernameAlreadyExists) {
		t.Fatalf("expected ErrUsernameAlreadyExists, got %v", err)
	}
}

func TestActivateAccountSuccess(t *testing.T) {
	repo, svc := setupAuthService(t)
	ctx := context.Background()

	user, _ := repo.Create(ctx, &entities.User{
		Email:    "activate@test.com",
		Username: "activate",
		Password: p("hashed"),
		Active:   false,
	})

	hashedSecret, _ := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	token, _ := repo.CreateUserToken(ctx, &entities.UserToken{
		UserID:    user.ID,
		Prefix:    "act-prefix",
		Secret:    string(hashedSecret),
		Type:      entities.TokenActivation,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})

	err := svc.ActivateAccount(ctx, "act-prefix.secret")
	if err != nil {
		t.Fatal(err)
	}

	updated, _ := repo.FindByID(ctx, user.ID)
	if !updated.Active {
		t.Fatal("user should be active after activation")
	}

	if !token.IsUsed {
		t.Fatal("token should be marked as used")
	}
}

func TestActivateAccountInvalidTokenFormat(t *testing.T) {
	_, svc := setupAuthService(t)
	ctx := context.Background()

	err := svc.ActivateAccount(ctx, "invalid-token")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestActivateAccountTokenNotFound(t *testing.T) {
	_, svc := setupAuthService(t)
	ctx := context.Background()

	err := svc.ActivateAccount(ctx, "valid.prefix.extra")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestActivateAccountTokenExpired(t *testing.T) {
	repo, svc := setupAuthService(t)
	ctx := context.Background()

	user, _ := repo.Create(ctx, &entities.User{
		Email:    "expired@test.com",
		Username: "expired",
		Password: p("hashed"),
	})

	prefix := "expired-prefix"
	repo.CreateUserToken(ctx, &entities.UserToken{
		UserID:    user.ID,
		Prefix:    prefix,
		Secret:    "secret",
		Type:      entities.TokenActivation,
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	})

	err := svc.ActivateAccount(ctx, prefix+".secret")
	if err == nil {
		t.Fatal("expected error for expired token")
	}
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestActivateAccountTokenUsed(t *testing.T) {
	repo, svc := setupAuthService(t)
	ctx := context.Background()

	user, _ := repo.Create(ctx, &entities.User{
		Email:    "used@test.com",
		Username: "used",
		Password: p("hashed"),
	})

	prefix := "expired-prefix"
	repo.CreateUserToken(ctx, &entities.UserToken{
		UserID:    user.ID,
		Prefix:    prefix,
		Secret:    "secret",
		Type:      entities.TokenActivation,
		ExpiresAt: time.Now().Add(1 * time.Hour),
		IsUsed:    true,
	})

	err := svc.ActivateAccount(ctx, prefix+".secret")
	if err == nil {
		t.Fatal("used token")
	}
	if !errors.Is(err, ErrUsedToken) {
		t.Fatalf("expected ErrUsedToken, got %v", err)
	}
}

func TestResendActivationWithEmail(t *testing.T) {
	repo, svc := setupAuthService(t)
	ctx := context.Background()

	hashed, _ := bcrypt.GenerateFromPassword([]byte("p"), bcrypt.DefaultCost)
	repo.Create(ctx, &entities.User{
		Email:    "resend@test.com",
		Username: "resend",
		Password: p(string(hashed)),
	})

	err := svc.ResendActivation(ctx, "resend@test.com")
	if err != nil {
		t.Fatal(err)
	}

	if len(repo.userTokens) != 1 {
		t.Fatal("expected a new activation token to be created")
	}
}

func TestResendActivationWithUsername(t *testing.T) {
	repo, svc := setupAuthService(t)
	ctx := context.Background()

	hashed, _ := bcrypt.GenerateFromPassword([]byte("p"), bcrypt.DefaultCost)
	repo.Create(ctx, &entities.User{
		Email:    "resenduser@test.com",
		Username: "resenduser",
		Password: p(string(hashed)),
	})

	err := svc.ResendActivation(ctx, "resenduser")
	if err != nil {
		t.Fatal(err)
	}
}

func TestResendActivationUserNotFound(t *testing.T) {
	_, svc := setupAuthService(t)
	ctx := context.Background()

	err := svc.ResendActivation(ctx, "nobody@test.com")
	if err == nil {
		t.Fatal("expected error for non-existent user")
	}
}

func TestForgotPasswordWithEmail(t *testing.T) {
	repo, svc := setupAuthService(t)
	ctx := context.Background()

	hashed, _ := bcrypt.GenerateFromPassword([]byte("oldpass"), bcrypt.DefaultCost)
	repo.Create(ctx, &entities.User{
		Email:    "forgot@test.com",
		Username: "forgot",
		Password: p(string(hashed)),
	})

	err := svc.ForgotPassword(ctx, "forgot@test.com")
	if err != nil {
		t.Fatal(err)
	}

	if len(repo.userTokens) != 1 {
		t.Fatal("expected a reset token to be created")
	}
}

func TestForgotPasswordWithUsername(t *testing.T) {
	repo, svc := setupAuthService(t)
	ctx := context.Background()

	hashed, _ := bcrypt.GenerateFromPassword([]byte("oldpass"), bcrypt.DefaultCost)
	repo.Create(ctx, &entities.User{
		Email:    "forgotuser@test.com",
		Username: "forgotuser",
		Password: p(string(hashed)),
	})

	err := svc.ForgotPassword(ctx, "forgotuser")
	if err != nil {
		t.Fatal(err)
	}
}

func TestForgotPasswordUserNotFound(t *testing.T) {
	_, svc := setupAuthService(t)
	ctx := context.Background()

	err := svc.ForgotPassword(ctx, "nobody@test.com")
	if err != nil {
		t.Fatal("should return nil for non-existent emails (security)")
	}
}

func TestResetPasswordSuccess(t *testing.T) {
	repo, svc := setupAuthService(t)
	ctx := context.Background()

	hashed, _ := bcrypt.GenerateFromPassword([]byte("oldpass"), bcrypt.DefaultCost)
	user, _ := repo.Create(ctx, &entities.User{
		Email:    "reset@test.com",
		Username: "reset",
		Password: p(string(hashed)),
	})

	hashedSecret, _ := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	token, _ := repo.CreateUserToken(ctx, &entities.UserToken{
		UserID:    user.ID,
		Prefix:    "reset-prefix",
		Secret:    string(hashedSecret),
		Type:      entities.TokenReset,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	})

	err := svc.ResetPassword(ctx, "reset-prefix.secret", "newpassword")
	if err != nil {
		t.Fatal(err)
	}

	updated, _ := repo.FindByID(ctx, user.ID)
	err = bcrypt.CompareHashAndPassword([]byte(*updated.Password), []byte("newpassword"))
	if err != nil {
		t.Fatal("password should be updated to newpassword")
	}

	if !token.IsUsed {
		t.Fatal("token should be marked as used")
	}
}

func TestResetPasswordInvalidTokenFormat(t *testing.T) {
	_, svc := setupAuthService(t)
	ctx := context.Background()

	err := svc.ResetPassword(ctx, "invalid", "newpass")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestResetPasswordTokenNotFound(t *testing.T) {
	_, svc := setupAuthService(t)
	ctx := context.Background()

	err := svc.ResetPassword(ctx, "valid.prefix", "newpass")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestServiceLogoutSuccess(t *testing.T) {
	repo, svc := setupAuthService(t)
	ctx := context.Background()

	user, _ := repo.Create(ctx, &entities.User{Email: "logout@test.com", Username: "logout", Password: p("hash")})

	rawSecret := "my-secret"
	hashedSecret, _ := bcrypt.GenerateFromPassword([]byte(rawSecret), bcrypt.DefaultCost)
	repo.CreateAuthToken(ctx, &entities.AuthToken{
		UserID:    user.ID,
		Prefix:    "logout-prefix",
		Secret:    string(hashedSecret),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})

	err := svc.Logout(ctx, "logout-prefix."+rawSecret)
	if err != nil {
		t.Fatal(err)
	}

	token, _ := repo.FindAuthToken(ctx, "logout-prefix")
	if !token.IsRevoked {
		t.Fatal("token should be revoked after logout")
	}
}

func TestServiceLogoutInvalidTokenFormat(t *testing.T) {
	_, svc := setupAuthService(t)
	ctx := context.Background()

	err := svc.Logout(ctx, "invalid-token")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestServiceLogoutTokenNotFound(t *testing.T) {
	_, svc := setupAuthService(t)
	ctx := context.Background()

	err := svc.Logout(ctx, "nonexistent.secret")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestServiceLogoutSecretMismatch(t *testing.T) {
	repo, svc := setupAuthService(t)
	ctx := context.Background()

	user, _ := repo.Create(ctx, &entities.User{Email: "logouterr@test.com", Username: "logouterr", Password: p("hash")})

	hashedSecret, _ := bcrypt.GenerateFromPassword([]byte("real-secret"), bcrypt.DefaultCost)
	repo.CreateAuthToken(ctx, &entities.AuthToken{
		UserID:    user.ID,
		Prefix:    "logout-prefix",
		Secret:    string(hashedSecret),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})

	err := svc.Logout(ctx, "logout-prefix.wrong-secret")
	if err == nil {
		t.Fatal("expected error for wrong secret")
	}
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestServiceRefreshTokenSuccess(t *testing.T) {
	repo, svc := setupAuthService(t)
	ctx := context.Background()

	user, _ := repo.Create(ctx, &entities.User{Email: "refresh@test.com", Username: "refresh", Password: p("hash")})

	rawRefreshSecret := "my-refresh-secret"
	hashedRefresh, _ := bcrypt.GenerateFromPassword([]byte(rawRefreshSecret), bcrypt.DefaultCost)
	repo.CreateAuthToken(ctx, &entities.AuthToken{
		UserID:           user.ID,
		Prefix:           "refresh-prefix",
		Secret:           "$2a$10$old-hashed-secret",
		RefreshSecret:    string(hashedRefresh),
		ExpiresAt:        time.Now().Add(-1 * time.Hour),
		RefreshExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	})

	bearer, err := svc.RefreshToken(ctx, "refresh-prefix."+rawRefreshSecret)
	if err != nil {
		t.Fatal(err)
	}
	if bearer.AccessToken == "" {
		t.Fatal("expected non-empty access token")
	}
	if bearer.RefreshToken == "" {
		t.Fatal("expected non-empty refresh token")
	}
}

func TestServiceRefreshTokenInvalidFormat(t *testing.T) {
	_, svc := setupAuthService(t)
	ctx := context.Background()

	_, err := svc.RefreshToken(ctx, "invalid")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestServiceRefreshTokenNotFound(t *testing.T) {
	_, svc := setupAuthService(t)
	ctx := context.Background()

	_, err := svc.RefreshToken(ctx, "nonexistent.secret")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestServiceRefreshTokenRevoked(t *testing.T) {
	repo, svc := setupAuthService(t)
	ctx := context.Background()

	user, _ := repo.Create(ctx, &entities.User{Email: "revref@test.com", Username: "revref", Password: p("hash")})
	now := time.Now()
	repo.CreateAuthToken(ctx, &entities.AuthToken{
		UserID:           user.ID,
		Prefix:           "revoked-ref",
		Secret:           "secret",
		RefreshSecret:    "refresh",
		ExpiresAt:        time.Now().Add(24 * time.Hour),
		RefreshExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		IsRevoked:        true,
		RevokedAt:        &now,
	})

	_, err := svc.RefreshToken(ctx, "revoked-ref.secret")
	if err == nil {
		t.Fatal("expected error for revoked token")
	}
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestServiceRefreshTokenExpiredRefresh(t *testing.T) {
	repo, svc := setupAuthService(t)
	ctx := context.Background()

	user, _ := repo.Create(ctx, &entities.User{Email: "expref@test.com", Username: "expref", Password: p("hash")})
	repo.CreateAuthToken(ctx, &entities.AuthToken{
		UserID:           user.ID,
		Prefix:           "expired-ref",
		Secret:           "secret",
		RefreshSecret:    "refresh",
		ExpiresAt:        time.Now().Add(24 * time.Hour),
		RefreshExpiresAt: time.Now().Add(-1 * time.Hour),
	})

	_, err := svc.RefreshToken(ctx, "expired-ref.secret")
	if err == nil {
		t.Fatal("expected error for expired refresh token")
	}
	if !errors.Is(err, ErrRefreshTokenExpired) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestServiceRefreshTokenWrongSecret(t *testing.T) {
	repo, svc := setupAuthService(t)
	ctx := context.Background()

	user, _ := repo.Create(ctx, &entities.User{Email: "wrongref@test.com", Username: "wrongref", Password: p("hash")})
	hashedRefresh, _ := bcrypt.GenerateFromPassword([]byte("real-refresh"), bcrypt.DefaultCost)
	repo.CreateAuthToken(ctx, &entities.AuthToken{
		UserID:           user.ID,
		Prefix:           "wrong-ref",
		Secret:           "secret",
		RefreshSecret:    string(hashedRefresh),
		ExpiresAt:        time.Now().Add(-1 * time.Hour),
		RefreshExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	})

	_, err := svc.RefreshToken(ctx, "wrong-ref.wrong-refresh-secret")
	if err == nil {
		t.Fatal("expected error for wrong refresh secret")
	}
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestResetPasswordTokenExpired(t *testing.T) {
	repo, svc := setupAuthService(t)
	ctx := context.Background()

	user, _ := repo.Create(ctx, &entities.User{
		Email:    "resetexp@test.com",
		Username: "resetexp",
		Password: p("hashed"),
	})

	prefix := "reset-exp-prefix"
	repo.CreateUserToken(ctx, &entities.UserToken{
		UserID:    user.ID,
		Prefix:    prefix,
		Secret:    "secret",
		Type:      entities.TokenReset,
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	})

	err := svc.ResetPassword(ctx, prefix+".secret", "newpass")
	if err == nil {
		t.Fatal("expected error for expired token")
	}
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}
func p(s string) *string { return &s }
