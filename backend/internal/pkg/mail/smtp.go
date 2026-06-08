package mail

import (
	"fmt"
	"net/smtp"
	"strings"

	"github.com/your-org/your-service/internal/config"
)

type Sender interface {
	Send(to, subject, body string) error
}

type NoopSender struct{}

func (NoopSender) Send(to, subject, body string) error { return nil }

type SMTPSender struct {
	cfg config.SMTP
}

func NewSender(cfg config.SMTP) Sender {
	if strings.TrimSpace(cfg.Host) == "" || strings.TrimSpace(cfg.Username) == "" {
		return NoopSender{}
	}
	return &SMTPSender{cfg: cfg}
}

// IsNoop reports whether s is the no-op sender (SMTP host/username not configured).
func IsNoop(s Sender) bool {
	_, ok := s.(NoopSender)
	return ok
}

func (s *SMTPSender) Send(to, subject, body string) error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	auth := smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
	from := s.cfg.FromEmail
	if from == "" {
		from = s.cfg.Username
	}
	fromName := s.cfg.FromName
	if fromName == "" {
		fromName = "App"
	}

	message := []byte(
		fmt.Sprintf("From: %s <%s>\r\n", fromName, from) +
			fmt.Sprintf("To: %s\r\n", to) +
			fmt.Sprintf("Subject: %s\r\n", subject) +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/html; charset=UTF-8\r\n" +
			"\r\n" +
			body,
	)

	return smtp.SendMail(addr, auth, from, []string{to}, message)
}
