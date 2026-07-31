package handler

import (
	"context"
	"errors"
	"io"

	"github.com/danielgtaylor/huma/v2"
	"github.com/rs/zerolog/log"

	"vnti/apps/api/middleware"
	"vnti/apps/api/presenter"
	"vnti/apps/api/schema"
	"vnti/pkg/entities"
	"vnti/pkg/user"
)

func GetProfile(svc user.Service) func(context.Context, *schema.GetProfileInput) (*presenter.SuccessResponse, error) {
	return func(ctx context.Context, _ *schema.GetProfileInput) (*presenter.SuccessResponse, error) {
		current_user := ctx.Value(middleware.CtxUserKey).(*entities.User)

		return &presenter.SuccessResponse{Body: presenter.SuccessBody{
			Success: true,
			Message: "Profile",
			Data:    toUserResponse(current_user),
		}}, nil
	}
}

func UpdateProfile(svc user.Service) func(context.Context, *schema.UpdateProfileInput) (*presenter.SuccessResponse, error) {
	return func(ctx context.Context, input *schema.UpdateProfileInput) (*presenter.SuccessResponse, error) {
		current_user := ctx.Value(middleware.CtxUserKey).(*entities.User)
		log.Info().Any("user", current_user).Any("input", input).Msg("Update profile")

		return &presenter.SuccessResponse{Body: presenter.SuccessBody{
			Success: true,
			Message: "Profile updated",
			Data:    toUserResponse(current_user),
		}}, nil
	}
}

func ChangePassword(svc user.Service) func(context.Context, *schema.ChangePasswordInput) (*presenter.SuccessResponse, error) {
	return func(ctx context.Context, input *schema.ChangePasswordInput) (*presenter.SuccessResponse, error) {
		current_user := ctx.Value(middleware.CtxUserKey).(*entities.User)

		err := svc.ChangePassword(ctx, current_user.ID, input.Body.CurrentPassword, input.Body.NewPassword)
		if err != nil {
			if errors.Is(err, user.ErrInvalidCredentials) {
				return nil, huma.Error400BadRequest("current password is incorrect")
			}
			return nil, huma.Error500InternalServerError("internal server error")
		}

		return &presenter.SuccessResponse{Body: presenter.SuccessBody{
			Success: true,
			Message: "Password changed",
		}}, nil
	}
}

func UploadAvatar(svc user.Service) func(context.Context, *schema.UploadAvatarInput) (*presenter.SuccessResponse, error) {
	return func(ctx context.Context, input *schema.UploadAvatarInput) (*presenter.SuccessResponse, error) {
		currentUser := ctx.Value(middleware.CtxUserKey).(*entities.User)
		fileData := input.RawBody.Data()

		bytes, err := io.ReadAll(fileData.Avatar)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to read file")
		}

		user, err := svc.UploadAvatar(ctx, currentUser.ID, bytes, fileData.Avatar.Filename)
		if err != nil {
			return nil, huma.Error500InternalServerError("internal server error")
		}

		return &presenter.SuccessResponse{Body: presenter.SuccessBody{
			Success: true,
			Message: "Avatar uploaded",
			Data:    toUserResponse(user),
		}}, nil
	}
}

func RequestEmailUpdate(svc user.Service) func(context.Context, *schema.EmailUpdateRequestInput) (*presenter.SuccessResponse, error) {
	return func(ctx context.Context, input *schema.EmailUpdateRequestInput) (*presenter.SuccessResponse, error) {
		currentUser := ctx.Value(middleware.CtxUserKey).(*entities.User)

		err := svc.RequestEmailUpdate(ctx, currentUser.ID, input.Body.Email, input.Body.Password)
		if err != nil {
			if errors.Is(err, user.ErrInvalidCredentials) {
				return nil, huma.Error400BadRequest("current password is incorrect")
			}
			if errors.Is(err, user.ErrDuplicateEmail) {
				return nil, huma.Error409Conflict("email already in use")
			}
			return nil, huma.Error500InternalServerError("internal server error")
		}

		return &presenter.SuccessResponse{Body: presenter.SuccessBody{
			Success: true,
			Message: "Confirmation link sent to new email",
		}}, nil
	}
}

func ConfirmEmailUpdate(svc user.Service) func(context.Context, *schema.EmailUpdateConfirmInput) (*presenter.SuccessResponse, error) {
	return func(ctx context.Context, input *schema.EmailUpdateConfirmInput) (*presenter.SuccessResponse, error) {
		err := svc.ConfirmEmailUpdate(ctx, input.Token)
		if err != nil {
			if errors.Is(err, user.ErrInvalidToken) {
				return nil, huma.Error400BadRequest("invalid or expired token")
			}
			return nil, huma.Error500InternalServerError("internal server error")
		}

		return &presenter.SuccessResponse{Body: presenter.SuccessBody{
			Success: true,
			Message: "Email updated successfully",
		}}, nil
	}
}
