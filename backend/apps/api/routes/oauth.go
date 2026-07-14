package routes

import (
	"github.com/danielgtaylor/huma/v2"

	"vnti/apps/api/handler"
	"vnti/apps/api/schema"
	"vnti/internal"
	"vnti/pkg/auth"
	"vnti/pkg/oauth2"
	"vnti/pkg/user"
)

func OAuthRouter(config *internal.Config, v1 *huma.Group, authSvc auth.Service, userSvc user.Service, oauthSvc oauth2.Service) {
	huma.Register(v1, schema.OAuthAuthorizeOp, handler.OAuthAuthorize(config, oauthSvc, authSvc))
	huma.Register(v1, schema.OAuthCallbackOp, handler.OAuthCallback(config, oauthSvc, authSvc))
}
