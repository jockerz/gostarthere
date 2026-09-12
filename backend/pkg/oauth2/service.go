package oauth2

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"vnti/internal"
	"vnti/pkg/auth"
	"vnti/pkg/entities"
	"vnti/pkg/user"
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

var (
	ErrProviderNotConfigured = errors.New("oauth provider not configured")
	ErrInvalidState          = errors.New("invalid or expired state")
	ErrStateProviderMismatch = errors.New("state provider mismatch")
	ErrUnsupportedProvider   = errors.New("unsupported oauth provider")
	ErrEmailNotProvided      = errors.New("could not retrieve email from provider")
	ErrUserNotFound          = errors.New("user not found")
	ErrProviderAlreadyLinked = errors.New("provider already linked to a different user")
	ErrTokenExchange         = errors.New("failed to exchange oauth code")
	ErrFetchUserInfo         = errors.New("failed to fetch user info from provider")
	ErrCreateUserFailed      = errors.New("failed to create user")
	ErrLinkProviderFailed    = errors.New("failed to link oauth provider")
	ErrGenerateTokenFailed   = errors.New("failed to generate auth token")
)

type Service interface {
	BuildAuthorizeURL(ctx context.Context, provider, codeChallenge string, userID *uint) (url string, state string, err error)
	HandleCallback(ctx context.Context, provider, code, state, codeVerifier string, authUserID *uint) (accessToken string, refreshToken string, user *entities.User, isNew bool, err error)
	GetUserAuthProviderData(context.Context, uint) []*entities.UserAuthProvider
}

type service struct {
	config     *internal.Config
	stateStore *StateStore
	userRepo   user.Repository
	authSvc    auth.Service
	oauthRepo  Repository
	jwtSecret  string
}

func NewService(
	config *internal.Config,
	stateStore *StateStore,
	userRepo user.Repository,
	authSvc auth.Service,
	oauthRepo Repository,
) Service {
	return &service{
		config:     config,
		stateStore: stateStore,
		userRepo:   userRepo,
		authSvc:    authSvc,
		oauthRepo:  oauthRepo,
		jwtSecret:  config.SECRET,
	}
}

func (s service) BuildRedirectURL(provider string) string {
	return fmt.Sprintf("%s/auth/oauth/callback?provider=%s", s.config.BASE_URL, provider)
}

func (s *service) BuildAuthorizeURL(ctx context.Context, provider, codeChallenge string, userID *uint) (string, string, error) {
	redirectURI := s.BuildRedirectURL(provider)
	state := s.stateStore.Generate(provider, userID, codeChallenge)

	switch provider {
	case "google":
		if s.config.OAuthGoogleClientID == "" {
			return "", "", fmt.Errorf("%w: google", ErrProviderNotConfigured)
		}
		params := url.Values{
			"client_id":             {s.config.OAuthGoogleClientID},
			"redirect_uri":          {redirectURI},
			"response_type":         {"code"},
			"scope":                 {"email profile"},
			"state":                 {state},
			"code_challenge":        {codeChallenge},
			"code_challenge_method": {"S256"},
		}
		return "https://accounts.google.com/o/oauth2/v2/auth?" + params.Encode(), state, nil

	case "github":
		if s.config.OAuthGitHubClientID == "" {
			return "", "", fmt.Errorf("%w: github", ErrProviderNotConfigured)
		}
		params := url.Values{
			"client_id":    {s.config.OAuthGitHubClientID},
			"redirect_uri": {redirectURI},
			"scope":        {"user:email"},
			"state":        {state},
		}
		return "https://github.com/login/oauth/authorize?" + params.Encode(), state, nil

	default:
		return "", "", fmt.Errorf("%w: %s", ErrUnsupportedProvider, provider)
	}
}

