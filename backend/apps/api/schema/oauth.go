package schema

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

var OAuthTags = []string{"OAuth"}

type OAuthAuthorizeInput struct {
	Provider      string `path:"provider" enum:"google,github"`
	CodeChallenge string `query:"code_challenge"`
	Authorization string `header:"Authorization"`
}

type OAuthAuthorizeOutput struct {
	Body struct {
		URL   string `json:"url"`
		State string `json:"state"`
	}
}

var OAuthAuthorizeOp = huma.Operation{
	OperationID: "oauth-authorize",
	Method:      http.MethodGet,
	Path:        "/auth/oauth/{provider}/authorize",
	Summary:     "Get OAuth2 authorize URL",
	Tags:        OAuthTags,
}

type OAuthCallbackInput struct {
	Provider      string `path:"provider" enum:"google,github"`
	Authorization string `header:"Authorization"`
	Body          struct {
		Code         string `json:"code"`
		State        string `json:"state"`
		CodeVerifier string `json:"code_verifier,omitempty"`
	}
}

type OAuthCallbackOutput struct {
	Body struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
		Data    any    `json:"data,omitempty"`
	}
}

var OAuthCallbackOp = huma.Operation{
	OperationID: "oauth-callback",
	Method:      http.MethodPost,
	Path:        "/auth/oauth/{provider}/callback",
	Summary:     "Exchange OAuth2 code for JWT",
	Tags:        OAuthTags,
}

type SetPasswordInput struct {
	Body struct {
		Password string `json:"password" minLength:"8" maxLength:"48"`
	}
}

var SetPasswordOp = huma.Operation{
	OperationID: "set-password",
	Method:      http.MethodPost,
	Path:        "/auth/set-password",
	Summary:     "Set password for OAuth-only account",
	Tags:        OAuthTags,
}
