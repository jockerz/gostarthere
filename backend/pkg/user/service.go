package user

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
	"vnti/pkg/tasks"

	"github.com/gofiber/fiber/v3/log"
	"github.com/gofiber/utils/v2"
	"github.com/hibiken/asynq"
	"golang.org/x/crypto/bcrypt"
)

type TokenAction string

const (
	TokenActionBaseURL string = "http://localhost:5173/auth"
	BaseURL            string = "http://localhost:5173"
)

var (
	ErrInvalidID          = errors.New("Invalid user ID")
	ErrNotFound           = errors.New("user not found")
	ErrDuplicateEmail     = errors.New("email already exists")
	ErrEmailUpdateFailed  = errors.New("email update failed")
	ErrDuplicateUsername  = errors.New("username already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUpdateFailed       = errors.New("user update failed")
	ErrDeleteFailed       = errors.New("user delete failed")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrUsedToken          = errors.New("token has been used")
	ErrCreateTokenFailed  = errors.New("create user token failed")
)

type Service interface {
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

type service struct {
	config     *internal.Config
	repository Repository
	tokenRepo  UserTokenRepository

	asynqClient   *asynq.Client
	skipTaskQueue bool
}

func NewService(
	config *internal.Config,
	repository Repository,
	tokenRepo UserTokenRepository,
	asyncClient *asynq.Client,
) Service {
	return &service{
		config:      config,
		repository:  repository,
		tokenRepo:   tokenRepo,
		asynqClient: asyncClient,
	}
}

func (s *service) Create(ctx context.Context, u *entities.User) (*entities.User, error) {
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

func (s *service) FindByID(ctx context.Context, id uint) (*entities.User, error) {
	u, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return nil, ErrNotFound
	}
	return u, nil
}

func (s *service) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	u, err := s.repository.FindByEmail(ctx, email)
	if err != nil {
		return nil, ErrNotFound
	}
	return u, nil
}

func (s *service) FindByUsername(ctx context.Context, username string) (*entities.User, error) {
	u, err := s.repository.FindByUsername(ctx, username)
	if err != nil {
		return nil, ErrNotFound
	}
	return u, nil
}

// func (s *service) GetMany(ctx context.Context) (*[]entities.User, error) {
// 	return s.repository.GetMany(ctx)
// }

func (s *service) Update(ctx context.Context, u *entities.User) (*entities.User, error) {
	if u.ID == 0 {
		return nil, ErrInvalidID
	} else if _, err := s.repository.FindByID(ctx, u.ID); err != nil {
		return nil, ErrNotFound
	}

	user, err := s.repository.Update(ctx, u)
	if err != nil {
		return nil, ErrUpdateFailed
	}
	return user, nil
}

func (s *service) Delete(ctx context.Context, user *entities.User) error {
	_, err := s.repository.FindByID(ctx, user.ID)
	if err != nil {
		return ErrNotFound
	}
	if err := s.repository.Delete(ctx, user); err != nil {
		return ErrDeleteFailed
	}
	return nil
}

func (s *service) UpdateProfile(ctx context.Context, userID uint, name, username, avatar string) (*entities.User, error) {
	// user, err := s.repository.FindByID(ctx, userID)
	// if err != nil {
	// 	return nil, ErrNotFound
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
		return nil, ErrUpdateFailed
	}
	user, _ := s.repository.FindByID(ctx, userID)
	return user, nil
}

func (s *service) UploadAvatar(ctx context.Context, userID uint, fileBytes []byte, filename string) (*entities.User, error) {
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
		return nil, ErrUpdateFailed
	}

	return s.repository.FindByID(ctx, userID)
}