func (s *service) HandleCallback(ctx context.Context, provider string, code string, state string, codeVerifier string, authUserID *uint) (string, string, *entities.User, bool, error) {
	entry := s.stateStore.Consume(state)
	if entry == nil {
		return "", "", nil, false, ErrInvalidState
	}
	if entry.Provider != provider {
		return "", "", nil, false, ErrStateProviderMismatch
	}
	if entry.UserID != nil && (authUserID == nil || *authUserID != *entry.UserID) {
		return "", "", nil, false, ErrInvalidState
	}

	var providerUserID string
	var email string
	var name string
	var err error

	switch provider {
	case "google":
		var tokenResp *googleTokenResponse
		tokenResp, err = s.exchangeGoogleCode(ctx, code, codeVerifier)
		if err != nil {
			return "", "", nil, false, err
		}

		var userInfo *googleUserInfo
		userInfo, err = s.fetchGoogleUserInfo(ctx, tokenResp.AccessToken)
		if err != nil {
			return "", "", nil, false, err
		}

		providerUserID = userInfo.ID
		email = userInfo.Email
		name = userInfo.Name

	case "github":
		var tokenResp *githubTokenResponse
		tokenResp, err = s.exchangeGitHubCode(ctx, code)
		if err != nil {
			return "", "", nil, false, err
		}

		var userInfo *githubUserInfo
		userInfo, err = s.fetchGitHubUserInfo(ctx, tokenResp.AccessToken)
		if err != nil {
			return "", "", nil, false, err
		}

		providerUserID = fmt.Sprintf("%d", userInfo.ID)
		name = userInfo.Name

		if userInfo.Email != "" {
			email = userInfo.Email
		} else {
			var emails []githubEmail
			emails, err = s.fetchGitHubEmails(ctx, tokenResp.AccessToken)
			if err != nil {
				return "", "", nil, false, err
			}
			email = primaryEmail(emails)
		}

	default:
		return "", "", nil, false, fmt.Errorf("%w: %s", ErrUnsupportedProvider, provider)
	}

	if email == "" {
		return "", "", nil, false, ErrEmailNotProvided
	}

	var linkedUserID uint
	isNew := false

	existingProv, findErr := s.oauthRepo.FindByProvider(ctx, entities.AuthProvider(provider), providerUserID)
	if findErr != nil {
		// Not linked yet
		var user *entities.User

		if authUserID != nil {
			// Linking to existing authenticated user
			user, err = s.userRepo.FindByID(ctx, *authUserID)
			if err != nil {
				return "", "", nil, false, ErrUserNotFound
			}
		} else {
			// New OAuth login — find or create user
			user, err = s.userRepo.FindByEmail(ctx, email)
			if err != nil {
				username := generateUsername(email)
				user = &entities.User{
					Email:     email,
					Username:  username,
					Name:      name,
					Password:  nil,
					Active:    true,
					CreatedAt: time.Now(),
				}
				user, err = s.userRepo.Create(ctx, user)
				if err != nil {
					return "", "", nil, false, fmt.Errorf("%w: %w", ErrCreateUserFailed, err)
				}
				isNew = true
			}
		}

		oauthEntry := &entities.UserAuthProvider{
			UserID:         user.ID,
			Provider:       entities.AuthProvider(provider),
			ProviderUserID: providerUserID,
			Email:          email,
		}
		if err := s.oauthRepo.Create(ctx, oauthEntry); err != nil {
			return "", "", nil, false, fmt.Errorf("%w: %w", ErrLinkProviderFailed, err)
		}

		linkedUserID = user.ID
	} else {
		// Already linked
		if authUserID != nil && *authUserID != existingProv.UserID {
			return "", "", nil, false, ErrProviderAlreadyLinked
		}
		linkedUserID = existingProv.UserID
	}

	user, err := s.userRepo.FindByID(ctx, linkedUserID)
	if err != nil {
		return "", "", nil, false, ErrUserNotFound
	}

	authToken, err := s.authSvc.GenerateAuthToken(ctx, user)
	if err != nil {
		return "", "", nil, false, ErrGenerateTokenFailed
	}

	bearer := authToken.ToBearerToken(s.jwtSecret)
	return bearer.AccessToken, bearer.RefreshToken, user, isNew, nil
}

