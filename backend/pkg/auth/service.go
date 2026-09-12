package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"vnti/internal"
	"vnti/internal/logger"
	"vnti/pkg/entities"
	"vnti/pkg/tasks"
	"vnti/pkg/user"

	"github.com/gofiber/utils/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials     = errors.New("invalid credentials")
	ErrInvalidToken           = errors.New("invalid or expired token")
	ErrUsedToken              = errors.New("token has been used")
	ErrEmailAlreadyExists     = errors.New("email already exists")
	ErrUsernameAlreadyExists  = errors.New("username already exists")
	ErrCreateTokenFailed      = errors.New("create auth token failed")
	ErrUserActivationFailed   = errors.New("user activation failed")
	ErrResendActivationFailed = errors.New("user activation reset failed")
	ErrInvalidRefreshToken    = errors.New("invalid or expired token")
	ErrRefreshTokenExpired    = errors.New("reset token expired")
	ErrUserNotFound           = errors.New("user not found")
)

type TokenAction string

const (
	TokenActionActivation      TokenAction = "activation"
	TokenActionResetActivation TokenAction = "reset_activation"
	TokenActionResetPassword   TokenAction = "reset_password"
)

type Service interface {
	Login(ctx context.Context, email, password string) (*entities.User, *entities.BearerToken, error)
	Register(context.Context, *entities.Register) (*entities.User, error)

	ActivateAccount(ctx context.Context, token string) error
	ResendActivation(ctx context.Context, email string) error

	ForgotPassword(ctx context.Context, email_or_username string) error
	ResetPassword(ctx context.Context, token, newPassword string) error

	Logout(ctx context.Context, token string) error
	RefreshToken(ctx context.Context, refreshToken string) (*entities.BearerToken, error)

	// Generate user token.
	// Return with un-hashed secret and refresh_secret.
	// But have to save the hashed one on DB.
	GenerateUserToken(context.Context, uint, entities.UserTokenType) (*entities.UserToken, error)

	// Generate auth token.
	// Return with un-hashed secret and refresh_secret.
	// But have to save the hashed one on DB.
	GenerateAuthToken(context.Context, *entities.User) (*entities.AuthToken, error)

	GetAuthTokenByJWTToken(context.Context, string) (*entities.AuthToken, error)

	AuthByToken(ctx context.Context, token string) (*entities.User, error)
	AuthByRefreshToken(ctx context.Context, token string) (*entities.User, error)

	// For testing only
	SetSkipTaskQueue(v bool)
}

/*
	{
		"access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
		"refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
		"token_type": "Bearer",
		"expires_in": 3600,
		"expires_at": "2026-05-26T15:30:45Z",
	 	"scope": "read write"
	  }
*/

type Claims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

type service struct {
	config    *internal.Config
	userRepo  user.Repository
	repo      Repository
	jwtSecret string

	asynqClient   *asynq.Client
	skipTaskQueue bool
}

func NewService(
	config *internal.Config,
	userRepo user.Repository,
	repo Repository,
	jwtSecret string,
	asynqClient *asynq.Client,
) Service {
	return &service{
		config:      config,
		userRepo:    userRepo,
		repo:        repo,
		jwtSecret:   jwtSecret,
		asynqClient: asynqClient,
	}
}

