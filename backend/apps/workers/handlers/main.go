package handlers

import (
	"vnti/apps/workers"
	"vnti/pkg/tasks"

	"github.com/hibiken/asynq"
)

func AddHandlers(worker *workers.AsynqWorker) *asynq.ServeMux {
	mux := asynq.NewServeMux()

	mux.HandleFunc(tasks.TypeAuthEmail, HanderSendAuthEmail(worker.Smtp))

	return mux
}

func Run(worker *workers.AsynqWorker) error {
	return worker.AsynqServer.Run(AddHandlers(worker))
}