func (s *service) GetUserAuthProviderData(ctx context.Context, userId uint) []*entities.UserAuthProvider {
	entries, err := s.oauthRepo.FindByUserID(ctx, userId)
	if err != nil {
		return []*entities.UserAuthProvider{}
	}
	return entries
}

func primaryEmail(emails []githubEmail) string {
	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email
		}
	}
	if len(emails) > 0 {
		return emails[0].Email
	}
	return ""
}

func (s *service) exchangeGoogleCode(ctx context.Context, code, codeVerifier string) (*googleTokenResponse, error) {
	redirectURI := s.BuildRedirectURL("google")
	data := url.Values{
		"code":          {code},
		"client_id":     {s.config.OAuthGoogleClientID},
		"client_secret": {s.config.OAuthGoogleClientSecret},
		"redirect_uri":  {redirectURI},
		"grant_type":    {"authorization_code"},
		"code_verifier": {codeVerifier},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://oauth2.googleapis.com/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTokenExchange, err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTokenExchange, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTokenExchange, err)
	}

	var tokenResp googleTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTokenExchange, err)
	}

	if tokenResp.AccessToken == "" {
		var errResp googleTokenError
		if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
			return nil, fmt.Errorf("%w: %s - %s", ErrTokenExchange, errResp.Error, errResp.ErrorDescription)
		}
		return nil, ErrTokenExchange
	}

	return &tokenResp, nil
}

func (s *service) fetchGoogleUserInfo(ctx context.Context, accessToken string) (*googleUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFetchUserInfo, err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFetchUserInfo, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFetchUserInfo, err)
	}

	var userInfo googleUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFetchUserInfo, err)
	}
	return &userInfo, nil
}

func (s *service) exchangeGitHubCode(ctx context.Context, code string) (*githubTokenResponse, error) {
	redirectURI := s.BuildRedirectURL("github")
	data := url.Values{
		"code":          {code},
		"client_id":     {s.config.OAuthGitHubClientID},
		"client_secret": {s.config.OAuthGitHubClientSecret},
		"redirect_uri":  {redirectURI},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://github.com/login/oauth/access_token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTokenExchange, err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTokenExchange, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTokenExchange, err)
	}

	var tokenResp githubTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTokenExchange, err)
	}

	if tokenResp.Error != "" {
		return nil, fmt.Errorf("%w: %s - %s", ErrTokenExchange, tokenResp.Error, tokenResp.ErrorDesc)
	}

	if tokenResp.AccessToken == "" {
		return nil, ErrTokenExchange
	}

	return &tokenResp, nil
}

func (s *service) fetchGitHubUserInfo(ctx context.Context, accessToken string) (*githubUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFetchUserInfo, err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFetchUserInfo, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFetchUserInfo, err)
	}

	var userInfo githubUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFetchUserInfo, err)
	}
	return &userInfo, nil
}

func (s *service) fetchGitHubEmails(ctx context.Context, accessToken string) ([]githubEmail, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFetchUserInfo, err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFetchUserInfo, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFetchUserInfo, err)
	}

	var emails []githubEmail
	if err := json.Unmarshal(body, &emails); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFetchUserInfo, err)
	}
	return emails, nil
}

func generateUsername(email string) string {
	parts := strings.Split(email, "@")
	username := strings.ToLower(parts[0])
	re := regexp.MustCompile(`[^a-z0-9]`)
	username = re.ReplaceAllString(username, "")
	if username == "" {
		username = fmt.Sprintf("user%d", time.Now().UnixMilli())
	}
	if len(username) < 3 {
		username = username + strings.Repeat("0", 3-len(username))
	}
	if len(username) > 32 {
		username = username[:32]
	}
	return username
}
