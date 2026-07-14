package presenter

import "vnti/pkg/entities"

type LoginRequest struct {
	Email    string `json:"email" minLength:"5" maxLength:"255"`
	Password string `json:"password" minLength:"6" maxLength:"48"`
	Remember bool   `json:"remember,omitempty"`
}

type RegisterRequest struct {
	Email            string `json:"email" format:"email"`
	Username         string `json:"username" minLength:"5" maxLength:"48" pattern:"[A-Za-z0-9_]+"`
	Password         string `json:"password" minLength:"10" maxLength:"48"`
	PasswordValidate string `json:"password_validate" dependentRequired:"password" minLength:"9" maxLength:"48"`
	Name             string `json:"name" minLength:"3" maxLength:"127"`
}

type AuthenticationData struct {
	AuthProviderData []*entities.UserAuthProvider `json:"data"`
	HasPassword      bool                         `json:"has_password"`
}
