package email

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"vnti/internal"
)

var Smtp *smtp.Client

func New(config *internal.Config) (*smtp.Client, error) {
	log.Println("SMTP starts")

	smtpClient, err := smtp.Dial(config.SMTPAddress())
	if err != nil {
		return nil, fmt.Errorf("SMTP: dial error %w", err)
	}
	// defer smtpClient.Close()

	if config.SMTP_STARTTLS {
		if err = smtpClient.StartTLS(&tls.Config{ServerName: config.SMTP_HOST}); err != nil {
			return nil, fmt.Errorf("SMTP: StartTLS error %w", err)
		}
	}

	// Authentication
	if config.SMTP_USERNAME != "" && config.SMTP_PASSWORD != "" {
		auth := smtp.PlainAuth("", config.SMTP_USERNAME, config.SMTP_PASSWORD, config.SMTP_HOST)
		if err = smtpClient.Auth(auth); err != nil {
			return nil, fmt.Errorf("SMTP: auth error %w", err)
		}
	}

	log.Println("SMTP is ready")

	return smtpClient, nil
}

func SendEmail(client *smtp.Client, from, to, message string) error {
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("Mail from error: %w", err)
	}

	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("Mail recipient error: %w", err)
	}

	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("Mail data error: %w", err)
	}
	defer wc.Close()

	_, err = wc.Write([]byte(message))
	return err
}
