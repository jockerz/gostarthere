package repository

import (
	"context"
	"time"

	"vnti/pkg/entities"

	"gorm.io/gorm"
)

type authTokenRepo interface {
	CreateAuthToken(ctx context.Context, token *entities.AuthToken) (*entities.AuthToken, error)
	FindAuthToken(ctx context.Context, prefix string) (*entities.AuthToken, error)
	RefreshAuthToken(ctx context.Context, token *entities.AuthToken) error
	RevokeAuthToken(ctx context.Context, prefix string) error
}

type userTokenRepo interface {
	CreateUserToken(ctx context.Context, token *entities.UserToken) (*entities.UserToken, error)
	FindUserToken(ctx context.Context, prefix string, tokenType entities.UserTokenType) (*entities.UserToken, error)
	FindUserTokenByUserId(context.Context, uint, entities.UserTokenType) (*[]entities.UserToken, error)
	RefreshUserToken(ctx context.Context, token *entities.UserToken) error
	RevokeUserToken(ctx context.Context, prefix string) error
	MarkAsUsedUserToken(ctx context.Context, tokenId uint) error
}

type AuthRepository interface {
	authTokenRepo
	userTokenRepo
}

type authRepositoryImpl struct {
	DB *gorm.DB
}

func NewAuthRepository(db *gorm.DB) AuthRepository {
	return &authRepositoryImpl{DB: db}
}

func (r *authRepositoryImpl) CreateAuthToken(ctx context.Context, token *entities.AuthToken) (*entities.AuthToken, error) {
	err := r.DB.WithContext(ctx).Create(token).Error
	if err != nil {
		return nil, err
	}
	return token, err
}

// Find token by its prefix.
// Note: Don't forget to validate the token
func (r *authRepositoryImpl) FindAuthToken(ctx context.Context, prefix string) (*entities.AuthToken, error) {
	var token entities.AuthToken
	err := r.DB.WithContext(ctx).Where("prefix = ?", prefix).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, err
}

func (r *authRepositoryImpl) RefreshAuthToken(ctx context.Context, token *entities.AuthToken) error {
	err := r.DB.WithContext(ctx).Where("prefix = ?", token.Prefix).Updates(&entities.AuthToken{
		Secret:        token.Secret,
		ExpiresAt:     token.ExpiresAt,
		RefreshSecret: token.RefreshSecret,
		RefreshedAt:   token.RefreshedAt,
	}).Error
	return err
}

func (r *authRepositoryImpl) RevokeAuthToken(ctx context.Context, prefix string) error {
	now := time.Now()
	err := r.DB.WithContext(ctx).Where("prefix = ?", prefix).Updates(&entities.AuthToken{
		RevokedAt: &now,
		IsRevoked: true,
	}).Error
	return err
}

func (r *authRepositoryImpl) CreateUserToken(ctx context.Context, token *entities.UserToken) (*entities.UserToken, error) {
	err := r.DB.WithContext(ctx).Create(token).Error
	if err != nil {
		return nil, err
	}
	return token, err
}

func (r *authRepositoryImpl) FindUserToken(ctx context.Context, prefix string, tokenType entities.UserTokenType) (*entities.UserToken, error) {
	var token entities.UserToken
	err := r.DB.WithContext(ctx).Where(
		"prefix = ? AND type = ?", prefix, tokenType,
	).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, err
}

func (r *authRepositoryImpl) FindUserTokenByUserId(ctx context.Context, userId uint, tokenType entities.UserTokenType) (*[]entities.UserToken, error) {
	var tokens []entities.UserToken

	err := r.DB.WithContext(ctx).Where(
		"user_id = ? AND type = ? AND is_used = ?", userId, tokenType, false,
	).Find(&tokens).Order("id DESC").Error

	if err != nil {
		return nil, err
	}
	return &tokens, err
}

func (r *authRepositoryImpl) RefreshUserToken(ctx context.Context, token *entities.UserToken) error {
	err := r.DB.WithContext(ctx).Where(
		"prefix = ? AND type = ?", token.Prefix, token.Type,
	).Updates(&entities.UserToken{
		Secret:        token.Secret,
		ExpiresAt:     token.ExpiresAt,
		RefreshSecret: token.RefreshSecret,
		RefreshedAt:   token.RefreshedAt,
	}).Error
	return err
}

func (r *authRepositoryImpl) RevokeUserToken(ctx context.Context, prefix string) error {
	now := time.Now()
	err := r.DB.WithContext(ctx).Where("prefix = ?", prefix).Updates(&entities.UserToken{
		RevokedAt: &now,
		IsRevoked: true,
	}).Error
	return err
}

func (r *authRepositoryImpl) MarkAsUsedUserToken(ctx context.Context, tokenId uint) error {
	now := time.Now()
	err := r.DB.WithContext(ctx).Where("id = ?", tokenId).Updates(&entities.UserToken{
		UsedAt: &now,
		IsUsed: true,
	}).Error
	return err
}
