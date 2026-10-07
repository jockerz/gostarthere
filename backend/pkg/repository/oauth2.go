package repository

import (
	"context"

	"vnti/pkg/entities"

	"gorm.io/gorm"
)

type OAuth2Repository interface {
	Create(ctx context.Context, entry *entities.UserAuthProvider) error
	FindByProvider(ctx context.Context, provider entities.AuthProvider, providerUserID string) (*entities.UserAuthProvider, error)
	FindByUserID(ctx context.Context, userID uint) ([]*entities.UserAuthProvider, error)
}

type oauth2RepositoryImpl struct {
	db *gorm.DB
}

func NewOAuth2Repository(db *gorm.DB) OAuth2Repository {
	return &oauth2RepositoryImpl{db: db}
}

func (r *oauth2RepositoryImpl) Create(ctx context.Context, entry *entities.UserAuthProvider) error {
	return r.db.WithContext(ctx).Create(entry).Error
}

func (r *oauth2RepositoryImpl) FindByProvider(ctx context.Context, provider entities.AuthProvider, providerUserID string) (*entities.UserAuthProvider, error) {
	var entry entities.UserAuthProvider

	err := r.db.WithContext(ctx).Where(
		"provider = ? AND provider_user_id = ?", provider, providerUserID,
	).First(&entry).Error
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func (r *oauth2RepositoryImpl) FindByUserID(ctx context.Context, userID uint) ([]*entities.UserAuthProvider, error) {
	var entries []*entities.UserAuthProvider
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&entries).Error
	return entries, err
}
