package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/stretchr/testify/assert"

	"vnti/apps/api/schema"
	"vnti/pkg/entities"
	"vnti/pkg/service"
)

type mockAuthSvc struct {
	loginFunc                  func(ctx context.Context, email, password string) (*entities.User, *entities.BearerToken, error)
	registerFunc               func(ctx context.Context, input *entities.Register) (*entities.User, error)
	activateAccountFunc        func(ctx context.Context, token string) error
	resendActivationFunc       func(ctx context.Context, email string) error
	forgotPasswordFunc         func(ctx context.Context, emailOrUsername string) error
	resetPasswordFunc          func(ctx context.Context, token, newPassword string) error
	logoutFunc                 func(ctx context.Context, token string) error
	refreshTokenFunc           func(ctx context.Context, refreshToken string) (*entities.BearerToken, error)
	getAuthTokenByJWTTokenFunc func(ctx context.Context, token string) (*entities.AuthToken, error)
	authByTokenFunc            func(ctx context.Context, token string) (*entities.User, error)
	authByRefreshTokenFunc     func(ctx context.Context, token string) (*entities.User, error)
	setSkipTaskQueueFunc       func(v bool)
}

func newMockAuthSvc() *mockAuthSvc {
	return &mockAuthSvc{
		loginFunc: func(_ context.Context, email, password string) (*entities.User, *entities.BearerToken, error) {
			return &entities.User{
					ID:       1,
					Email:    email,
					Username: "user",
					Name:     "User",
					Active:   true,
				}, &entities.BearerToken{
					AccessToken:  "access-token",
					RefreshToken: "refresh-token",
					TokenType:    "Bearer",
					ExpiresIn:    86400,
					ExpiresAt:    time.Now().Add(24 * time.Hour),
				}, nil
		},
		registerFunc: func(_ context.Context, input *entities.Register) (*entities.User, error) {
			return &entities.User{
				ID:       1,
				Email:    input.Email,
				Username: input.Username,
				Name:     input.Name,
				Active:   false,
			}, nil
		},
		activateAccountFunc:  func(_ context.Context, _ string) error { return nil },
		resendActivationFunc: func(_ context.Context, _ string) error { return nil },
		forgotPasswordFunc:   func(_ context.Context, _ string) error { return nil },
		resetPasswordFunc:    func(_ context.Context, _, _ string) error { return nil },
		logoutFunc:           func(_ context.Context, _ string) error { return nil },
		refreshTokenFunc: func(_ context.Context, _ string) (*entities.BearerToken, error) {
			return &entities.BearerToken{
				AccessToken:  "new-access-token",
				RefreshToken: "new-refresh-token",
				TokenType:    "Bearer",
				ExpiresIn:    86400,
				ExpiresAt:    time.Now().Add(24 * time.Hour),
			}, nil
		},
		getAuthTokenByJWTTokenFunc: func(_ context.Context, _ string) (*entities.AuthToken, error) {
			return &entities.AuthToken{
				ID:     1,
				UserID: 1,
				Prefix: "prefix",
				Secret: "secret",
			}, nil
		},
		authByTokenFunc: func(_ context.Context, _ string) (*entities.User, error) {
			return &entities.User{
				ID:       1,
				Email:    "user@test.com",
				Username: "user",
				Name:     "User",
				Active:   true,
			}, nil
		},
		authByRefreshTokenFunc: func(_ context.Context, _ string) (*entities.User, error) {
			return &entities.User{
				ID:       1,
				Email:    "user@test.com",
				Username: "user",
				Name:     "User",
				Active:   true,
			}, nil
		},
	}
}

func (m *mockAuthSvc) Login(ctx context.Context, email, password string) (*entities.User, *entities.BearerToken, error) {
	return m.loginFunc(ctx, email, password)
}

func (m *mockAuthSvc) Register(ctx context.Context, input *entities.Register) (*entities.User, error) {
	return m.registerFunc(ctx, input)
}

func (m *mockAuthSvc) ActivateAccount(ctx context.Context, token string) error {
	return m.activateAccountFunc(ctx, token)
}

func (m *mockAuthSvc) ResendActivation(ctx context.Context, email string) error {
	return m.resendActivationFunc(ctx, email)
}

