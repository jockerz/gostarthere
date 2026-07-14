package schema

import (
	"net/http"

	"vnti/apps/api/presenter"

	"github.com/danielgtaylor/huma/v2"
)

var authTags = []string{"Authentication"}

type LoginInput struct {
	Body presenter.LoginRequest
}

type RegisterInput struct {
	Body presenter.RegisterRequest
}

type ActivateInput struct {
	// Body struct {
	// 	Token string `json:"token"`
	// }
	Token string `path:"token"`
}

type ResetActivationInput struct {
	Body struct {
		Email string `json:"email" minLength:"5" maxLength:"255"`
	}
}

type ForgotPasswordInput struct {
	Body struct {
		Email string `json:"email" minLength:"5" maxLength:"255"`
	}
}

type ResetPasswordInput struct {
	Body struct {
		Token    string `json:"token" minLength:"1"`
		Password string `json:"password" minLength:"8" maxLength:"48"`
	}
}

type LogoutInput struct {
	Body struct {
		Token string `json:"token" minLength:"1"`
	}
}

type GetAuthProviderInput struct{}

var LogoutOp = huma.Operation{
	Path:        "/auth/logout",
	Method:      http.MethodPost,
	Tags:        authTags,
	Summary:     "Logout",
	Description: "Revoke an auth token",
}

type RefreshTokenInput struct {
	Body struct {
		RefreshToken string `json:"refresh_token" minLength:"1"`
	}
}

var RefreshTokenOp = huma.Operation{
	Path:        "/auth/refresh",
	Method:      http.MethodPost,
	Tags:        authTags,
	Summary:     "Refresh token",
	Description: "Refresh an expired auth token using a refresh token",
}

var LoginOp = huma.Operation{
	Path:        "/auth/login",
	Method:      http.MethodPost,
	Tags:        authTags,
	Summary:     "Login",
	Description: "Login with email and password",
}

var RegisterOp = huma.Operation{
	Path:        "/auth/register",
	Method:      http.MethodPost,
	Tags:        authTags,
	Summary:     "Register",
	Description: "Register a new account",
}

var ActivateOp = huma.Operation{
	Path:        "/auth/activate/{token}",
	Method:      http.MethodPost,
	Tags:        authTags,
	Summary:     "Activate account",
	Description: "Activate account using activation token",
}

var ResendActivationOp = huma.Operation{
	Path:        "/auth/resend-activation",
	Method:      http.MethodPost,
	Tags:        authTags,
	Summary:     "Reset activation",
	Description: "Request a new activation token",
}

var ForgotPasswordOp = huma.Operation{
	Path:        "/auth/forgot-password",
	Method:      http.MethodPost,
	Tags:        authTags,
	Summary:     "Forgot password",
	Description: "Request a password reset token",
}

var ResetPasswordOp = huma.Operation{
	Path:        "/auth/reset-password",
	Method:      http.MethodPost,
	Tags:        authTags,
	Summary:     "Reset password",
	Description: "Reset password using reset token",
}

var GetAuthProviderOp = huma.Operation{
	Path:        "/auth/auth_data",
	Method:      http.MethodGet,
	Tags:        authTags,
	Summary:     "Get User's authentication data",
	Description: "Authentication data which if the user has password or not and third party authentication data",
}
