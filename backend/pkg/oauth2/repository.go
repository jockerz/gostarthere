package oauth2

import (
	"context"

	"vnti/pkg/entities"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, entry *entities.UserAuthProvider) error
	FindByProvider(ctx context.Context, provider entities.AuthProvider, providerUserID string) (*entities.UserAuthProvider, error)
	FindByUserID(ctx context.Context, userID uint) ([]*entities.UserAuthProvider, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, entry *entities.UserAuthProvider) error {
	return r.db.WithContext(ctx).Create(entry).Error
}

func (r *repository) FindByProvider(ctx context.Context, provider entities.AuthProvider, providerUserID string) (*entities.UserAuthProvider, error) {
	var entry entities.UserAuthProvider

	err := r.db.WithContext(ctx).Where(
		"provider = ? AND provider_user_id = ?", provider, providerUserID,
	).First(&entry).Error
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func (r *repository) FindByUserID(ctx context.Context, userID uint) ([]*entities.UserAuthProvider, error) {
	var entries []*entities.UserAuthProvider
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&entries).Error
	return entries, err
}
