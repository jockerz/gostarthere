package workers

import (
	"net/smtp"
	"vnti/extensions/database"
	"vnti/extensions/email"
	"vnti/extensions/logger"
	"vnti/internal"

	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

type AsynqWorker struct {
	Config *internal.Config

	DB   *gorm.DB
	Smtp *smtp.Client

	AsynqServer *asynq.Server
}

func New(config *internal.Config) *AsynqWorker {
	logger.InitLogger("backend", "WORKER", config.Debug)
	database.Connect(config.DB_URL)

	smtpClient, err := email.New(config)
	if err != nil {
		panic(err)
	}

	srv := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     config.RedisAddress(),
			DB:       config.REDIS_DB_ASYNQ,
			Password: config.REDIS_PASS,
		},
		asynq.Config{Concurrency: 10},
	)

	worker := &AsynqWorker{
		Config:      config,
		DB:          database.DB,
		Smtp:        smtpClient,
		AsynqServer: srv,
	}

	return worker
}
