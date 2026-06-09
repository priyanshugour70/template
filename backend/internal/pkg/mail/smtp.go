// Package mail is the single transactional-email entry point for the app.
// Every module that needs to send mail (auth invites, password resets,
// welcome emails, billing receipts) goes through Sender — no module talks to
// net/smtp directly. The implementation auto-falls-back to NoopSender when
// SMTP isn't configured, so dev environments don't crash.
package mail

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"mime/quotedprintable"
	"net/smtp"
	"strings"

	"github.com/your-org/your-service/internal/config"
)

// Attachment is a single file attached to an outgoing email.
type Attachment struct {
	Filename string // shown to the recipient (e.g. "invoice-INV-2026-000004.pdf")
	MIMEType string // e.g. "application/pdf"
	Data     []byte
}

// Sender is the abstraction every caller depends on. The interface is small
// on purpose — adding attachments was the only meaningful extension since the
// scaffold and it stays additive.
type Sender interface {
	Send(to, subject, body string) error
	SendWithAttachments(to, subject, body string, attachments ...Attachment) error
}

// NoopSender silently drops mail. Used in dev/test when SMTP isn't configured.
type NoopSender struct{}

func (NoopSender) Send(_, _, _ string) error { return nil }
func (NoopSender) SendWithAttachments(_, _, _ string, _ ...Attachment) error {
	return nil
}

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

// Send dispatches an HTML-only email. Body is treated as UTF-8 HTML and
// served with quoted-printable to keep it 7-bit-clean for older SMTP relays.
func (s *SMTPSender) Send(to, subject, body string) error {
	return s.SendWithAttachments(to, subject, body)
}

// SendWithAttachments builds a multipart/mixed message with one HTML body
// part and one base64-encoded part per attachment. Designed for transactional
// mail where attachment counts are small (1–3). For bulk sends, swap the impl.
func (s *SMTPSender) SendWithAttachments(to, subject, body string, atts ...Attachment) error {
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

	var msg bytes.Buffer
	fmt.Fprintf(&msg, "From: %s <%s>\r\n", fromName, from)
	fmt.Fprintf(&msg, "To: %s\r\n", to)
	fmt.Fprintf(&msg, "Subject: %s\r\n", subject)
	msg.WriteString("MIME-Version: 1.0\r\n")

	if len(atts) == 0 {
		// Single-part HTML: quoted-printable body inline.
		msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		msg.WriteString("Content-Transfer-Encoding: quoted-printable\r\n")
		msg.WriteString("\r\n")
		if err := writeQuotedPrintable(&msg, body); err != nil {
			return err
		}
	} else {
		// multipart/mixed: HTML body + each attachment.
		boundary := randomBoundary()
		fmt.Fprintf(&msg, "Content-Type: multipart/mixed; boundary=%q\r\n\r\n", boundary)

		// Body part.
		fmt.Fprintf(&msg, "--%s\r\n", boundary)
		msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		msg.WriteString("Content-Transfer-Encoding: quoted-printable\r\n\r\n")
		if err := writeQuotedPrintable(&msg, body); err != nil {
			return err
		}
		msg.WriteString("\r\n")

		// Attachment parts.
		for _, a := range atts {
			mt := a.MIMEType
			if mt == "" {
				mt = "application/octet-stream"
			}
			fmt.Fprintf(&msg, "--%s\r\n", boundary)
			fmt.Fprintf(&msg, "Content-Type: %s\r\n", mt)
			fmt.Fprintf(&msg, "Content-Disposition: attachment; filename=%q\r\n", a.Filename)
			msg.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")
			msg.WriteString(base64Wrap(a.Data))
			msg.WriteString("\r\n")
		}
		fmt.Fprintf(&msg, "--%s--\r\n", boundary)
	}

	return smtp.SendMail(addr, auth, from, []string{to}, msg.Bytes())
}

func writeQuotedPrintable(w *bytes.Buffer, body string) error {
	qp := quotedprintable.NewWriter(w)
	if _, err := qp.Write([]byte(body)); err != nil {
		return err
	}
	return qp.Close()
}

// base64Wrap encodes to base64 and inserts CRLF every 76 chars (SMTP-safe).
func base64Wrap(data []byte) string {
	enc := base64.StdEncoding.EncodeToString(data)
	var out strings.Builder
	const lineLen = 76
	for i := 0; i < len(enc); i += lineLen {
		end := i + lineLen
		if end > len(enc) {
			end = len(enc)
		}
		out.WriteString(enc[i:end])
		out.WriteString("\r\n")
	}
	return out.String()
}

// randomBoundary generates a unique multipart boundary.
func randomBoundary() string {
	var buf [12]byte
	_, _ = rand.Read(buf[:])
	return "----=_Part_" + hex.EncodeToString(buf[:])
}
