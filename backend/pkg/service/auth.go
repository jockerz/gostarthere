package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"vnti/internal"
	"vnti/internal/logger"
	"vnti/pkg/entities"
	"vnti/pkg/repository"
	"vnti/pkg/tasks"

	"github.com/gofiber/utils/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrAuthInvalidCredentials     = errors.New("invalid credentials")
	ErrAuthInvalidToken           = errors.New("invalid or expired token")
	ErrAuthUsedToken              = errors.New("token has been used")
	ErrAuthEmailAlreadyExists     = errors.New("email already exists")
	ErrAuthUsernameAlreadyExists  = errors.New("username already exists")
	ErrAuthCreateTokenFailed      = errors.New("create auth token failed")
	ErrAuthUserActivationFailed   = errors.New("user activation failed")
	ErrAuthResendActivationFailed = errors.New("user activation reset failed")
	ErrAuthInvalidRefreshToken    = errors.New("invalid or expired token")
	ErrAuthRefreshTokenExpired    = errors.New("reset token expired")
	ErrAuthUserNotFound           = errors.New("user not found")
)

type AuthTokenAction string

const (
	TokenActionActivation      AuthTokenAction = "activation"
	TokenActionResetActivation AuthTokenAction = "reset_activation"
	TokenActionResetPassword   AuthTokenAction = "reset_password"
)

type AuthService interface {
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

type AuthClaims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

type authServiceImpl struct {
	config    *internal.Config
	userRepo  repository.UserRepository
	repo      repository.AuthRepository
	jwtSecret string

	asynqClient   *asynq.Client
	skipTaskQueue bool
}

func NewAuthService(
	config *internal.Config,
	userRepo repository.UserRepository,
	repo repository.AuthRepository,
	jwtSecret string,
	asynqClient *asynq.Client,
) AuthService {
	return &authServiceImpl{
		config:      config,
		userRepo:    userRepo,
		repo:        repo,
		jwtSecret:   jwtSecret,
		asynqClient: asynqClient,
	}
}

func (s *authServiceImpl) GenerateUserToken(ctx context.Context, user_id uint, token_type entities.UserTokenType) (*entities.UserToken, error) {
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
		return nil, ErrAuthCreateTokenFailed
	}
	to_save_token.Secret = string(hashed_secret)
	hashed_refresh_secret, err := bcrypt.GenerateFromPassword([]byte(to_save_token.RefreshSecret), bcrypt.DefaultCost)
	if err != nil {
		return nil, ErrAuthCreateTokenFailed
	}
	to_save_token.RefreshSecret = string(hashed_refresh_secret)

	_, err = s.repo.CreateUserToken(ctx, &to_save_token)
	if err != nil {
		return nil, ErrAuthCreateTokenFailed
	}
	return &token, nil
}

func (s *authServiceImpl) GenerateAuthToken(ctx context.Context, user *entities.User) (*entities.AuthToken, error) {
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
		return nil, ErrAuthCreateTokenFailed
	}
	to_save_token.Secret = string(hashed_secret)
	hashed_refresh_secret, err := bcrypt.GenerateFromPassword([]byte(to_save_token.RefreshSecret), bcrypt.DefaultCost)
	if err != nil {
		return nil, ErrAuthCreateTokenFailed
	}
	to_save_token.RefreshSecret = string(hashed_refresh_secret)

	_, err = s.repo.CreateAuthToken(ctx, &to_save_token)
	if err != nil {
		return nil, ErrAuthCreateTokenFailed
	}

	l := logger.LogContext(ctx, "auth:service:GenAuthToken", "AuthToken created")
	l.Debug().Str("tokenPrefix", token.Prefix).Msg("User login")

	return &token, nil
}

func (s *authServiceImpl) Login(ctx context.Context, email_or_username, password string) (*entities.User, *entities.BearerToken, error) {
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
		return nil, nil, ErrAuthInvalidCredentials
	}

	if user.Password == nil || !s.checkTokenSecret(*user.Password, password) {
		log.Debug().Str("username/email", email_or_username).Msg("Invalid password")
		return nil, nil, ErrAuthInvalidCredentials
	}

	auth_token, err := s.GenerateAuthToken(ctx, user)
	if err != nil {
		return nil, nil, err
	}

	bearer_token := s.toBearerToken(auth_token)

	return user, &bearer_token, nil
}

