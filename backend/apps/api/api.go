package api

import (
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humafiber"
	_ "github.com/danielgtaylor/huma/v2/formats/cbor"
	zerologMiddleware "github.com/gofiber/contrib/v3/zerolog"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/healthcheck"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/gofiber/fiber/v3/middleware/responsetime"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/hibiken/asynq"
	"github.com/hibiken/asynqmon"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"

	"vnti/apps/api/middleware"
	"vnti/apps/api/routes"
	"vnti/extensions/database"
	"vnti/extensions/logger"
	"vnti/extensions/taskqueue"
	"vnti/internal"
	"vnti/pkg/entities"
	"vnti/pkg/repository"
	"vnti/pkg/service"
)

func NewApi(config *internal.Config) *fiber.App {
	inititeLog(config)

	fiber_app := fiber.New(fiber.Config{
		AppName:      "API",
		ErrorHandler: middleware.ErrorHandler,
	})

	fiber_app.Use(requestid.New(requestid.Config{
		Generator: middleware.GenRequestId,
	}))
	logField := zerologMiddleware.ConfigDefault.Fields
	logField = append(logField, zerologMiddleware.FieldRequestID)
	fiber_app.Use(zerologMiddleware.New(zerologMiddleware.Config{
		Logger: &logger.AccessLogger,
		Fields: logField,
	}))
	fiber_app.Use(recover.New(recover.Config{
		PanicHandler: middleware.PanicHander,
		// EnableStackTrace: true,
	}))
	// fiber_app.Use(middleware.LoggerToCtxMiddleware(config))
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

type CustomRespWriter struct {
	header        http.Header
	resp          *fasthttp.Response
	headerWritten bool
	statusCode    int
}

func (w CustomRespWriter) Header() http.Header {
	// header := http.Header{}
	// for key, value := range w.resp.Header.All() {
	// 	header.Add(string(key), string(value))
	// }
	// return header
	return w.header
}

func (w CustomRespWriter) WriteHeader(statusCode int) {
	// if w.headerWritten {
	// 	return
	// }

	w.resp.SetStatusCode(statusCode)
	w.headerWritten = true
	for k, v := range w.header {
		for _, vv := range v {
			w.resp.Header.Add(k, vv)
		}
	}
}

func (w CustomRespWriter) Write(b []byte) (int, error) {
	if !w.headerWritten {
		for key, value := range w.resp.Header.All() {
			println("  ", string(key), string(value))
		}
		w.WriteHeader(http.StatusOK)
	}

	w.resp.AppendBody(b)
	return len(b), nil
}

func NewCustomResponseWriter(resp *fasthttp.Response) *CustomRespWriter {
	return &CustomRespWriter{
		resp:       resp,
		header:     make(http.Header),
		statusCode: http.StatusOK,
	}
}

func CopyHTTPHeader(dst *fasthttp.ResponseHeader, src http.Header) {
	for k, v := range src {
		for _, vv := range v {
			dst.Add(k, vv)
		}
	}
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
	huma_config.OpenAPIPath = "/api/openapi"
	huma_api := humafiber.New(app, huma_config)
	group_v1 := huma.NewGroup(huma_api, "/v1")

	taskQueueClient := taskqueue.New(config)

	asynqMonHandler := asynqmon.New(asynqmon.Options{
		RootPath: "/monitoring",
		RedisConnOpt: asynq.RedisClientOpt{
			Addr:     config.RedisAddress(),
			Password: config.REDIS_PASS,
			DB:       config.REDIS_DB_ASYNQ,
		},
	})

	// app.All(asynqMonHandler.RootPath(), adaptor.HTTPHandler(asynqMonHandler))
	app.All(asynqMonHandler.RootPath(), func(c fiber.Ctx) error {
		req, err := adaptor.ConvertRequest(c, true)
		if err != nil {
			return err
		}

		// resp := CustomRespWriter{
		// 	resp:          c.Response(),
		// 	headerWritten: false,
		// }
		resp := NewCustomResponseWriter(c.Response())
		asynqMonHandler.ServeHTTP(*resp, req)
		return nil
	})

	userRepo := repository.NewUserRepository(database.DB)
	authRepo := repository.NewAuthRepository(database.DB)
	userSvc := service.NewUserService(config, userRepo, authRepo, taskQueueClient)
	authSvc := service.NewAuthService(config, userRepo, authRepo, config.SECRET, taskQueueClient)

	oauthStateStore := service.NewOAuthStateStore()
	oauthRepo := repository.NewOAuth2Repository(database.DB)
	oauthSvc := service.NewOAuthService(config, oauthStateStore, userRepo, authSvc, oauthRepo)

	routes.APICheck(group_v1)
	routes.AuthRouter(config, group_v1, authSvc, userSvc, oauthSvc)
	routes.UserRouter(config, group_v1, authSvc, userSvc)
	routes.OAuthRouter(config, group_v1, authSvc, userSvc, oauthSvc)
}

func initialDatabase(config *internal.Config) {
	database.Connect(config.DB_URL)

	if !config.Debug || !strings.Contains(config.DB_URL, "sqlite") {
		// Prod DB migration using atlas
		log.Debug().Msg("Skips database migration for non dev or non SQLite DB")
		return
	}

	database.DB.AutoMigrate(
		&entities.User{},
		&entities.AuthToken{},
		&entities.UserToken{},
		&entities.UserAuthProvider{},
	)
}

func inititeLog(config *internal.Config) {
	// fiber_app.Use(logger.New(logger.Config{TimeZone: "Asia/jakarta"}))
	logger.InitLogger("backend", "INFO", config.Debug)
	logger.InitAccessLogger(config.Debug)
}
