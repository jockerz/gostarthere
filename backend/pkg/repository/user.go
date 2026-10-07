package repository

import (
	"context"
	"errors"
	"strings"
	"time"
	"vnti/pkg/entities"

	"gorm.io/gorm"
)

var (
	ErrUserNoUpdate error = errors.New("not found")
)

type UserRepository interface {
	Create(context.Context, *entities.User) (*entities.User, error)
	FindByEmail(ctx context.Context, email string) (*entities.User, error)
	FindByID(ctx context.Context, id uint) (*entities.User, error)
	FindByUsername(ctx context.Context, username string) (*entities.User, error)
	Update(context.Context, *entities.User) (*entities.User, error)
	UpdateByID(context.Context, uint, map[string]any, uint) error
	// GetMany() (*[]entities.User, error)
	Delete(context.Context, *entities.User) error
}

type userRepositoryImpl struct {
	DB *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepositoryImpl{DB: db}
}

func (r *userRepositoryImpl) Create(ctx context.Context, user *entities.User) (*entities.User, error) {
	user.Email = strings.ToLower(user.Email)
	user.Username = strings.ToLower(user.Username)

	err := r.DB.WithContext(ctx).Create(user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *userRepositoryImpl) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	var user entities.User

	email = strings.ToLower(email)
	err := r.DB.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, err
}

func (r *userRepositoryImpl) FindByID(ctx context.Context, id uint) (*entities.User, error) {
	var u entities.User
	err := r.DB.WithContext(ctx).First(&u, id).Error
	if err != nil {
		return nil, err
	}
	return &u, err
}

func (r *userRepositoryImpl) FindByUsername(ctx context.Context, username string) (*entities.User, error) {
	var user entities.User

	u := strings.ToLower(username)
	err := r.DB.WithContext(ctx).Where("username = ?", u).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, err
}

func (r *userRepositoryImpl) GetMany(ctx context.Context) (*[]entities.User, error) {
	var users []entities.User
	err := r.DB.WithContext(ctx).Find(&users).Error
	return &users, err
}

func (r *userRepositoryImpl) Update(ctx context.Context, user *entities.User) (*entities.User, error) {
	rowsAffected, err := gorm.G[entities.User](r.DB).Updates(ctx, *user)
	if err != nil {
		return nil, err
	} else if rowsAffected == 0 {
		return nil, ErrUserNoUpdate
	}
	return user, nil
}

func (r *userRepositoryImpl) UpdateByID(ctx context.Context, userId uint, data map[string]any, updatedBy uint) error {
	data["updated_by"] = updatedBy
	data["updated_at"] = time.Now()

	rowsAffected, err := gorm.G[map[string]any](r.DB).
		Table(entities.User{}.TableName()).
		Where("id = ?", userId).
		Updates(ctx, data)

	if err != nil {
		return err
	} else if rowsAffected == 0 {
		return ErrUserNoUpdate
	}
	return nil
}

func (r *userRepositoryImpl) Delete(ctx context.Context, u *entities.User) error {
	return r.DB.WithContext(ctx).Delete(u).Error
}
