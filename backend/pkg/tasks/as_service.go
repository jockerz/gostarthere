package tasks

import (
	"fmt"

	"github.com/gofiber/fiber/v2/log"
	"github.com/hibiken/asynq"
)

type AsService struct {
	asynqClient *asynq.Client
}

func (s *AsService) EnqueueTask(task *asynq.Task) error {
	taskInfo, err := s.asynqClient.Enqueue(task)
	fmt.Printf("task queue info: %+v, err: %v\n", taskInfo, err)
	if err != nil {
		log.Errorf("Fail to enqueue task %s: %s", TypeAuthEmail, err.Error())
		return ErrTokenActionEnqueueTask
	}
	return nil
}
