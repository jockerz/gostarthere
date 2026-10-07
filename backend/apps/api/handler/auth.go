package handler

import (
	"context"
	"errors"

	"github.com/danielgtaylor/huma/v2"
	"github.com/rs/zerolog/log"

	"vnti/apps/api/middleware"
	"vnti/apps/api/presenter"
	"vnti/apps/api/schema"
	"vnti/pkg/entities"
	"vnti/pkg/service"
)

func toUserResponse(user *entities.User) presenter.UserResponse {
	return presenter.UserResponse{
		ID:       user.ID,
		Email:    user.Email,
		Username: user.Username,
		Name:     user.Name,
		Avatar:   user.Avatar,
		Active:   user.Active,
	}
}

func Logout(svc service.AuthService) func(context.Context, *schema.LogoutInput) (*presenter.SuccessResponse, error) {
	return func(ctx context.Context, input *schema.LogoutInput) (*presenter.SuccessResponse, error) {
		err := svc.Logout(ctx, input.Body.Token)
		if err != nil {
			if errors.Is(err, service.ErrAuthInvalidToken) {
				return nil, huma.Error400BadRequest(err.Error())
			}
			return nil, huma.Error500InternalServerError("internal server error")
		}
		return &presenter.SuccessResponse{Body: presenter.SuccessBody{
			Success: true,
			Message: "Logged out successfully",
		}}, nil
	}
}

func RefreshToken(svc service.AuthService) func(context.Context, *schema.RefreshTokenInput) (*presenter.SuccessResponse, error) {
	return func(ctx context.Context, input *schema.RefreshTokenInput) (*presenter.SuccessResponse, error) {
		tokens, err := svc.RefreshToken(ctx, input.Body.RefreshToken)
		if err != nil {
			if errors.Is(err, service.ErrAuthInvalidToken) {
				return nil, huma.Error401Unauthorized(err.Error())
			}
			return nil, huma.Error500InternalServerError("internal server error")
		}
		return &presenter.SuccessResponse{Body: presenter.SuccessBody{
			Success: true,
			Message: "Token refreshed successfully",
			Data:    tokens,
		}}, nil
	}
}

func Login(svc service.AuthService) func(context.Context, *schema.LoginInput) (*presenter.SuccessResponse, error) {
	return func(ctx context.Context, input *schema.LoginInput) (*presenter.SuccessResponse, error) {
		_, token, err := svc.Login(ctx, input.Body.Email, input.Body.Password)
		if err != nil {
			if errors.Is(err, service.ErrAuthInvalidCredentials) {
				return nil, huma.Error401Unauthorized(err.Error())
			}
			return nil, huma.Error500InternalServerError("internal server error")
		}
		return &presenter.SuccessResponse{Body: presenter.SuccessBody{
			Success: true,
			Message: "Login successful",
			Data:    token,
		}}, nil
	}
}

func Register(svc service.AuthService) func(context.Context, *schema.RegisterInput) (*presenter.SuccessResponse, error) {
	return func(ctx context.Context, input *schema.RegisterInput) (*presenter.SuccessResponse, error) {
		registerReq := &entities.Register{
			Email:            input.Body.Email,
			Username:         input.Body.Username,
			Name:             input.Body.Name,
			Password:         input.Body.Password,
			PasswordValidate: input.Body.PasswordValidate,
		}

		user, err := svc.Register(ctx, registerReq)
		if err != nil {
			if errors.Is(err, service.ErrAuthEmailAlreadyExists) || errors.Is(err, service.ErrAuthUsernameAlreadyExists) {
				return nil, huma.Error409Conflict(err.Error())
			}
			return nil, huma.Error422UnprocessableEntity(err.Error())
		}
		log.Debug().Any("user", user).Msg("New user registered")

		return &presenter.SuccessResponse{Body: presenter.SuccessBody{
			Success: true,
			Message: "Registration successful",
		}}, nil
	}
}

func Activate(svc service.AuthService) func(context.Context, *schema.ActivateInput) (*presenter.SuccessResponse, error) {
	return func(ctx context.Context, input *schema.ActivateInput) (*presenter.SuccessResponse, error) {
		err := svc.ActivateAccount(ctx, input.Token)
		if err != nil {
			if errors.Is(err, service.ErrAuthInvalidToken) {
				return nil, huma.Error400BadRequest(err.Error())
			} else if errors.Is(err, service.ErrAuthUsedToken) {
				return nil, huma.Error400BadRequest(err.Error())
			}
			return nil, huma.Error500InternalServerError("internal server error")
		}
		return &presenter.SuccessResponse{Body: presenter.SuccessBody{
			Success: true,
			Message: "Account activated successfully",
		}}, nil
	}
}

func ResendActivation(svc service.AuthService) func(context.Context, *schema.ResetActivationInput) (*presenter.SuccessResponse, error) {
	return func(ctx context.Context, input *schema.ResetActivationInput) (*presenter.SuccessResponse, error) {
		err := svc.ResendActivation(ctx, input.Body.Email)
		if err != nil {
			// TODO: log err
			// return nil, huma.Error500InternalServerError("internal server error")
		}
		return &presenter.SuccessResponse{Body: presenter.SuccessBody{
			Success: true,
			Message: "If the email exists, a new activation token has been sent",
		}}, nil
	}
}

func ForgotPassword(svc service.AuthService) func(context.Context, *schema.ForgotPasswordInput) (*presenter.SuccessResponse, error) {
	return func(ctx context.Context, input *schema.ForgotPasswordInput) (*presenter.SuccessResponse, error) {
		err := svc.ForgotPassword(ctx, input.Body.Email)
		if err != nil {
			return nil, huma.Error500InternalServerError("internal server error")
		}
		return &presenter.SuccessResponse{Body: presenter.SuccessBody{
			Success: true,
			Message: "If the email exists, a reset token has been sent",
		}}, nil
	}
}

func ResetPassword(svc service.AuthService) func(context.Context, *schema.ResetPasswordInput) (*presenter.SuccessResponse, error) {
	return func(ctx context.Context, input *schema.ResetPasswordInput) (*presenter.SuccessResponse, error) {
		err := svc.ResetPassword(ctx, input.Body.Token, input.Body.Password)
		if err != nil {
			if errors.Is(err, service.ErrAuthInvalidToken) {
				return nil, huma.Error400BadRequest(err.Error())
			} else if errors.Is(err, service.ErrAuthUsedToken) {
				return nil, huma.Error400BadRequest(err.Error())
			}
			return nil, huma.Error500InternalServerError("internal server error")
		}
		return &presenter.SuccessResponse{Body: presenter.SuccessBody{
			Success: true,
			Message: "Password reset successfully",
		}}, nil
	}
}

func GetAuthData(oauthSrv service.OAuthService) func(context.Context, *schema.GetAuthProviderInput) (*presenter.SuccessResponse, error) {
	return func(ctx context.Context, _ *schema.GetAuthProviderInput) (*presenter.SuccessResponse, error) {
		current_user := ctx.Value(middleware.CtxUserKey).(*entities.User)
		userAuthProviderData := oauthSrv.GetUserAuthProviderData(ctx, current_user.ID)
		return &presenter.SuccessResponse{Body: presenter.SuccessBody{
			Success: true,
			Message: "User Authentication provider data",
			Data: &presenter.AuthenticationData{
				AuthProviderData: userAuthProviderData,
				HasPassword:      current_user.HasPassword(),
			},
		}}, nil
	}
}
