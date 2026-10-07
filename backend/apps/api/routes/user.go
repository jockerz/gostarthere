package routes

import (
	"github.com/danielgtaylor/huma/v2"

	"vnti/apps/api/handler"
	"vnti/apps/api/middleware"
	"vnti/apps/api/schema"
	"vnti/internal"
	"vnti/pkg/service"
)

func UserRouter(config *internal.Config, v1 *huma.Group, authServ service.AuthService, userSvc service.UserService) {
	protected_profile := huma.NewGroup(v1, "")
	protected_profile.UseMiddleware(middleware.AuthAllowNonActive(config.SECRET, authServ, userSvc))
	huma.Register(protected_profile, schema.GetProfile, handler.GetProfile(userSvc))

	protected := huma.NewGroup(v1, "")
	protected.UseMiddleware(middleware.Auth(config.SECRET, authServ, userSvc))
	huma.Register(protected, schema.UpdateProfile, handler.UpdateProfile(userSvc))
	huma.Register(protected, schema.UploadAvatarOp, handler.UploadAvatar(userSvc))
	huma.Register(protected, schema.ChangePasswordOp, handler.ChangePassword(userSvc))
	huma.Register(protected, schema.EmailUpdateRequestOp, handler.RequestEmailUpdate(userSvc))
	huma.Register(protected, schema.EmailUpdateConfirmOp, handler.ConfirmEmailUpdate(userSvc))
}
