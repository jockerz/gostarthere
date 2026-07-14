package routes

import (
	"github.com/danielgtaylor/huma/v2"

	"vnti/apps/api/handler"
	"vnti/apps/api/middleware"
	"vnti/apps/api/schema"
	"vnti/internal"
	"vnti/pkg/auth"
	"vnti/pkg/oauth2"
	"vnti/pkg/user"
)

func AuthRouter(config *internal.Config, v1 *huma.Group, authSvc auth.Service, userSvc user.Service, oauthSvc oauth2.Service) {
	huma.Register(v1, schema.LogoutOp, handler.Logout(authSvc))
	huma.Register(v1, schema.RefreshTokenOp, handler.RefreshToken(authSvc))
	huma.Register(v1, schema.LoginOp, handler.Login(authSvc))
	huma.Register(v1, schema.RegisterOp, handler.Register(authSvc))
	huma.Register(v1, schema.ActivateOp, handler.Activate(authSvc))
	huma.Register(v1, schema.ResendActivationOp, handler.ResendActivation(authSvc))
	huma.Register(v1, schema.ForgotPasswordOp, handler.ForgotPassword(authSvc))
	huma.Register(v1, schema.ResetPasswordOp, handler.ResetPassword(authSvc))

	protected := huma.NewGroup(v1, "")
	protected.UseMiddleware(middleware.Auth(config.SECRET, authSvc, userSvc))
	huma.Register(protected, schema.SetPasswordOp, handler.SetPassword(userSvc))
	huma.Register(protected, schema.GetAuthProviderOp, handler.GetAuthData(oauthSvc))
}
