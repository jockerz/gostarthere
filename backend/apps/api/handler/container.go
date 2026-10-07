package handler

import (
	"vnti/internal"
	"vnti/pkg/service"

	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

type HandlerContainer interface {
	EnqueueTask(*asynq.Task) (*asynq.TaskInfo, error)
}

type handlerContainer struct {
	config      *internal.Config
	DB          *gorm.DB
	Asynq       *asynq.Client
	AuthService *service.AuthService
	UserService *service.UserService
}

func NewHanderContainer(
	config *internal.Config,
	asynqClient *asynq.Client,
	db *gorm.DB,
	authService *service.AuthService,
	userService *service.UserService,
) HandlerContainer {
	return &handlerContainer{
		config: config,
		Asynq:  asynqClient,
		DB:     db,

		AuthService: authService,
		UserService: userService,
	}
}

// func NewTestHanderContainer(
// 	config *internal.Config,
// 	asynqClient *asynq.Client,
// 	db *gorm.DB,
// 	authService *service.AuthService,
// 	userService *service.UserService,
// ) *HandlerContainer {
// 	return &HandlerContainer{
// 		config: config,
// 		Asynq:  asynqClient,
// 		DB:     db,

// 		AuthService: authService,
// 		UserService: userService,
// 	}
// }

func (hc *handlerContainer) EnqueueTask(task *asynq.Task) (*asynq.TaskInfo, error) {
	return hc.Asynq.Enqueue(task)
}
