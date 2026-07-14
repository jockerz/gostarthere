package entities

import "time"

type AuthProvider string

const (
	AuthProviderGoogle AuthProvider = "google"
	AuthProviderGitHub AuthProvider = "github"
)

type UserAuthProvider struct {
	ID             uint         `gorm:"primaryKey" json:"id"`
	UserID         uint         `gorm:"not null;index" json:"user_id"`
	User           User         `gorm:"foreignKey:UserID" json:"-"`
	Provider       AuthProvider `gorm:"type:text;not null;check:provider IN ('google','github')" json:"provider"`
	ProviderUserID string       `gorm:"not null" json:"provider_user_id"`
	Email          string       `gorm:"not null" json:"email"`
	CreatedAt      time.Time    `json:"created_at"`
}

func (UserAuthProvider) TableName() string {
	return "user_auth_providers"
}
