package entities

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const JWTIssuer = "BE - Vnti"

type Login struct {
	Email    string `json:"email" `
	Password string `json:"password"`
}

type Register struct {
	Email            string `json:"email" maxLength:"127"`
	Username         string `json:"username" validate:"required,min=6,max=32"`
	Name             string `json:"name" validate:"required,min=3,max=127"`
	Password         string `json:"password" validate:"required,min=9,max=127"`
	PasswordValidate string `json:"validate" validate:"required"`
}

type BearerToken struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
	Scope        string    `json:"scope"`
}

// Token for authentication
type AuthToken struct {
	ID               uint      `gorm:"primaryKey"`
	UserID           uint      `gorm:"index;not null"`
	Prefix           string    `gorm:"uniqueIndex;not null;size:48"`
	Secret           string    `gorm:"not null;size:127"`
	RefreshSecret    string    `gorm:"size:127"`
	ExpiresAt        time.Time `gorm:"not null"`
	RefreshedAt      time.Time
	RefreshExpiresAt time.Time
	IsRevoked        bool `gorm:"not null;default:false"`
	RevokedAt        *time.Time
	CreatedAt        time.Time `json:"created_at"`
	CreatedBy        uint
}

func (AuthToken) TableName() string {
	return "auth_token"
}

// func (auth_token *AuthToken) ToBearerToken(signineKey string) BearerToken {
func (auth_token *AuthToken) ToBearerToken(jwtSecret string) BearerToken {
	var refresh_token string

	accessToken := AsSignedAccessToken(auth_token, jwtSecret)
	if strings.TrimSpace(auth_token.RefreshSecret) != "" {
		refresh_token = AsSigndeRefreshToken(auth_token, jwtSecret)
	}

	return BearerToken{
		AccessToken:  accessToken,
		RefreshToken: refresh_token,
		ExpiresIn:    int(time.Until(auth_token.ExpiresAt).Seconds()),
		ExpiresAt:    auth_token.ExpiresAt,
		Scope:        "user",
		TokenType:    "Bearer",
	}
}

func (token *AuthToken) AsTokenWithSecret(secret string) string {
	return fmt.Sprintf("%s.%s", token.Prefix, secret)
}

type JWTClaims struct {
	ID string `json:"id"`
	jwt.RegisteredClaims
}

func asSignedToken(signingKey string, tokenId string, created, expired time.Time) string {
	claims := JWTClaims{
		ID: tokenId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expired),
			IssuedAt:  jwt.NewNumericDate(created),
			Issuer:    JWTIssuer,
			// Bisa masalah urusan TZ
			// NotBefore: jwt.NewNumericDate(created),
		},
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := jwtToken.SignedString([]byte(signingKey))
	if err != nil {
		panic(err)
	}
	return signed
}

// Use token with unhased secret here
func AsSignedAccessToken(token *AuthToken, signingKey string) string {
	return asSignedToken(signingKey, token.AsTokenWithSecret(token.Secret), token.CreatedAt, token.ExpiresAt)
}

// Use token with unhased refresh secret here
func AsSigndeRefreshToken(token *AuthToken, signingKey string) string {
	if token.RefreshSecret == "" {
		return ""
	}
	return asSignedToken(signingKey, token.AsTokenWithSecret(token.RefreshSecret), token.CreatedAt, token.ExpiresAt)
}

type UserTokenType string

const (
	TokenActivation  UserTokenType = "activation"
	TokenReset       UserTokenType = "reset_password"
	TokenEmailUpdate UserTokenType = "email_update"
)

// Token for activation, password reset, etc
type UserToken struct {
	ID            uint          `gorm:"primaryKey"`
	UserID        uint          `gorm:"index;not null"`
	Prefix        string        `gorm:"uniqueIndex;not null;size:48"`
	Secret        string        `gorm:"not null;size:127"`
	RefreshSecret string        `gorm:"size:127"`
	Type          UserTokenType `gorm:"index;size:50;not null"`
	// NewEmail         string        `gorm:"size:255"`
	Data             string    `gorm:"size:255"`
	ExpiresAt        time.Time `gorm:"not null"`
	RefreshedAt      time.Time
	RefreshExpiresAt time.Time
	IsUsed           bool `gorm:"index;default:false"`
	UsedAt           *time.Time
	IsRevoked        bool `gorm:"not null;default:false"`
	RevokedAt        *time.Time
	CreatedAt        time.Time `json:"created_at"`
	CreatedBy        uint
}

func (UserToken) TableName() string {
	return "user_token"
}

func (token *UserToken) AsTokenWithSecret(secret string) string {
	return fmt.Sprintf("%s.%s", token.Prefix, secret)
}
