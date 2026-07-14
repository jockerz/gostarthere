package schema

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"vnti/apps/api/presenter"
)

var userTags = []string{"User"}
var authScheme = []map[string][]string{
	{"bearer": {"user"}},
}

type EmailUpdateRequestInput struct {
	Body presenter.EmailUpdateRequest
}

type EmailUpdateConfirmInput struct {
	Token string `path:"token"`
}

var EmailUpdateRequestOp = huma.Operation{
	Path:        "/auth/update-email",
	Method:      http.MethodPost,
	Security:    authScheme,
	Tags:        userTags,
	Summary:     "Request email update",
	Description: "Verify password and send confirmation to new email",
}

var EmailUpdateConfirmOp = huma.Operation{
	Path:        "/auth/update-email/{token}/confirm",
	Method:      http.MethodPost,
	Security:    authScheme,
	Tags:        userTags,
	Summary:     "Confirm email update",
	Description: "Confirm email change via confirmation token",
}

type GetProfileInput struct{}

type UpdateProfileInput struct {
	Body presenter.UpdateProfileRequest
}

var GetProfile = huma.Operation{
	Path:        "/profile",
	Summary:     "Profile",
	Description: "Get profile data",
	Method:      http.MethodGet,
	Security:    authScheme,
	Tags:        userTags,
	Responses: map[string]*huma.Response{
		"200": {
			Description: "OK",
		},
	},
}

var UpdateProfile = huma.Operation{
	Path:        "/profile",
	Method:      http.MethodPut,
	Security:    authScheme,
	Tags:        userTags,
	Summary:     "Update profile",
	Description: "Update user profile information",
}

type ChangePasswordInput struct {
	Body presenter.ChangePasswordRequest
}

var ChangePasswordOp = huma.Operation{
	Path:        "/auth/change-password",
	Method:      http.MethodPut,
	Security:    authScheme,
	Tags:        userTags,
	Summary:     "Change password",
	Description: "Change current password by providing current password",
}

type UploadAvatarInput struct {
	RawBody huma.MultipartFormFiles[struct {
		Avatar huma.FormFile `form:"avatar" contentType:"image/jpeg,image/png,image/gif,image/webp" required:"true"`
	}]
}

var UploadAvatarOp = huma.Operation{
	Path:        "/profile/avatar",
	Method:      http.MethodPost,
	Security:    authScheme,
	Tags:        userTags,
	Summary:     "Upload avatar",
	Description: "Upload user avatar image",
}
