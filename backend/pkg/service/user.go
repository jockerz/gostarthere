package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"vnti/internal"
	"vnti/pkg/entities"
	"vnti/pkg/repository"
	"vnti/pkg/tasks"

	"github.com/gofiber/utils/v2"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

const (
	userTokenPrefixLength      = 30
	userTokenSecretLength      = 40
	userTokenRefreshSecretLen  = 40
	userTokenExpireDays        = 1
	userTokenRefreshExpireDays = 7
)

type UserTokenAction string

var (
	ErrUserInvalidID          = errors.New("Invalid user ID")
	ErrUserNotFound           = errors.New("user not found")
	ErrUserDuplicateEmail     = errors.New("email already exists")
	ErrUserEmailUpdateFailed  = errors.New("email update failed")
	ErrUserDuplicateUsername  = errors.New("username already exists")
	ErrUserInvalidCredentials = errors.New("invalid credentials")
	ErrUserUpdateFailed       = errors.New("user update failed")
	ErrUserDeleteFailed       = errors.New("user delete failed")
	ErrUserInvalidToken       = errors.New("invalid or expired token")
	ErrUserUsedToken          = errors.New("token has been used")
	ErrUserCreateTokenFailed  = errors.New("create user token failed")
)

type UserService interface {
	Create(context.Context, *entities.User) (*entities.User, error)
	FindByID(context.Context, uint) (*entities.User, error)
	FindByEmail(context.Context, string) (*entities.User, error)
	FindByUsername(context.Context, string) (*entities.User, error)
	// GetMany(context.Context) (*[]entities.User, error)
	Update(context.Context, *entities.User) (*entities.User, error)
	// UpdateByMap(context.Context, string, map[string]any) (*entities.User, error)
	Delete(context.Context, *entities.User) error
	UpdateProfile(ctx context.Context, userID uint, name, username, avatar string) (*entities.User, error)
	ChangePassword(ctx context.Context, userID uint, currentPassword, newPassword string) error
	SetPassword(ctx context.Context, userID uint, password string) error
	UploadAvatar(ctx context.Context, userID uint, fileBytes []byte, filename string) (*entities.User, error)
	RequestEmailUpdate(ctx context.Context, userID uint, newEmail string, password string) error
	ConfirmEmailUpdate(ctx context.Context, tokenStr string) error

	// For testing only
	SetSkipTaskQueue(v bool)
}

type UserTokenRepository interface {
	CreateUserToken(ctx context.Context, token *entities.UserToken) (*entities.UserToken, error)
	FindUserToken(ctx context.Context, prefix string, tokenType entities.UserTokenType) (*entities.UserToken, error)
	MarkAsUsedUserToken(ctx context.Context, id uint) error
}

type userServiceImpl struct {
	config     *internal.Config
	repository repository.UserRepository
	tokenRepo  UserTokenRepository

	asynqClient   *asynq.Client
	skipTaskQueue bool
}

func NewUserService(
	config *internal.Config,
	repository repository.UserRepository,
	tokenRepo UserTokenRepository,
	asyncClient *asynq.Client,
) UserService {
	return &userServiceImpl{
		config:      config,
		repository:  repository,
		tokenRepo:   tokenRepo,
		asynqClient: asyncClient,
	}
}

func (s *userServiceImpl) Create(ctx context.Context, u *entities.User) (*entities.User, error) {
	if strings.TrimSpace(u.Email) == "" {
		return nil, errors.New("email is required")
	}
	if strings.TrimSpace(u.Username) == "" {
		return nil, errors.New("username is required")
	}
	if u.Password == nil || strings.TrimSpace(*u.Password) == "" {
		return nil, errors.New("password is required")
	}
	u.Email = strings.ToLower(u.Email)
	u.Username = strings.ToLower(u.Username)

	user, err := s.repository.Create(ctx, u)
	if err != nil {
		// TODO: log error
		return nil, errors.New("create user failed")
	}
	return user, nil
}

