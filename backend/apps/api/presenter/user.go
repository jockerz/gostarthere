package presenter

import "vnti/pkg/entities"

type UserResponse struct {
	ID       uint   `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Avatar   string `json:"avatar"`
	Active   bool   `json:"active"`
}

type BearerTokenResponse struct{}

type AuthResponse struct {
	Token *entities.BearerToken `json:"token"`
	User  UserResponse          `json:"user"`
}

type UpdateProfileRequest struct {
	Name     string `json:"name,omitempty"`
	Username string `json:"username,omitempty"`
	Avatar   string `json:"avatar,omitempty"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" minLength:"8" maxLength:"48"`
	NewPassword     string `json:"new_password" minLength:"8" maxLength:"48"`
}

type EmailUpdateRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=48"`
}