func (s *service) GenerateUserToken(ctx context.Context, user_id uint, token_type entities.UserTokenType) (*entities.UserToken, error) {
	prefix := utils.GenerateSecureToken(30)
	// Generate random secret
	secret := utils.GenerateSecureToken(40)
	refresh_secret := utils.GenerateSecureToken(40)

	token := entities.UserToken{
		UserID:           user_id,
		Prefix:           prefix,
		Secret:           secret,
		RefreshSecret:    refresh_secret,
		ExpiresAt:        time.Now().AddDate(0, 0, 1),
		RefreshExpiresAt: time.Now().AddDate(0, 0, 7),
		CreatedBy:        user_id,
		Type:             token_type,
	}

	to_save_token := token
	hashed_secret, err := bcrypt.GenerateFromPassword([]byte(to_save_token.Secret), bcrypt.DefaultCost)
	if err != nil {
		return nil, ErrCreateTokenFailed
	}
	to_save_token.Secret = string(hashed_secret)
	hashed_refresh_secret, err := bcrypt.GenerateFromPassword([]byte(to_save_token.RefreshSecret), bcrypt.DefaultCost)
	if err != nil {
		return nil, ErrCreateTokenFailed
	}
	to_save_token.RefreshSecret = string(hashed_refresh_secret)

	_, err = s.repo.CreateUserToken(ctx, &to_save_token)
	if err != nil {
		return nil, ErrCreateTokenFailed
	}
	return &token, nil
}

func (s *service) GenerateAuthToken(ctx context.Context, user *entities.User) (*entities.AuthToken, error) {
	prefix := utils.GenerateSecureToken(TOKEN_PREFIX_LENGTH)
	secret := utils.GenerateSecureToken(ACCESS_SECRET_LENGTH)
	refresh_secret := utils.GenerateSecureToken(REFRESH_SECRET_LENGTH)

	token := entities.AuthToken{
		UserID:           user.ID,
		Prefix:           prefix,
		Secret:           secret,
		RefreshSecret:    refresh_secret,
		ExpiresAt:        time.Now().AddDate(0, 0, TOKEN_EXPIRE_DAYS),
		RefreshExpiresAt: time.Now().AddDate(0, 0, TOKEN_REFRESH_EXPIRE_DAYS),
		CreatedBy:        user.ID,
	}
	to_save_token := token
	hashed_secret, err := bcrypt.GenerateFromPassword([]byte(to_save_token.Secret), bcrypt.DefaultCost)
	if err != nil {
		return nil, ErrCreateTokenFailed
	}
	to_save_token.Secret = string(hashed_secret)
	hashed_refresh_secret, err := bcrypt.GenerateFromPassword([]byte(to_save_token.RefreshSecret), bcrypt.DefaultCost)
	if err != nil {
		return nil, ErrCreateTokenFailed
	}
	to_save_token.RefreshSecret = string(hashed_refresh_secret)

	_, err = s.repo.CreateAuthToken(ctx, &to_save_token)
	if err != nil {
		return nil, ErrCreateTokenFailed
	}

	l := logger.LogContext(ctx, "auth:service:GenAuthToken", "AuthToken created")
	l.Debug().Str("tokenPrefix", token.Prefix).Msg("User login")

	return &token, nil
}

func (s *service) Login(ctx context.Context, email_or_username, password string) (*entities.User, *entities.BearerToken, error) {
	l := logger.LogContext(ctx, "auth:service:Login", "AuthToken created")
	l.Info().Str("config", string(s.config.ToJSON()))
	var user *entities.User
	var err error

	if strings.Contains(email_or_username, "@") {
		user, err = s.userRepo.FindByEmail(ctx, email_or_username)
	} else {
		user, err = s.userRepo.FindByUsername(ctx, email_or_username)
	}
	log := logger.LogContext(ctx, "auth:service:Login", "login")
	if err != nil {
		log.Debug().Str("username/email", email_or_username).Msg("Invalid email or username")
		return nil, nil, ErrInvalidCredentials
	}

	if user.Password == nil || !s.checkTokenSecret(*user.Password, password) {
		log.Debug().Str("username/email", email_or_username).Msg("Invalid password")
		return nil, nil, ErrInvalidCredentials
	}

	auth_token, err := s.GenerateAuthToken(ctx, user)
	if err != nil {
		return nil, nil, err
	}

	bearer_token := s.toBearerToken(auth_token)

	return user, &bearer_token, nil
}