func (s *userServiceImpl) FindByID(ctx context.Context, id uint) (*entities.User, error) {
	u, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (s *userServiceImpl) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	u, err := s.repository.FindByEmail(ctx, email)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (s *userServiceImpl) FindByUsername(ctx context.Context, username string) (*entities.User, error) {
	u, err := s.repository.FindByUsername(ctx, username)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return u, nil
}

// func (s *userServiceImpl) GetMany(ctx context.Context) (*[]entities.User, error) {
// 	return s.repository.GetMany(ctx)
// }

func (s *userServiceImpl) Update(ctx context.Context, u *entities.User) (*entities.User, error) {
	if u.ID == 0 {
		return nil, ErrUserInvalidID
	} else if _, err := s.repository.FindByID(ctx, u.ID); err != nil {
		return nil, ErrUserNotFound
	}

	user, err := s.repository.Update(ctx, u)
	if err != nil {
		return nil, ErrUserUpdateFailed
	}
	return user, nil
}

func (s *userServiceImpl) Delete(ctx context.Context, user *entities.User) error {
	_, err := s.repository.FindByID(ctx, user.ID)
	if err != nil {
		return ErrUserNotFound
	}
	if err := s.repository.Delete(ctx, user); err != nil {
		return ErrUserDeleteFailed
	}
	return nil
}

func (s *userServiceImpl) UpdateProfile(ctx context.Context, userID uint, name, username, avatar string) (*entities.User, error) {
	// user, err := s.repository.FindByID(ctx, userID)
	// if err != nil {
	// 	return nil, ErrUserNotFound
	// }

	updateData := map[string]any{}
	if name != "" {
		updateData["name"] = name
	}
	if username != "" {
		updateData["username"] = username
	}
	if avatar != "" {
		updateData["avatar"] = avatar
	}

	err := s.repository.UpdateByID(ctx, userID, updateData, userID)
	if err != nil {
		// TODO: log error
		return nil, ErrUserUpdateFailed
	}
	user, _ := s.repository.FindByID(ctx, userID)
	return user, nil
}

func (s *userServiceImpl) UploadAvatar(ctx context.Context, userID uint, fileBytes []byte, filename string) (*entities.User, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	validExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	if !validExts[ext] {
		return nil, errors.New("invalid file extension")
	}

	newFilename := fmt.Sprintf("%d-%d%s", userID, time.Now().UnixNano(), ext)
	savePath := filepath.Join(s.config.MEDIA_PATH, "avatar", newFilename)

	if err := os.WriteFile(savePath, fileBytes, 0644); err != nil {
		return nil, errors.New("failed to save file")
	}

	avatarPath := filepath.Join("avatar", newFilename)
	if err := s.repository.UpdateByID(ctx, userID, map[string]any{
		"avatar": avatarPath,
	}, userID); err != nil {
		return nil, ErrUserUpdateFailed
	}

	return s.repository.FindByID(ctx, userID)
}

func (s *userServiceImpl) SetPassword(ctx context.Context, userID uint, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	err = s.repository.UpdateByID(ctx, userID, map[string]any{
		"password":     string(hashedPassword),
		"has_password": true,
	}, userID)
	return err
}

func (s *userServiceImpl) ChangePassword(ctx context.Context, userID uint, currentPassword, newPassword string) error {
	user, err := s.repository.FindByID(ctx, userID)
	if err != nil {
		return ErrUserInvalidCredentials
	}

	// Third party registered user might not have password yet
	err = bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(currentPassword))
	if user.HasPassword() && err != nil {
		return ErrUserInvalidCredentials
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	err = s.repository.UpdateByID(ctx, user.ID, map[string]any{
		"password": string(hashedPassword),
	}, user.ID)
	return err
}

func (s *userServiceImpl) generateUserToken(ctx context.Context, userID uint, tokenType entities.UserTokenType, newEmail string) (*entities.UserToken, error) {
	prefix := utils.GenerateSecureToken(userTokenPrefixLength)
	secret := utils.GenerateSecureToken(userTokenSecretLength)
	refreshSecret := utils.GenerateSecureToken(userTokenRefreshSecretLen)

	token := entities.UserToken{
		UserID:           userID,
		Prefix:           prefix,
		Secret:           secret,
		RefreshSecret:    refreshSecret,
		Type:             tokenType,
		Data:             newEmail,
		ExpiresAt:        time.Now().AddDate(0, 0, userTokenExpireDays),
		RefreshExpiresAt: time.Now().AddDate(0, 0, userTokenRefreshExpireDays),
		CreatedBy:        userID,
	}

	toSave := token
	hashedSecret, err := bcrypt.GenerateFromPassword([]byte(toSave.Secret), bcrypt.DefaultCost)
	if err != nil {
		return nil, ErrUserCreateTokenFailed
	}
	toSave.Secret = string(hashedSecret)
	hashedRefreshSecret, err := bcrypt.GenerateFromPassword([]byte(toSave.RefreshSecret), bcrypt.DefaultCost)
	if err != nil {
		return nil, ErrUserCreateTokenFailed
	}
	toSave.RefreshSecret = string(hashedRefreshSecret)

	_, err = s.tokenRepo.CreateUserToken(ctx, &toSave)
	if err != nil {
		return nil, ErrUserCreateTokenFailed
	}
	return &token, nil
}

func (s *userServiceImpl) RequestEmailUpdate(ctx context.Context, userID uint, newEmail string, password string) error {
	user, err := s.repository.FindByID(ctx, userID)
	if err != nil {
		return ErrUserInvalidCredentials
	}

	if user.Password == nil || bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(password)) != nil {
		return ErrUserInvalidCredentials
	}

	// Email already exists
	if existing, _ := s.repository.FindByEmail(ctx, newEmail); existing != nil {
		return ErrUserDuplicateEmail
	}

	token, err := s.generateUserToken(ctx, user.ID, entities.TokenEmailUpdate, newEmail)
	if err != nil {
		return err
	}

	// Send email
	if err = s.sendTokenEmail(
		ctx,
		tasks.TypeUserEmailUpdate,
		token.AsTokenWithSecret(token.Secret),
		newEmail,
	); err != nil {
		log.Err(err)
	}
	return nil
}

