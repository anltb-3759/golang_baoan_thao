package services

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/awesome-academy/golang_baoan_thao/internal/utils"
	"github.com/wneessen/go-mail"
)

type Mailer interface {
	Send(toEmail, subject, body string) error
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	FromName string
}

func LoadSMTPConfigFromEnv() SMTPConfig {
	host := utils.EnvOr("SMTP_HOST", "smtp.gmail.com")
	port, err := strconv.Atoi(utils.EnvOr("SMTP_PORT", "587"))
	if err != nil {
		log.Fatalf("invalid SMTP_PORT: %v", err)
	}
	user := os.Getenv("SMTP_USERNAME")
	pass := os.Getenv("SMTP_PASSWORD")
	if user == "" || pass == "" {
		log.Fatal("SMTP_USERNAME and SMTP_PASSWORD must be set (use a Gmail App Password)")
	}
	from := utils.EnvOr("SMTP_FROM", user)
	fromName := utils.EnvOr("SMTP_FROM_NAME", "Cổng dịch vụ công")

	return SMTPConfig{
		Host:     host,
		Port:     port,
		Username: user,
		Password: pass,
		From:     from,
		FromName: fromName,
	}
}

type SMTPMailer struct {
	cfg SMTPConfig
}

func NewSMTPMailer(cfg SMTPConfig) *SMTPMailer {
	return &SMTPMailer{cfg: cfg}
}

func (m *SMTPMailer) Send(toEmail, subject, body string) error {
	msg := mail.NewMsg()
	if err := msg.FromFormat(m.cfg.FromName, m.cfg.From); err != nil {
		return fmt.Errorf("set FROM: %w", err)
	}
	if err := msg.To(toEmail); err != nil {
		return fmt.Errorf("set TO: %w", err)
	}
	msg.Subject(subject)
	msg.SetBodyString(mail.TypeTextPlain, body)

	client, err := mail.NewClient(
		m.cfg.Host,
		mail.WithPort(m.cfg.Port),
		mail.WithSMTPAuth(mail.SMTPAuthLogin),
		mail.WithUsername(m.cfg.Username),
		mail.WithPassword(m.cfg.Password),
		mail.WithTLSPolicy(mail.TLSMandatory),
	)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}

	return client.DialAndSend(msg)
}
