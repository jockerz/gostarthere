package user

import (
	"context"
	"errors"
	"strings"
	"time"
	"vnti/pkg/entities"

	"gorm.io/gorm"
)

const (
	userTokenPrefixLength      = 30
	userTokenSecretLength      = 40
	userTokenRefreshSecretLen  = 40
	userTokenExpireDays        = 1
	userTokenRefreshExpireDays = 7
)

var (
	ErrNoUpdate error = errors.New("not found")
)

type Repository interface {
	Create(context.Context, *entities.User) (*entities.User, error)
	FindByEmail(ctx context.Context, email string) (*entities.User, error)
	FindByID(ctx context.Context, id uint) (*entities.User, error)
	FindByUsername(ctx context.Context, username string) (*entities.User, error)
	Update(context.Context, *entities.User) (*entities.User, error)
	UpdateByID(context.Context, uint, map[string]any, uint) error
	// GetMany() (*[]entities.User, error)
	Delete(context.Context, *entities.User) error
}

type repository struct {
	DB *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{DB: db}
}

func (r *repository) Create(ctx context.Context, user *entities.User) (*entities.User, error) {
	user.Email = strings.ToLower(user.Email)
	user.Username = strings.ToLower(user.Username)

	err := r.DB.WithContext(ctx).Create(user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *repository) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	var user entities.User

	email = strings.ToLower(email)
	err := r.DB.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, err
}

func (r *repository) FindByID(ctx context.Context, id uint) (*entities.User, error) {
	var u entities.User
	err := r.DB.WithContext(ctx).First(&u, id).Error
	if err != nil {
		return nil, err
	}
	return &u, err
}

func (r *repository) FindByUsername(ctx context.Context, username string) (*entities.User, error) {
	var user entities.User

	u := strings.ToLower(username)
	err := r.DB.WithContext(ctx).Where("username = ?", u).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, err
}

func (r *repository) GetMany(ctx context.Context) (*[]entities.User, error) {
	var users []entities.User
	err := r.DB.WithContext(ctx).Find(&users).Error
	return &users, err
}

func (r *repository) Update(ctx context.Context, user *entities.User) (*entities.User, error) {
	rowsAffected, err := gorm.G[entities.User](r.DB).Updates(ctx, *user)
	if err != nil {
		return nil, err
	} else if rowsAffected == 0 {
		return nil, ErrNoUpdate
	}
	return user, nil
}

func (r *repository) UpdateByID(ctx context.Context, userId uint, data map[string]any, updatedBy uint) error {
	data["updated_by"] = updatedBy
	data["updated_at"] = time.Now()

	rowsAffected, err := gorm.G[map[string]any](r.DB).
		Table(entities.User{}.TableName()).
		Where("id = ?", userId).
		Updates(ctx, data)

	if err != nil {
		return err
	} else if rowsAffected == 0 {
		return ErrNoUpdate
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, u *entities.User) error {
	return r.DB.WithContext(ctx).Delete(u).Error
}

type UserTokenRepository interface {
	CreateUserToken(ctx context.Context, token *entities.UserToken) (*entities.UserToken, error)
	FindUserToken(ctx context.Context, prefix string, tokenType entities.UserTokenType) (*entities.UserToken, error)
	MarkAsUsedUserToken(ctx context.Context, id uint) error
}