func (m *mockAuthSvc) ForgotPassword(ctx context.Context, emailOrUsername string) error {
	return m.forgotPasswordFunc(ctx, emailOrUsername)
}

func (m *mockAuthSvc) ResetPassword(ctx context.Context, token, newPassword string) error {
	return m.resetPasswordFunc(ctx, token, newPassword)
}

func (m *mockAuthSvc) Logout(ctx context.Context, token string) error {
	return m.logoutFunc(ctx, token)
}

func (m *mockAuthSvc) RefreshToken(ctx context.Context, refreshToken string) (*entities.BearerToken, error) {
	return m.refreshTokenFunc(ctx, refreshToken)
}

func (m *mockAuthSvc) GenerateUserToken(_ context.Context, _ uint, _ entities.UserTokenType) (*entities.UserToken, error) {
	return &entities.UserToken{}, nil
}

func (m *mockAuthSvc) GenerateAuthToken(_ context.Context, _ *entities.User) (*entities.AuthToken, error) {
	return &entities.AuthToken{}, nil
}

func (m *mockAuthSvc) GetAuthTokenByJWTToken(ctx context.Context, token string) (*entities.AuthToken, error) {
	return m.getAuthTokenByJWTTokenFunc(ctx, token)
}

func (m *mockAuthSvc) AuthByToken(ctx context.Context, token string) (*entities.User, error) {
	return m.authByTokenFunc(ctx, token)
}

func (m *mockAuthSvc) AuthByRefreshToken(ctx context.Context, token string) (*entities.User, error) {
	return m.authByTokenFunc(ctx, token)
}

func (m *mockAuthSvc) SetSkipTaskQueue(_ bool) {}

func setupAuthHandlerTest(t *testing.T) (humatest.TestAPI, *mockAuthSvc) {
	t.Helper()
	_, api := humatest.New(t)
	svc := newMockAuthSvc()

	huma.Register(api, schema.LogoutOp, Logout(svc))
	huma.Register(api, schema.RefreshTokenOp, RefreshToken(svc))
	huma.Register(api, schema.LoginOp, Login(svc))
	huma.Register(api, schema.RegisterOp, Register(svc))
	huma.Register(api, schema.ActivateOp, Activate(svc))
	huma.Register(api, schema.ResendActivationOp, ResendActivation(svc))
	huma.Register(api, schema.ForgotPasswordOp, ForgotPassword(svc))
	huma.Register(api, schema.ResetPasswordOp, ResetPassword(svc))

	return humatest.Wrap(t, api), svc
}

