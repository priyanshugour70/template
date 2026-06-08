# internal/pkg/mail/

Email sending via SMTP with a graceful no-op fallback.

- `Sender` interface with `Send(to, subject, body)` method.
- `SMTPSender` — real SMTP implementation, HTML content type.
- `NoopSender` — silent no-op when SMTP host/username are not configured.
- `NewSender(cfg)` — factory; chooses based on config.
- `IsNoop(s)` — true when sender is the no-op (handy for warnings at startup).

Templates (`templates.go`) provide a consistent transactional layout used by auth (invitations, password reset) and notification flows. Replace the brand name in `EmailInvitation` / `EmailPasswordReset` for your project.
