package handler

import (
	"context"
	"errors"
	"fmt"

	"github.com/danielgtaylor/huma/v2"
	"github.com/golang-jwt/jwt/v5"

	"vnti/apps/api/middleware"
	"vnti/apps/api/presenter"
	"vnti/apps/api/schema"
	"vnti/internal"
	"vnti/pkg/auth"
	"vnti/pkg/entities"
	"vnti/pkg/oauth2"
	"vnti/pkg/user"
)

const bearerPrefix = "Bearer "

func OAuthAuthorize(config *internal.Config, svc oauth2.Service, authSvc auth.Service) func(context.Context, *schema.OAuthAuthorizeInput) (*schema.OAuthAuthorizeOutput, error) {
	return func(ctx context.Context, input *schema.OAuthAuthorizeInput) (*schema.OAuthAuthorizeOutput, error) {
		var authUserID *uint
		fmt.Printf("[OAuthAuthorize] input: %v\n", input)

		if input.Authorization != "" {
			jwtToken, err := parseBearerToken(input.Authorization)
			if err != nil {
				return nil, huma.Error401Unauthorized("invalid authorization header")
			}
			userID, err := resolveUserFromJWT(ctx, jwtToken, config.SECRET, authSvc)
			if err != nil {
				return nil, huma.Error401Unauthorized("invalid or expired token")
			}
			authUserID = &userID
		}

		fmt.Printf("*authUserID  : %d\n", authUserID)
		url, state, err := svc.BuildAuthorizeURL(ctx, input.Provider, input.CodeChallenge, authUserID)
		fmt.Printf("url  : %s\n", url)
		fmt.Printf("state: %s\n", state)
		fmt.Printf("err  : %v\n", err)
		if err != nil {
			return nil, huma.Error500InternalServerError(err.Error())
		}

		return &schema.OAuthAuthorizeOutput{
			Body: struct {
				URL   string `json:"url"`
				State string `json:"state"`
			}{URL: url, State: state},
		}, nil
	}
}

func OAuthCallback(config *internal.Config, svc oauth2.Service, authSvc auth.Service) func(context.Context, *schema.OAuthCallbackInput) (*schema.OAuthCallbackOutput, error) {
	return func(ctx context.Context, input *schema.OAuthCallbackInput) (*schema.OAuthCallbackOutput, error) {
		var authUserID *uint

		if input.Authorization != "" {
			jwtToken, err := parseBearerToken(input.Authorization)
			if err != nil {
				return nil, huma.Error401Unauthorized("invalid authorization header")
			}
			userID, err := resolveUserFromJWT(ctx, jwtToken, config.SECRET, authSvc)
			if err != nil {
				return nil, huma.Error401Unauthorized("invalid or expired token")
			}
			authUserID = &userID
		}

		accessToken, refreshToken, user, isNew, err := svc.HandleCallback(
			ctx, input.Provider, input.Body.Code, input.Body.State, input.Body.CodeVerifier, authUserID,
		)
		if err != nil {
			if errors.Is(err, oauth2.ErrInvalidState) {
				return nil, huma.Error400BadRequest(err.Error())
			}
			if errors.Is(err, oauth2.ErrProviderAlreadyLinked) {
				return nil, huma.Error409Conflict(err.Error())
			}
			return nil, huma.Error500InternalServerError(err.Error())
		}

		resp := &schema.OAuthCallbackOutput{}
		resp.Body.Success = true
		resp.Body.Message = "OAuth login successful"
		resp.Body.Data = &presenter.OAuthCallbackData{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			User:         toUserResponse(user),
			IsNewUser:    isNew,
		}
		return resp, nil
	}
}

func SetPassword(svc user.Service) func(context.Context, *schema.SetPasswordInput) (*presenter.SuccessResponse, error) {
	return func(ctx context.Context, input *schema.SetPasswordInput) (*presenter.SuccessResponse, error) {
		currentUser := ctx.Value(middleware.CtxUserKey).(*entities.User)

		if currentUser.HasPassword() {
			return nil, huma.Error400BadRequest("user already has a password")
		}

		err := svc.SetPassword(ctx, currentUser.ID, input.Body.Password)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to set password")
		}

		return &presenter.SuccessResponse{Body: presenter.SuccessBody{
			Success: true,
			Message: "Password set successfully",
		}}, nil
	}
}

func parseBearerToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("missing authorization header")
	}
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		return "", errors.New("invalid authorization header format")
	}
	return authHeader[len(bearerPrefix):], nil
}

func resolveUserFromJWT(ctx context.Context, jwtTokenStr string, jwtSecret string, authSvc auth.Service) (uint, error) {
	claims := entities.JWTClaims{}
	token, err := jwt.ParseWithClaims(jwtTokenStr, &claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return 0, errors.New("invalid token")
	}

	tokenStr := token.Claims.(*entities.JWTClaims).ID
	user, err := authSvc.AuthByToken(ctx, tokenStr)
	if err != nil {
		return 0, err
	}

	return user.ID, nil
}
