package workers

import (
	"net/smtp"
	emailExt "vnti/extensions/email"
	"vnti/internal"

	"github.com/hibiken/asynq"
)

type AsynqWorker struct {
	Config *internal.Config
	Smtp   *smtp.Client

	AsynqServer *asynq.Server
}

func New(config *internal.Config) *AsynqWorker {
	smtpClient, err := emailExt.New(config)
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
