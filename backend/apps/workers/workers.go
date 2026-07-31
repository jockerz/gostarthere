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
	logger.InitLogger("backend", "worker", config.Debug)
	database.Connect(config.DB_URL)

	smtpClient, err := email.New(config)
	if err != nil {
		panic(err)
	}

	srv := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     config.RedisAddress(),
			DB:       config.REDIS_DB,
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

// type TaskHandler struct {
// 	config *internal.Config
// }

// func (th *TaskHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
// 	var err error

// 	switch t.Type() {
// 	case auth.TypeAuthEmail:
// 		err = HanderSendAuthEmail(ctx, t)

// 	default:
// 		return fmt.Errorf("Invalid task type %s", t.Type())
// 	}

// 	return err
// }

// func logggingMiddleware(h asynq.Handler) asynq.Handler {
// 	return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
// 		start := time.Now()

// 		log.Printf("Processing task: type=%s payload=%s\n", t.Type(), string(t.Payload()))
// 		if err := h.ProcessTask(ctx, t); err != nil {
// 			log.Println(err)
// 		} else {
// 			log.Printf("Task finished %q: time = %v", t.Type(), time.Since(start))
// 		}
// 		return nil
// 	})
// }