func (s *service) Register(ctx context.Context, input *entities.Register) (*entities.User, error) {
	if input.Password != input.PasswordValidate {
		return nil, errors.New("passwords do not match")
	}

	existing, _ := s.userRepo.FindByEmail(ctx, input.Email)
	if existing != nil {
		return nil, ErrEmailAlreadyExists
	}

	existing, _ = s.userRepo.FindByUsername(ctx, input.Username)
	if existing != nil {
		return nil, ErrUsernameAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	hashedStr := string(hashedPassword)
	user := &entities.User{
		Email:     input.Email,
		Username:  input.Username,
		Name:      input.Name,
		Password:  &hashedStr,
		Active:    false,
		CreatedAt: time.Now(),
	}

	user, err = s.userRepo.Create(ctx, user)
	if err != nil {
		// TODO: error
		return nil, err
	}

	token, err := s.GenerateUserToken(ctx, user.ID, entities.TokenActivation)
	if err != nil {
		return nil, err
	}

	// TODO: sent email
	if err = s.sendTokenEmail(
		ctx,
		TokenActionActivation,
		token.AsTokenWithSecret(token.Secret),
		user.Email,
	); err != nil {
		log.Error().Err(err)
	}
	return user, nil
}

func (s *service) ActivateAccount(ctx context.Context, tokenStr string) error {
	tokenParts, err := s.splitToken(tokenStr)
	if err != nil {
		return ErrInvalidToken
	}

	token, err := s.repo.FindUserToken(ctx, tokenParts[0], entities.TokenActivation)
	if err != nil {
		return ErrInvalidToken
	} else if token.IsUsed {
		return ErrUsedToken
	} else if token.ExpiresAt.Before(time.Now()) {
		return ErrInvalidToken
	} else if !s.checkTokenSecret(token.Secret, tokenParts[1]) {
		return ErrInvalidToken
	}

	user, err := s.userRepo.FindByID(ctx, token.UserID)
	if err != nil {
		// TODO: log, because token is found but the user is not
		return ErrInvalidToken
	}

	user.Active = true
	// _, err = s.userRepo.Update(ctx, user)
	err = s.userRepo.UpdateByID(ctx, user.ID, map[string]any{"active": true}, user.ID)
	if err != nil {
		// TODO: log error
		return ErrUserActivationFailed
	}

	_ = s.repo.MarkAsUsedUserToken(ctx, token.ID)
	return nil
}

func (s *service) ResendActivation(ctx context.Context, email_or_username string) error {
	var err error
	var user *entities.User

	if strings.Contains(email_or_username, "@") {
		user, err = s.userRepo.FindByEmail(ctx, email_or_username)
	} else {
		user, err = s.userRepo.FindByUsername(ctx, email_or_username)
	}
	if err != nil {
		// TODO: log because, token found with invalid user
		return ErrUserNotFound
	}

	token, err := s.GenerateUserToken(ctx, user.ID, entities.TokenActivation)
	if err != nil {
		// TODO: error
		return ErrResendActivationFailed
	}

	// sent email task
	if err = s.sendTokenEmail(
		ctx,
		TokenActionResetActivation,
		token.AsTokenWithSecret(token.Secret),
		user.Email,
	); err != nil {
		log.Error().Err(err)
	}

	return nil
}

func (s *service) ForgotPassword(ctx context.Context, email_or_username string) error {
	var user *entities.User
	var err error

	if strings.Contains(email_or_username, "@") {
		user, err = s.userRepo.FindByEmail(ctx, email_or_username)
	} else {
		user, err = s.userRepo.FindByUsername(ctx, email_or_username)
	}
	if err != nil {
		return nil
	}

	token, err := s.GenerateUserToken(ctx, user.ID, entities.TokenReset)
	if err != nil {
		// TODO: error
		return err
	}

	if err = s.sendTokenEmail(
		ctx,
		TokenActionResetPassword,
		token.AsTokenWithSecret(token.Secret),
		user.Email,
	); err != nil {
		log.Error().Err(err)
	}

	return nil
}

func (s *service) ResetPassword(ctx context.Context, tokenStr, newPassword string) error {
	tokenParts, err := s.splitToken(tokenStr)
	if err != nil {
		return ErrInvalidToken
	}

	token, err := s.repo.FindUserToken(ctx, tokenParts[0], entities.TokenReset)
	if err != nil {
		return ErrInvalidToken
	} else if token.IsUsed {
		return ErrUsedToken
	} else if token.ExpiresAt.Before(time.Now()) {
		return ErrInvalidToken
	} else if !s.checkTokenSecret(token.Secret, tokenParts[1]) {
		return ErrInvalidToken
	}

	user, err := s.userRepo.FindByID(ctx, token.UserID)
	if err != nil {
		// TODO: log, because token is found but the user is not
		return ErrInvalidToken
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	hashedStr := string(hashedPassword)
	user.Password = &hashedStr
	err = s.userRepo.UpdateByID(ctx, user.ID, map[string]any{
		"password":     hashedStr,
		"has_password": true,
	}, user.ID)
	if err != nil {
		return err
	}

	// TODO: log error
	// if err = s.repo.MarkuAsUsedUserToken(ctx, token.Prefix, entities.TokenReset); err != nil {
	// 	return err
	// }
	_ = s.repo.MarkAsUsedUserToken(ctx, token.ID)

	return nil
}

func (s *service) Logout(ctx context.Context, tokenStr string) error {
	tokenParts, err := s.splitToken(tokenStr)
	if err != nil {
		return ErrInvalidToken
	}

	token, err := s.repo.FindAuthToken(ctx, tokenParts[0])
	if err != nil {
		return ErrInvalidToken
	}

	if err := bcrypt.CompareHashAndPassword([]byte(token.Secret), []byte(tokenParts[1])); err != nil {
		return ErrInvalidToken
	}

	l := logger.LogContext(ctx, "auth:service:Logout", "logout")
	l.Debug().Str("tokenPrefix", tokenParts[0]).Msg("User logout")

	return s.repo.RevokeAuthToken(ctx, token.Prefix)
}

func (s *service) RefreshToken(ctx context.Context, refreshTokenStr string) (*entities.BearerToken, error) {
	tokenParts, err := s.splitToken(refreshTokenStr)
	if err != nil {
		return nil, ErrInvalidToken
	}

	token, err := s.repo.FindAuthToken(ctx, tokenParts[0])
	if err != nil {
		return nil, ErrInvalidToken
	}

	if token.IsRevoked {
		return nil, ErrInvalidToken
	}

	if token.RefreshExpiresAt.Before(time.Now()) {
		return nil, ErrRefreshTokenExpired
	}

	if err := bcrypt.CompareHashAndPassword([]byte(token.RefreshSecret), []byte(tokenParts[1])); err != nil {
		return nil, ErrInvalidToken
	}

	newSecret := utils.GenerateSecureToken(ACCESS_SECRET_LENGTH)
	newRefreshSecret := utils.GenerateSecureToken(REFRESH_SECRET_LENGTH)

	newToken := *token
	newToken.Secret = newSecret
	newToken.RefreshSecret = newRefreshSecret
	newToken.ExpiresAt = time.Now().AddDate(0, 0, 1)
	newToken.RefreshExpiresAt = time.Now().AddDate(0, 0, 7)
	newToken.RefreshedAt = time.Now()

	toSaveToken := newToken
	hashedSecret, err := bcrypt.GenerateFromPassword([]byte(toSaveToken.Secret), bcrypt.DefaultCost)
	if err != nil {
		return nil, ErrCreateTokenFailed
	}
	toSaveToken.Secret = string(hashedSecret)
	hashedRefreshSecret, err := bcrypt.GenerateFromPassword([]byte(toSaveToken.RefreshSecret), bcrypt.DefaultCost)
	if err != nil {
		return nil, ErrCreateTokenFailed
	}
	toSaveToken.RefreshSecret = string(hashedRefreshSecret)

	if err := s.repo.RefreshAuthToken(ctx, &toSaveToken); err != nil {
		return nil, ErrCreateTokenFailed
	}

	bearer := newToken.ToBearerToken(s.jwtSecret)
	return &bearer, nil
}

func (s *service) GetAuthTokenByJWTToken(ctx context.Context, tokenStr string) (*entities.AuthToken, error) {
	tokenParts, err := s.splitToken(tokenStr)
	if err != nil {
		return nil, ErrInvalidToken
	}

	token, err := s.repo.FindAuthToken(ctx, tokenParts[0])
	if err != nil {
		return nil, ErrInvalidToken
	}

	if err := bcrypt.CompareHashAndPassword([]byte(token.Secret), []byte(tokenParts[1])); err != nil {
		return nil, ErrInvalidToken
	}
	return token, nil
}

func (s *service) GetAuthTokenByJWTRefreshToken(ctx context.Context, tokenStr string) (*entities.AuthToken, error) {
	tokenParts, err := s.splitToken(tokenStr)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	token, err := s.repo.FindAuthToken(ctx, tokenParts[0])
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	if err := bcrypt.CompareHashAndPassword([]byte(token.RefreshSecret), []byte(tokenParts[1])); err != nil {
		return nil, ErrInvalidRefreshToken
	}
	return token, nil
}

func (s *service) AuthByToken(ctx context.Context, tokenStr string) (*entities.User, error) {
	token, err := s.GetAuthTokenByJWTToken(ctx, tokenStr)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByID(ctx, token.UserID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *service) AuthByRefreshToken(ctx context.Context, tokenStr string) (*entities.User, error) {
	token, err := s.GetAuthTokenByJWTRefreshToken(ctx, tokenStr)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByID(ctx, token.UserID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *service) splitToken(tokenStr string) ([]string, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 2 {
		return nil, errors.New("Invalid token")
	}
	return parts, nil
}

func (s *service) toBearerToken(authToken *entities.AuthToken) entities.BearerToken {
	return authToken.ToBearerToken(s.jwtSecret)
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
	case TokenActionActivation:
		subject = "Activation"
		url = fmt.Sprintf(
			"%s/activate?token=%s", s.config.BASE_URL, tokenWithSecret,
		)
		message = fmt.Sprintf("The activation link: %s", url)
	case TokenActionResetActivation:
		subject = "Activation"
		url = fmt.Sprintf(
			"%s/activate?token=%s", s.config.BASE_URL, tokenWithSecret,
		)
		message = fmt.Sprintf("The reset activation link: %s", url)
	case TokenActionResetPassword:
		subject = "Reset Password"
		url = fmt.Sprintf(
			"%s/reset-password?token=%s", s.config.BASE_URL, tokenWithSecret,
		)
		message = fmt.Sprintf("The reset password link: %s", url)
	default:
		return errors.New("Invalid token action")
	}

	task, err := tasks.SendEmailTasks(ctx, s.config.SMTP_USERNAME, email, subject, message)
	if err != nil {
		log.Error().Err(err).Str("task_type", tasks.TypeAuthEmail).
			Msg("Fail to create task")
		return tasks.ErrTokenActionCreateTask
	} else if !s.skipTaskQueue {
		_ = s.EnqueueTask(task, actionType)
	}
	return nil

}

func (s *service) EnqueueTask(task *asynq.Task, taskType TokenAction) error {
	taskInfo, err := s.asynqClient.Enqueue(task)
	log.Debug().Any("task_info", taskInfo).Err(err).Send()
	if err != nil {
		log.Error().Str("task_type", string(taskType)).Err(err).
			Msg("Fail to enqueue task")
		return tasks.ErrTokenActionEnqueueTask
	}
	return nil
}

// For testing purposes only
func (s *service) SetSkipTaskQueue(v bool) {
	s.skipTaskQueue = v
}
