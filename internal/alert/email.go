package alert

import (
	"fmt"
	"net/smtp"
	"strings"
)

// EmailConfig holds SMTP configuration for sending email alerts.
type EmailConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	To       []string
}

// EmailWriter sends alert messages via SMTP email.
type EmailWriter struct {
	cfg  EmailConfig
	auth smtp.Auth
}

// NewEmailWriter creates a new EmailWriter with the given SMTP configuration.
func NewEmailWriter(cfg EmailConfig) *EmailWriter {
	var auth smtp.Auth
	if cfg.Username != "" {
		auth = smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	}
	return &EmailWriter{cfg: cfg, auth: auth}
}

// Write sends the message as an email to all configured recipients.
func (e *EmailWriter) Write(message string) error {
	addr := fmt.Sprintf("%s:%d", e.cfg.Host, e.cfg.Port)
	subject := "VaultWatch Alert"
	body := fmt.Sprintf("To: %s\r\nFrom: %s\r\nSubject: %s\r\n\r\n%s",
		strings.Join(e.cfg.To, ", "),
		e.cfg.From,
		subject,
		message,
	)
	return smtp.SendMail(addr, e.auth, e.cfg.From, e.cfg.To, []byte(body))
}
