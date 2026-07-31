package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/smtp"
	"vnti/extensions/email"
	"vnti/extensions/logger"
	"vnti/pkg/tasks"

	"github.com/hibiken/asynq"
)

const messageTemplate string = "From: %s\r\n" +
	"To: %s\r\n" +
	"Subject: %s\r\n\r\n" +
	"%s\r\n"

func writeMessage(from, to, subject, message string) string {
	return fmt.Sprintf(messageTemplate, from, to, subject, message)
}

func HanderSendAuthEmail(smtpClient *smtp.Client) func(ctx context.Context, t *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		var payload tasks.SendEmailPayload
		if err := json.Unmarshal(t.Payload(), &payload); err != nil {
			return err
		}
		logger.Logger.Info().Msg(fmt.Sprintf("Processing the email %s\n", t.Payload()))

		message := writeMessage(payload.From, payload.Recipient, payload.Subject, payload.Message)
		if err := email.SendEmail(smtpClient, payload.From, payload.Recipient, message); err != nil {
			logger.Logger.Info().Msg(fmt.Sprintf("Error handler=HanderSendAuthEmail %v", err))
		}
		return nil
	}
}