func (s *userServiceImpl) ConfirmEmailUpdate(ctx context.Context, tokenStr string) error {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 2 {
		return ErrUserInvalidToken
	}

	token, err := s.tokenRepo.FindUserToken(ctx, parts[0], entities.TokenEmailUpdate)
	if err != nil {
		return ErrUserInvalidToken
	} else if token.ExpiresAt.Before(time.Now()) {
		return ErrUserInvalidToken
	} else if token.IsUsed {
		return ErrUserUsedToken
	} else if !s.checkTokenSecret(token.Secret, parts[1]) {
		return ErrUserInvalidToken
	}

	user, err := s.repository.FindByID(ctx, token.UserID)
	if err != nil {
		return ErrUserInvalidToken
	}

	if err = s.repository.UpdateByID(ctx, user.ID, map[string]any{
		"email": token.Data,
	}, user.ID); err != nil {
		return ErrUserEmailUpdateFailed
	}

	_ = s.tokenRepo.MarkAsUsedUserToken(ctx, token.ID)
	return nil
}

func (s *userServiceImpl) checkTokenSecret(hashed, plain string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain)); err != nil {
		return false
	}
	return true
}

func (s *userServiceImpl) sendTokenEmail(ctx context.Context, actionType UserTokenAction, tokenWithSecret, email string) error {
	var url, message, subject string

	switch actionType {
	case tasks.TypeUserEmailUpdate:
		subject = "Email Update"
		url = fmt.Sprintf(
			"%s/%s/update-email-confirm?token=%s", s.config.BASE_URL, s.config.FE_DASHBOARD_PATH, tokenWithSecret,
		)
		message = fmt.Sprintf("The confirm email update link: %s", url)
	default:
		return errors.New("Invalid token action")
	}

	task, err := tasks.SendEmailTasks(ctx, s.config.SMTP_USERNAME, email, subject, message)
	if err != nil {
		log.Error().
			Str("action_type", string(actionType)).
			Errs("err", []error{err}).
			Msg("Fail to create task")
		return tasks.ErrTokenActionCreateTask
	} else if !s.skipTaskQueue {
		_ = s.EnqueueTask(task, actionType)
	}
	return nil
}

func (s *userServiceImpl) EnqueueTask(task *asynq.Task, taskType UserTokenAction) error {
	taskInfo, err := s.asynqClient.Enqueue(task)
	log.Info().Any("task_info", taskInfo).Errs("error", []error{err})

	if err != nil {
		log.Error().
			Str("action_type", "enqueue_task").
			Errs("err", []error{err}).
			Msg("Fail to enqueue task")
		return tasks.ErrTokenActionEnqueueTask
	}
	return nil
}

// For testing purposes only
func (s *userServiceImpl) SetSkipTaskQueue(v bool) {
	s.skipTaskQueue = v
}