func (s *service) SetPassword(ctx context.Context, userID uint, password string) error {
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

func (s *service) ChangePassword(ctx context.Context, userID uint, currentPassword, newPassword string) error {
	user, err := s.repository.FindByID(ctx, userID)
	if err != nil {
		return ErrInvalidCredentials
	}

	// Third party registered user might not have password yet
	err = bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(currentPassword))
	if user.HasPassword() && err != nil {
		return ErrInvalidCredentials
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

func (s *service) generateUserToken(ctx context.Context, userID uint, tokenType entities.UserTokenType, newEmail string) (*entities.UserToken, error) {
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
		return nil, ErrCreateTokenFailed
	}
	toSave.Secret = string(hashedSecret)
	hashedRefreshSecret, err := bcrypt.GenerateFromPassword([]byte(toSave.RefreshSecret), bcrypt.DefaultCost)
	if err != nil {
		return nil, ErrCreateTokenFailed
	}
	toSave.RefreshSecret = string(hashedRefreshSecret)

	_, err = s.tokenRepo.CreateUserToken(ctx, &toSave)
	if err != nil {
		return nil, ErrCreateTokenFailed
	}
	return &token, nil
}

func (s *service) RequestEmailUpdate(ctx context.Context, userID uint, newEmail string, password string) error {
	user, err := s.repository.FindByID(ctx, userID)
	if err != nil {
		return ErrInvalidCredentials
	}

	if user.Password == nil || bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(password)) != nil {
		return ErrInvalidCredentials
	}

	// Email already exists
	if existing, _ := s.repository.FindByEmail(ctx, newEmail); existing != nil {
		return ErrDuplicateEmail
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
		log.Error(err)
	}
	return nil
}

func (s *service) ConfirmEmailUpdate(ctx context.Context, tokenStr string) error {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 2 {
		return ErrInvalidToken
	}

	token, err := s.tokenRepo.FindUserToken(ctx, parts[0], entities.TokenEmailUpdate)
	if err != nil {
		return ErrInvalidToken
	} else if token.ExpiresAt.Before(time.Now()) {
		return ErrInvalidToken
	} else if token.IsUsed {
		return ErrUsedToken
	} else if !s.checkTokenSecret(token.Secret, parts[1]) {
		return ErrInvalidToken
	}

	user, err := s.repository.FindByID(ctx, token.UserID)
	if err != nil {
		return ErrInvalidToken
	}

	if err = s.repository.UpdateByID(ctx, user.ID, map[string]any{
		"email": token.Data,
	}, user.ID); err != nil {
		return ErrEmailUpdateFailed
	}

	_ = s.tokenRepo.MarkAsUsedUserToken(ctx, token.ID)
	return nil
}

func (s *service) checkTokenSecret(hashed, plain string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain)); err != nil {
		return false
	}
	return true
}

func (s *service) sendTokenEmail(ctx context.Context, actionType TokenAction, tokenWithSecret, email string) error {
	var url, message, subject string

	switch actionType {
	case tasks.TypeUserEmailUpdate:
		subject = "Email Update"
		url = fmt.Sprintf(
			"%s/dashboard/update-email-confirm?token=%s", BaseURL, tokenWithSecret,
		)
		message = fmt.Sprintf("The confirm email update link: %s", url)
	default:
		return errors.New("Invalid token action")
	}

	task, err := tasks.SendEmailTasks(ctx, s.config.SMTP_USERNAME, email, subject, message)
	if err != nil {
		log.Errorf("Fail to create task %s: %s", actionType, err.Error())
		return tasks.ErrTokenActionCreateTask
	} else if !s.skipTaskQueue {
		_ = s.EnqueueTask(task, actionType)
	}
	return nil
}

func (s *service) EnqueueTask(task *asynq.Task, taskType TokenAction) error {
	taskInfo, err := s.asynqClient.Enqueue(task)
	fmt.Printf("task queue info: %+v, err: %v\n", taskInfo, err)
	if err != nil {
		log.Errorf("Fail to enqueue task %s: %s", taskType, err.Error())
		return tasks.ErrTokenActionEnqueueTask
	}
	return nil
}

// For testing purposes only
func (s *service) SetSkipTaskQueue(v bool) {
	s.skipTaskQueue = v
}
