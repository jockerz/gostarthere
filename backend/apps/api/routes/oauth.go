package routes

import (
	"github.com/danielgtaylor/huma/v2"

	"vnti/apps/api/handler"
	"vnti/apps/api/schema"
	"vnti/internal"
	"vnti/pkg/service"
)

func OAuthRouter(config *internal.Config, v1 *huma.Group, authSvc service.AuthService, userSvc service.UserService, oauthSvc service.OAuthService) {
	huma.Register(v1, schema.OAuthAuthorizeOp, handler.OAuthAuthorize(config, oauthSvc, authSvc))
	huma.Register(v1, schema.OAuthCallbackOp, handler.OAuthCallback(config, oauthSvc, authSvc))
}