func (s *authServiceImpl) Register(ctx context.Context, input *entities.Register) (*entities.User, error) {
	if input.Password != input.PasswordValidate {
		return nil, errors.New("passwords do not match")
	}

	existing, _ := s.userRepo.FindByEmail(ctx, input.Email)
	if existing != nil {
		return nil, ErrAuthEmailAlreadyExists
	}

	existing, _ = s.userRepo.FindByUsername(ctx, input.Username)
	if existing != nil {
		return nil, ErrAuthUsernameAlreadyExists
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

func (s *authServiceImpl) ActivateAccount(ctx context.Context, tokenStr string) error {
	tokenParts, err := s.splitToken(tokenStr)
	if err != nil {
		return ErrAuthInvalidToken
	}

	token, err := s.repo.FindUserToken(ctx, tokenParts[0], entities.TokenActivation)
	if err != nil {
		return ErrAuthInvalidToken
	} else if token.IsUsed {
		return ErrAuthUsedToken
	} else if token.ExpiresAt.Before(time.Now()) {
		return ErrAuthInvalidToken
	} else if !s.checkTokenSecret(token.Secret, tokenParts[1]) {
		return ErrAuthInvalidToken
	}

	user, err := s.userRepo.FindByID(ctx, token.UserID)
	if err != nil {
		// TODO: log, because token is found but the user is not
		return ErrAuthInvalidToken
	}

	user.Active = true
	// _, err = s.userRepo.Update(ctx, user)
	err = s.userRepo.UpdateByID(ctx, user.ID, map[string]any{"active": true}, user.ID)
	if err != nil {
		// TODO: log error
		return ErrAuthUserActivationFailed
	}

	_ = s.repo.MarkAsUsedUserToken(ctx, token.ID)
	return nil
}

func (s *authServiceImpl) ResendActivation(ctx context.Context, email_or_username string) error {
	var err error
	var user *entities.User

	if strings.Contains(email_or_username, "@") {
		user, err = s.userRepo.FindByEmail(ctx, email_or_username)
	} else {
		user, err = s.userRepo.FindByUsername(ctx, email_or_username)
	}
	if err != nil {
		// TODO: log because, token found with invalid user
		return ErrAuthUserNotFound
	}

	token, err := s.GenerateUserToken(ctx, user.ID, entities.TokenActivation)
	if err != nil {
		// TODO: error
		return ErrAuthResendActivationFailed
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

func (s *authServiceImpl) ForgotPassword(ctx context.Context, email_or_username string) error {
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

func (s *authServiceImpl) ResetPassword(ctx context.Context, tokenStr, newPassword string) error {
	tokenParts, err := s.splitToken(tokenStr)
	if err != nil {
		return ErrAuthInvalidToken
	}

	token, err := s.repo.FindUserToken(ctx, tokenParts[0], entities.TokenReset)
	if err != nil {
		return ErrAuthInvalidToken
	} else if token.IsUsed {
		return ErrAuthUsedToken
	} else if token.ExpiresAt.Before(time.Now()) {
		return ErrAuthInvalidToken
	} else if !s.checkTokenSecret(token.Secret, tokenParts[1]) {
		return ErrAuthInvalidToken
	}

	user, err := s.userRepo.FindByID(ctx, token.UserID)
	if err != nil {
		// TODO: log, because token is found but the user is not
		return ErrAuthInvalidToken
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

func (s *authServiceImpl) Logout(ctx context.Context, tokenStr string) error {
	tokenParts, err := s.splitToken(tokenStr)
	if err != nil {
		return ErrAuthInvalidToken
	}

	token, err := s.repo.FindAuthToken(ctx, tokenParts[0])
	if err != nil {
		return ErrAuthInvalidToken
	}

	if err := bcrypt.CompareHashAndPassword([]byte(token.Secret), []byte(tokenParts[1])); err != nil {
		return ErrAuthInvalidToken
	}

	l := logger.LogContext(ctx, "auth:service:Logout", "logout")
	l.Debug().Str("tokenPrefix", tokenParts[0]).Msg("User logout")

	return s.repo.RevokeAuthToken(ctx, token.Prefix)
}

func (s *authServiceImpl) RefreshToken(ctx context.Context, refreshTokenStr string) (*entities.BearerToken, error) {
	tokenParts, err := s.splitToken(refreshTokenStr)
	if err != nil {
		return nil, ErrAuthInvalidToken
	}

	token, err := s.repo.FindAuthToken(ctx, tokenParts[0])
	if err != nil {
		return nil, ErrAuthInvalidToken
	}

	if token.IsRevoked {
		return nil, ErrAuthInvalidToken
	}

	if token.RefreshExpiresAt.Before(time.Now()) {
		return nil, ErrAuthRefreshTokenExpired
	}

	if err := bcrypt.CompareHashAndPassword([]byte(token.RefreshSecret), []byte(tokenParts[1])); err != nil {
		return nil, ErrAuthInvalidToken
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
		return nil, ErrAuthCreateTokenFailed
	}
	toSaveToken.Secret = string(hashedSecret)
	hashedRefreshSecret, err := bcrypt.GenerateFromPassword([]byte(toSaveToken.RefreshSecret), bcrypt.DefaultCost)
	if err != nil {
		return nil, ErrAuthCreateTokenFailed
	}
	toSaveToken.RefreshSecret = string(hashedRefreshSecret)

	if err := s.repo.RefreshAuthToken(ctx, &toSaveToken); err != nil {
		return nil, ErrAuthCreateTokenFailed
	}

	bearer := newToken.ToBearerToken(s.jwtSecret)
	return &bearer, nil
}

func (s *authServiceImpl) GetAuthTokenByJWTToken(ctx context.Context, tokenStr string) (*entities.AuthToken, error) {
	tokenParts, err := s.splitToken(tokenStr)
	if err != nil {
		return nil, ErrAuthInvalidToken
	}

	token, err := s.repo.FindAuthToken(ctx, tokenParts[0])
	if err != nil {
		return nil, ErrAuthInvalidToken
	}

	if err := bcrypt.CompareHashAndPassword([]byte(token.Secret), []byte(tokenParts[1])); err != nil {
		return nil, ErrAuthInvalidToken
	}
	return token, nil
}

func (s *authServiceImpl) GetAuthTokenByJWTRefreshToken(ctx context.Context, tokenStr string) (*entities.AuthToken, error) {
	tokenParts, err := s.splitToken(tokenStr)
	if err != nil {
		return nil, ErrAuthInvalidRefreshToken
	}

	token, err := s.repo.FindAuthToken(ctx, tokenParts[0])
	if err != nil {
		return nil, ErrAuthInvalidRefreshToken
	}

	if err := bcrypt.CompareHashAndPassword([]byte(token.RefreshSecret), []byte(tokenParts[1])); err != nil {
		return nil, ErrAuthInvalidRefreshToken
	}
	return token, nil
}

func (s *authServiceImpl) AuthByToken(ctx context.Context, tokenStr string) (*entities.User, error) {
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

func (s *authServiceImpl) AuthByRefreshToken(ctx context.Context, tokenStr string) (*entities.User, error) {
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

func (s *authServiceImpl) splitToken(tokenStr string) ([]string, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 2 {
		return nil, errors.New("Invalid token")
	}
	return parts, nil
}

func (s *authServiceImpl) toBearerToken(authToken *entities.AuthToken) entities.BearerToken {
	return authToken.ToBearerToken(s.jwtSecret)
}

func (s *authServiceImpl) checkTokenSecret(hashed, plain string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain)); err != nil {
		return false
	}
	return true
}

func (s *authServiceImpl) sendTokenEmail(ctx context.Context, actionType AuthTokenAction, tokenWithSecret, email string) error {
	var url, message, subject string

	switch actionType {
	case TokenActionActivation:
		subject = "Activation"
		// Frontend URL
		url = fmt.Sprintf(
			"%s/auth/activate?token=%s", s.config.BASE_URL, tokenWithSecret,
		)
		message = fmt.Sprintf("The activation link: %s", url)
	case TokenActionResetActivation:
		subject = "Activation"
		// Frontend URL
		url = fmt.Sprintf(
			"%s/auth/activate?token=%s", s.config.BASE_URL, tokenWithSecret,
		)
		message = fmt.Sprintf("The reset activation link: %s", url)
	case TokenActionResetPassword:
		subject = "Reset Password"
		// Frontend URL
		url = fmt.Sprintf(
			"%s/auth/reset-password?token=%s", s.config.BASE_URL, tokenWithSecret,
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

func (s *authServiceImpl) EnqueueTask(task *asynq.Task, taskType AuthTokenAction) error {
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
func (s *authServiceImpl) SetSkipTaskQueue(v bool) {
	s.skipTaskQueue = v
}