func TestLoginHandlerSuccess(t *testing.T) {
	api, _ := setupAuthHandlerTest(t)

	resp := api.Post("/auth/login", map[string]any{
		"email":    "user@test.com",
		"password": "password123",
	})

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestLoginHandlerInvalidCredentials(t *testing.T) {
	api, svc := setupAuthHandlerTest(t)
	svc.loginFunc = func(_ context.Context, _, _ string) (*entities.User, *entities.BearerToken, error) {
		return nil, nil, service.ErrAuthInvalidCredentials
	}

	resp := api.Post("/auth/login", map[string]any{
		"email":    "wrong@test.com",
		"password": "wrongpass",
	})

	assert.Equal(t, http.StatusUnauthorized, resp.Code)
}

func TestLoginHandlerInternalError(t *testing.T) {
	api, svc := setupAuthHandlerTest(t)
	svc.loginFunc = func(_ context.Context, _, _ string) (*entities.User, *entities.BearerToken, error) {
		return nil, nil, errors.New("unexpected error")
	}

	resp := api.Post("/auth/login", map[string]any{
		"email":    "user@test.com",
		"password": "password123",
	})

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
}

func TestRegisterHandlerSuccess(t *testing.T) {
	api, _ := setupAuthHandlerTest(t)

	resp := api.Post("/auth/register", map[string]any{
		"email":             "new@test.com",
		"username":          "newuser",
		"password":          "password123",
		"password_validate": "password123",
		"name":              "New User",
	})

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestRegisterHandlerPasswordMismatch(t *testing.T) {
	api, svc := setupAuthHandlerTest(t)
	svc.registerFunc = func(_ context.Context, _ *entities.Register) (*entities.User, error) {
		return nil, errors.New("passwords do not match")
	}

	resp := api.Post("/auth/register", map[string]any{
		"email":             "test@test.com",
		"username":          "testuser",
		"password":          "password123",
		"password_validate": "different",
		"name":              "Test",
	})

	assert.Equal(t, http.StatusUnprocessableEntity, resp.Code)
}

func TestRegisterHandlerDuplicateEmail(t *testing.T) {
	api, svc := setupAuthHandlerTest(t)
	svc.registerFunc = func(_ context.Context, _ *entities.Register) (*entities.User, error) {
		return nil, service.ErrAuthEmailAlreadyExists
	}

	resp := api.Post("/auth/register", map[string]any{
		"email":             "dup@test.com",
		"username":          "dupuser",
		"password":          "password123",
		"password_validate": "password123",
		"name":              "Dup",
	})

	assert.Equal(t, http.StatusConflict, resp.Code)
}

func TestRegisterHandlerDuplicateUsername(t *testing.T) {
	api, svc := setupAuthHandlerTest(t)
	svc.registerFunc = func(_ context.Context, _ *entities.Register) (*entities.User, error) {
		return nil, service.ErrAuthUsernameAlreadyExists
	}

	resp := api.Post("/auth/register", map[string]any{
		"email":             "other@test.com",
		"username":          "dupuser",
		"password":          "password123",
		"password_validate": "password123",
		"name":              "Dup",
	})

	assert.Equal(t, http.StatusConflict, resp.Code)
}

func TestActivateHandlerSuccess(t *testing.T) {
	api, _ := setupAuthHandlerTest(t)

	resp := api.Post("/auth/activate/prefix.secret")

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestActivateHandlerInvalidToken(t *testing.T) {
	api, svc := setupAuthHandlerTest(t)
	svc.activateAccountFunc = func(_ context.Context, _ string) error {
		return service.ErrAuthInvalidToken
	}

	resp := api.Post("/auth/activate/invalid")

	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestActivateHandlerInternalError(t *testing.T) {
	api, svc := setupAuthHandlerTest(t)
	svc.activateAccountFunc = func(_ context.Context, _ string) error {
		return errors.New("unexpected error")
	}

	resp := api.Post("/auth/activate/prefix.secret")

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
}

func TestResetActivationHandlerSuccess(t *testing.T) {
	api, _ := setupAuthHandlerTest(t)

	resp := api.Post("/auth/resend-activation", map[string]any{
		"email": "user@test.com",
	})

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestResetActivationHandlerInternalError(t *testing.T) {
	api, svc := setupAuthHandlerTest(t)
	svc.resendActivationFunc = func(_ context.Context, _ string) error {
		return errors.New("unexpected error")
	}

	resp := api.Post("/auth/resend-activation", map[string]any{
		"email": "user@test.com",
	})

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestForgotPasswordHandlerSuccess(t *testing.T) {
	api, _ := setupAuthHandlerTest(t)

	resp := api.Post("/auth/forgot-password", map[string]any{
		"email": "user@test.com",
	})

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestForgotPasswordHandlerInternalError(t *testing.T) {
	api, svc := setupAuthHandlerTest(t)
	svc.forgotPasswordFunc = func(_ context.Context, _ string) error {
		return errors.New("unexpected error")
	}

	resp := api.Post("/auth/forgot-password", map[string]any{
		"email": "user@test.com",
	})

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
}

func TestResetPasswordHandlerSuccess(t *testing.T) {
	api, _ := setupAuthHandlerTest(t)

	resp := api.Post("/auth/reset-password", map[string]any{
		"token":    "prefix.secret",
		"password": "newpassword",
	})

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestResetPasswordHandlerInvalidToken(t *testing.T) {
	api, svc := setupAuthHandlerTest(t)
	svc.resetPasswordFunc = func(_ context.Context, _, _ string) error {
		return service.ErrAuthInvalidToken
	}

	resp := api.Post("/auth/reset-password", map[string]any{
		"token":    "invalid",
		"password": "newpassword",
	})

	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestResetPasswordHandlerInternalError(t *testing.T) {
	api, svc := setupAuthHandlerTest(t)
	svc.resetPasswordFunc = func(_ context.Context, _, _ string) error {
		return errors.New("unexpected error")
	}

	resp := api.Post("/auth/reset-password", map[string]any{
		"token":    "prefix.secret",
		"password": "newpassword",
	})

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
}

func TestLogoutHandlerSuccess(t *testing.T) {
	api, _ := setupAuthHandlerTest(t)

	resp := api.Post("/auth/logout", map[string]any{
		"token": "prefix.secret",
	})

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestLogoutHandlerInvalidToken(t *testing.T) {
	api, svc := setupAuthHandlerTest(t)
	svc.logoutFunc = func(_ context.Context, _ string) error {
		return service.ErrAuthInvalidToken
	}

	resp := api.Post("/auth/logout", map[string]any{
		"token": "invalid",
	})

	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestLogoutHandlerInternalError(t *testing.T) {
	api, svc := setupAuthHandlerTest(t)
	svc.logoutFunc = func(_ context.Context, _ string) error {
		return errors.New("unexpected error")
	}

	resp := api.Post("/auth/logout", map[string]any{
		"token": "prefix.secret",
	})

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
}

func TestRefreshTokenHandlerSuccess(t *testing.T) {
	api, _ := setupAuthHandlerTest(t)

	resp := api.Post("/auth/refresh", map[string]any{
		"refresh_token": "prefix.secret",
	})

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestRefreshTokenHandlerInvalidToken(t *testing.T) {
	api, svc := setupAuthHandlerTest(t)
	svc.refreshTokenFunc = func(_ context.Context, _ string) (*entities.BearerToken, error) {
		return nil, service.ErrAuthInvalidToken
	}

	resp := api.Post("/auth/refresh", map[string]any{
		"refresh_token": "invalid",
	})

	assert.Equal(t, http.StatusUnauthorized, resp.Code)
}

func TestMockGetAuthTokenByJWTTokenSuccess(t *testing.T) {
	_, svc := setupAuthHandlerTest(t)

	token, err := svc.GetAuthTokenByJWTToken(context.Background(), "prefix.secret")
	assert.NoError(t, err)
	assert.Equal(t, uint(1), token.UserID)
	assert.Equal(t, "prefix", token.Prefix)
}

func TestMockGetAuthTokenByJWTTokenError(t *testing.T) {
	_, svc := setupAuthHandlerTest(t)
	svc.getAuthTokenByJWTTokenFunc = func(_ context.Context, _ string) (*entities.AuthToken, error) {
		return nil, service.ErrAuthInvalidToken
	}

	token, err := svc.GetAuthTokenByJWTToken(context.Background(), "invalid")
	assert.Error(t, err)
	assert.Nil(t, token)
	assert.ErrorIs(t, err, service.ErrAuthInvalidToken)
}

func TestMockAuthByTokenWithSecretSuccess(t *testing.T) {
	_, svc := setupAuthHandlerTest(t)

	user, err := svc.AuthByToken(context.Background(), "prefix.secret")
	assert.NoError(t, err)
	assert.Equal(t, "user@test.com", user.Email)
	assert.True(t, user.Active)
}

func TestMockAuthByTokenWithSecretError(t *testing.T) {
	_, svc := setupAuthHandlerTest(t)
	svc.authByTokenFunc = func(_ context.Context, _ string) (*entities.User, error) {
		return nil, service.ErrAuthInvalidToken
	}

	user, err := svc.AuthByToken(context.Background(), "invalid")
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.ErrorIs(t, err, service.ErrAuthInvalidToken)
}

func TestMockAuthByTokenWithSecretInactiveUser(t *testing.T) {
	_, svc := setupAuthHandlerTest(t)
	svc.authByTokenFunc = func(_ context.Context, _ string) (*entities.User, error) {
		return &entities.User{
			ID:       1,
			Email:    "inactive@test.com",
			Username: "inactive",
			Name:     "Inactive",
			Active:   false,
		}, nil
	}

	user, err := svc.AuthByToken(context.Background(), "prefix.secret")
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.False(t, user.Active)
}

func TestRefreshTokenHandlerInternalError(t *testing.T) {
	api, svc := setupAuthHandlerTest(t)
	svc.refreshTokenFunc = func(_ context.Context, _ string) (*entities.BearerToken, error) {
		return nil, errors.New("unexpected error")
	}

	resp := api.Post("/auth/refresh", map[string]any{
		"refresh_token": "prefix.secret",
	})

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
}
