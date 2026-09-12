package taskqueue

import (
	"vnti/internal"

	"github.com/hibiken/asynq"
)

func New(config *internal.Config) *asynq.Client {
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     config.RedisAddress(),
		DB:       config.REDIS_DB_ASYNQ,
		Password: config.REDIS_PASS,
		// PoolSize: 10,
	})
	return client
}
