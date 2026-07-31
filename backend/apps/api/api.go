package api

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humafiber"
	ZerologMiddleware "github.com/gofiber/contrib/v3/zerolog"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/healthcheck"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/gofiber/fiber/v3/middleware/responsetime"
	"github.com/gofiber/fiber/v3/middleware/static"

	"vnti/apps/api/middleware"
	"vnti/apps/api/routes"
	"vnti/extensions/database"
	"vnti/extensions/logger"
	"vnti/extensions/taskqueue"
	"vnti/internal"
	"vnti/pkg/auth"
	"vnti/pkg/entities"
	"vnti/pkg/oauth2"
	"vnti/pkg/user"

	_ "github.com/danielgtaylor/huma/v2/formats/cbor"
)

func NewApi(config *internal.Config) *fiber.App {
	fiber_app := fiber.New(fiber.Config{
		AppName:      "API",
		ErrorHandler: middleware.ErrorHandler,
	})

	fiber_app.Use(recover.New(
		recover.Config{
			PanicHandler: middleware.PanicHander,
		},
	))
	fiber_app.Use(requestid.New())

	// fiber_app.Use(logger.New(logger.Config{TimeZone: "Asia/jakarta"}))
	logger.InitLogger("backend", "info", config.Debug)
	fiber_app.Use(ZerologMiddleware.New(ZerologMiddleware.Config{
		Logger: &logger.AccessLogger,
	}))
	fiber_app.Use(middleware.NewLoggingMiddleware(config))

	fiber_app.Use(cors.New(cors.Config{
		AllowOrigins: []string{config.ORIGINS},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	}))
	if config.Debug {
		fiber_app.Use(responsetime.New())
	}

	fiber_app.Get("/healthz", healthcheck.New())
	fiber_app.Get("/media/*", static.New(config.MEDIA_PATH))

	initialDatabase(config)
	initiateAPIv1(fiber_app, config)

	return fiber_app
}

func initiateAPIv1(app *fiber.App, config *internal.Config) {
	huma_config := huma.DefaultConfig("API", "1.0.0")
	// Disable `Link: </schemas/SuccessBody.json>; rel="describedBy"` like header
	huma_config.CreateHooks = nil
	huma_config.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"bearer": {
			Type:         "http",
			Scheme:       "Bearer",
			BearerFormat: "JWT",
		},
	}
	huma_api := humafiber.New(app, huma_config)
	group_v1 := huma.NewGroup(huma_api, "/v1")

	taskQueueClient := taskqueue.New(config)

	routes.APICheck(group_v1)

	userRepo := user.NewRepository(database.DB)
	authRepo := auth.NewRepository(database.DB)
	userSvc := user.NewService(config, userRepo, authRepo, taskQueueClient)
	authSvc := auth.NewService(config, userRepo, authRepo, config.SECRET, taskQueueClient)

	oauthStateStore := oauth2.NewStateStore()
	oauthRepo := oauth2.NewRepository(database.DB)
	oauthSvc := oauth2.NewService(config, oauthStateStore, userRepo, authSvc, oauthRepo)

	routes.AuthRouter(config, group_v1, authSvc, userSvc, oauthSvc)
	routes.UserRouter(config, group_v1, authSvc, userSvc)
	routes.OAuthRouter(config, group_v1, authSvc, userSvc, oauthSvc)
}

func initialDatabase(config *internal.Config) {
	database.Connect(config.DB_URL)
	database.DB.AutoMigrate(
		&entities.User{},
		&entities.AuthToken{},
		&entities.UserToken{},
		&entities.UserAuthProvider{},
	)
}
