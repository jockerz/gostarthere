package tasks

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

type SendEmailPayload struct {
	From      string `json:"from"`
	Recipient string `json:"recipient"`
	Subject   string `json:"subject"`
	Message   string `json:"message"`
}

func (payload *SendEmailPayload) toJsonLog() []byte {
	data := payload.toLog()
	jsonBytes, _ := json.Marshal(data)
	return jsonBytes

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
		log.Error().AnErr("SendEmailTasks", err)
		return nil, err
	} else {
		log.Info().Str("task", "SendEmailTasks").RawJSON("payload", payload.toJsonLog())
	}
	return asynq.NewTask(TypeAuthEmail, jsonPayload), nil
}
