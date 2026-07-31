package tasks

import (
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

type AsService struct {
	asynqClient *asynq.Client
}

func (s *AsService) EnqueueTask(task *asynq.Task) error {
	taskInfo, err := s.asynqClient.Enqueue(task)

	log.Info().Msg(fmt.Sprintf("task queue info: %+v, err: %v\n", taskInfo, err))
	if err != nil {
		log.Error().Msg(fmt.Sprintf("Fail to enqueue task %s: %s", TypeAuthEmail, err.Error()))
		return ErrTokenActionEnqueueTask
	}
	return nil
}
