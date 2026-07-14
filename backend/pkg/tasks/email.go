package tasks

import (
	"context"
	"encoding/json"

	"github.com/gofiber/fiber/v2/log"
	"github.com/hibiken/asynq"
)

type SendEmailPayload struct {
	From      string `json:"from"`
	Recipient string `json:"recipient"`
	Subject   string `json:"subject"`
	Message   string `json:"message"`
}

func (payload *SendEmailPayload) toLog() SendEmailPayload {
	return SendEmailPayload{
		From:      payload.From,
		Recipient: payload.Recipient,
		Subject:   payload.Subject,
		Message:   payload.Message[:25] + "...",
	}
}

func SendEmailTasks(ctx context.Context, from, to, subject, message string) (*asynq.Task, error) {
	payload := SendEmailPayload{
		From:      from,
		Recipient: to,
		Subject:   subject,
		Message:   message,
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Errorf("New SendEmailTasks %v", err)
		return nil, err
	} else {
		log.Infof("task Payload: %+v", payload.toLog())
	}
	return asynq.NewTask(TypeAuthEmail, jsonPayload), nil
}
