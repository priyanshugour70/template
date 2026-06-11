package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/your-org/your-service/internal/cache"
	"github.com/your-org/your-service/internal/config"
	"github.com/your-org/your-service/internal/modules/audit"
	"github.com/your-org/your-service/internal/modules/billing"
	"github.com/your-org/your-service/internal/modules/rbac"
	"github.com/your-org/your-service/internal/pkg/logger"
	"github.com/your-org/your-service/internal/pkg/mail"
	pkgmodel "github.com/your-org/your-service/internal/pkg/model"
	"github.com/your-org/your-service/internal/queue"
	"github.com/your-org/your-service/internal/repository"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	log, err := logger.NewWithConfig(cfg.App.Env, logger.FileConfig{
		BaseDir:    "logs",
		Purpose:    "worker",
		MaxSizeMB:  10,
		MaxBackups: 30,
	})
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ── Postgres ───────────────────────────────────────────────────────────
	db, err := repository.OpenPostgres(ctx, repository.DBConfig{
		DSN:            cfg.Postgres.DSN(),
		MaxOpenConns:   cfg.Postgres.MaxOpenConns,
		MaxIdleConns:   cfg.Postgres.MaxIdleConns,
		ConnMaxLifeSec: cfg.Postgres.ConnMaxLife,
		LogLevel:       "warn",
	}, log)
	if err != nil {
		log.Fatal("postgres open failed", zap.Error(err))
	}
	if err := pkgmodel.RegisterCallbacks(db); err != nil {
		log.Warn("model callbacks registration failed", zap.Error(err))
	}

	// ── Redis (cache + queue dbs) ──────────────────────────────────────────
	cacheRdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err := cacheRdb.Ping(ctx).Err(); err != nil {
		log.Fatal("redis cache ping failed", zap.Error(err))
	}
	defer cacheRdb.Close()

	queueRdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.QueueDB,
	})
	if err := queueRdb.Ping(ctx).Err(); err != nil {
		log.Fatal("redis queue ping failed", zap.Error(err))
	}
	defer queueRdb.Close()

	cacheSvc := cache.NewRedisCache(cacheRdb)
	producer := queue.NewRedisProducer(queueRdb)

	// Mail sender — same NewSender as the API. NoopSender when SMTP host /
	// username aren't configured so dev workers still run without spamming.
	mailer := mail.NewSender(cfg.SMTP)
	if mail.IsNoop(mailer) {
		log.Info("worker mail sender: noop (SMTP host/username not set)")
	} else {
		log.Info("worker mail sender: SMTP", zap.String("host", cfg.SMTP.Host))
	}

	// ── Module services the worker needs (read-only or write-only) ─────────
	auditM := audit.New(db, log)
	rbacM := rbac.New(db, log, cacheSvc, producer)
	// Billing service is wired for cache-invalidation + future cycle/trial
	// jobs. S3 + tenantSvc stay nil because the worker doesn't render PDFs.
	subM := billing.New(db, log, cacheSvc, producer, nil, nil, mailer)

	consumer := queue.NewRedisConsumer(queueRdb)

	// audit.events → write to audit_log
	go runConsumer(ctx, consumer, queue.ChannelAudit, auditM.Service.HandlerFunc(), log)

	// permission cache invalidations
	go runConsumer(ctx, consumer, queue.ChannelPermissionInvalidate,
		permissionInvalidateHandler(rbacM.Service, log), log)

	// subscription cache invalidations
	go runConsumer(ctx, consumer, queue.ChannelSubscriptionInvalidate,
		subscriptionInvalidateHandler(subM.Service, log), log)

	// Transactional email consumers — each one decodes the publisher's
	// payload, builds the template, and dispatches via the shared Sender.
	// Falls back to NoopSender locally so we just log on dev.
	go runConsumer(ctx, consumer, queue.ChannelInviteEmail,
		inviteEmailHandler(mailer, cfg, log), log)
	go runConsumer(ctx, consumer, queue.ChannelPasswordResetEmail,
		passwordResetEmailHandler(mailer, log), log)
	go runConsumer(ctx, consumer, queue.ChannelUserWelcomeEmail,
		welcomeEmailHandler(mailer, log), log)
	// Billing cycle tick — daily 02:00 IST or on-demand publishes by admins.
	go runConsumer(ctx, consumer, queue.ChannelBillingCycle,
		billingCycleHandler(subM.Service, log), log)
	// Internal scheduler that publishes the daily tick. Uses a goroutine so
	// it doesn't block shutdown.
	go scheduleBillingCycle(ctx, producer, log)

	go runConsumer(ctx, consumer, queue.ChannelNotifications, logHandler("notifications", log), log)
	go runConsumer(ctx, consumer, queue.ChannelSearchSync, logHandler("search.sync", log), log)

	log.Info("worker started (redis pub/sub)")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("worker shutting down")
	cancel()
	time.Sleep(2 * time.Second)
}

func runConsumer(ctx context.Context, consumer queue.Consumer, channel string, handler queue.Handler, log *zap.Logger) {
	log.Info("subscribing to redis pub/sub channel", zap.String("channel", channel))
	if err := consumer.Consume(ctx, channel, handler); err != nil && ctx.Err() == nil {
		log.Error("pub/sub consumer exited with error", zap.String("channel", channel), zap.Error(err))
	}
}

func permissionInvalidateHandler(svc *rbac.Service, log *zap.Logger) queue.Handler {
	return func(ctx context.Context, msg *queue.Message) error {
		var payload struct {
			MembershipID string `json:"membershipId,omitempty"`
			RoleID       string `json:"roleId,omitempty"`
			UserID       string `json:"userId,omitempty"`
			OrgID        string `json:"organizationId,omitempty"`
		}
		if err := json.Unmarshal([]byte(msg.Payload), &payload); err != nil {
			log.Warn("permission invalidate decode failed", zap.Error(err))
			return nil
		}
		if payload.UserID != "" && payload.OrgID != "" {
			uid, err1 := uuid.Parse(payload.UserID)
			oid, err2 := uuid.Parse(payload.OrgID)
			if err1 == nil && err2 == nil {
				svc.InvalidateUserOrgCache(ctx, uid, oid)
			}
		}
		return nil
	}
}

func subscriptionInvalidateHandler(svc *billing.Service, log *zap.Logger) queue.Handler {
	return func(ctx context.Context, msg *queue.Message) error {
		var payload struct {
			OrgID string `json:"organizationId"`
		}
		if err := json.Unmarshal([]byte(msg.Payload), &payload); err != nil {
			log.Warn("subscription invalidate decode failed", zap.Error(err))
			return nil
		}
		if payload.OrgID != "" {
			if id, err := uuid.Parse(payload.OrgID); err == nil {
				svc.InvalidateCacheForOrg(ctx, id)
			}
		}
		return nil
	}
}

// inviteEmailHandler decodes auth.publishInviteEmail payloads and sends the
// invitation template. The accept URL is composed per-tenant from the slug +
// apex sent in the payload — falls back to the legacy password-reset-base
// origin if the publisher didn't include them (older versions of the auth
// service).
func inviteEmailHandler(mailer mail.Sender, cfg *config.Config, log *zap.Logger) queue.Handler {
	return func(_ context.Context, msg *queue.Message) error {
		var p struct {
			InviteID       string `json:"inviteId"`
			Email          string `json:"email"`
			Token          string `json:"token"`
			OrganizationID string `json:"organizationId"`
			TenantID       string `json:"tenantId"`
			// Multi-tenant subdomain hints — populated by the modern publisher.
			TenantSlug     string `json:"tenantSlug"`
			Apex           string `json:"apex"`
			FrontendScheme string `json:"frontendScheme"`
			FrontendPort   string `json:"frontendPort"`
		}
		if err := json.Unmarshal([]byte(msg.Payload), &p); err != nil {
			log.Warn("invite email decode failed", zap.Error(err))
			return nil
		}
		if p.Email == "" || p.Token == "" {
			log.Warn("invite email missing email or token", zap.String("inviteId", p.InviteID))
			return nil
		}
		params := map[string]string{"token": p.Token, "email": p.Email}
		acceptURL := buildTenantURL(p.TenantSlug, p.Apex, p.FrontendScheme, p.FrontendPort,
			"/auth/accept-invite", params)
		if acceptURL == "" {
			// Legacy fallback — older publishers (or misconfigured envs) don't
			// include the tenant subdomain. Use the password-reset base URL's
			// origin, which at worst lands on the apex.
			acceptURL = buildFrontendURL(cfg.Auth.PasswordResetBaseURL, "/auth/accept-invite", params)
		}
		// 7-day expiry is the auth module's default; render UTC for clarity.
		expiresAt := time.Now().Add(7 * 24 * time.Hour).UTC().Format("02 Jan 2006 15:04 MST")
		body := mail.EmailInvitation(acceptURL, expiresAt, "")
		if err := mailer.Send(p.Email, "You're invited", body); err != nil {
			log.Warn("invite email send failed", zap.Error(err), zap.String("to", p.Email))
			return err
		}
		log.Info("invite email dispatched", zap.String("to", p.Email), zap.String("acceptUrl", acceptURL))
		return nil
	}
}

// buildTenantURL composes https://{slug}.{apex}[:port]{path}?{params} from
// the tenant-hosting hints carried in pub/sub payloads. Returns "" when the
// hints are missing so callers can fall back to a legacy path. Dev-friendly:
// when scheme is empty defaults to https, port may be "" in prod.
func buildTenantURL(slug, apex, scheme, port, path string, params map[string]string) string {
	slug = strings.TrimSpace(slug)
	apex = strings.TrimSpace(apex)
	if slug == "" || apex == "" {
		return ""
	}
	if scheme == "" {
		scheme = "https"
	}
	host := slug + "." + apex
	if port != "" && port != "80" && port != "443" {
		host = host + ":" + port
	}
	u := url.URL{Scheme: scheme, Host: host, Path: path}
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

// passwordResetEmailHandler decodes the publisher's payload, builds the reset
// URL ({baseUrl}?token=...) and dispatches the template.
func passwordResetEmailHandler(mailer mail.Sender, log *zap.Logger) queue.Handler {
	return func(_ context.Context, msg *queue.Message) error {
		var p struct {
			UserID  string `json:"userId"`
			Email   string `json:"email"`
			Token   string `json:"token"`
			BaseURL string `json:"baseUrl"`
		}
		if err := json.Unmarshal([]byte(msg.Payload), &p); err != nil {
			log.Warn("password reset email decode failed", zap.Error(err))
			return nil
		}
		if p.Email == "" || p.Token == "" {
			log.Warn("password reset email missing email or token")
			return nil
		}
		base := strings.TrimSpace(p.BaseURL)
		if base == "" {
			base = "http://localhost:3000/reset-password"
		}
		resetURL := appendQuery(base, "token", p.Token)
		body := mail.EmailPasswordReset(resetURL)
		if err := mailer.Send(p.Email, "Reset your password", body); err != nil {
			log.Warn("password reset email send failed", zap.Error(err), zap.String("to", p.Email))
			return err
		}
		log.Info("password reset email dispatched", zap.String("to", p.Email))
		return nil
	}
}

// welcomeEmailHandler dispatches the welcome template on first sign-in.
func welcomeEmailHandler(mailer mail.Sender, log *zap.Logger) queue.Handler {
	return func(_ context.Context, msg *queue.Message) error {
		var p struct {
			UserID      string `json:"userId,omitempty"`
			Email       string `json:"email"`
			DisplayName string `json:"displayName,omitempty"`
		}
		if err := json.Unmarshal([]byte(msg.Payload), &p); err != nil {
			log.Warn("welcome email decode failed", zap.Error(err))
			return nil
		}
		if p.Email == "" {
			log.Warn("welcome email missing recipient")
			return nil
		}
		body := mail.EmailWelcome(p.DisplayName)
		if err := mailer.Send(p.Email, "Welcome aboard", body); err != nil {
			log.Warn("welcome email send failed", zap.Error(err), zap.String("to", p.Email))
			return err
		}
		log.Info("welcome email dispatched", zap.String("to", p.Email))
		return nil
	}
}

// billingCycleHandler is the worker-side glue. Decodes the (optional) `at`
// timestamp from the message (so on-demand admins can backdate), defaults to
// now, and runs the cycle.
func billingCycleHandler(svc *billing.Service, log *zap.Logger) queue.Handler {
	return func(ctx context.Context, msg *queue.Message) error {
		var p struct {
			At string `json:"at,omitempty"`
		}
		_ = json.Unmarshal([]byte(msg.Payload), &p)
		at := time.Now()
		if p.At != "" {
			if parsed, err := time.Parse(time.RFC3339, p.At); err == nil {
				at = parsed
			}
		}
		rep, err := svc.RunBillingCycle(ctx, at)
		if err != nil {
			log.Warn("billing cycle failed", zap.Error(err))
			return err
		}
		log.Info("billing cycle report",
			zap.Int("trials_expired", rep.TrialsExpired),
			zap.Int("invoices_issued", rep.InvoicesIssued),
			zap.Int("errors", len(rep.Errors)))
		return nil
	}
}

// scheduleBillingCycle publishes ChannelBillingCycle once a day at 02:00 IST.
// Re-computes "next 02:00" on every iteration so DST + manual clock changes
// don't drift the schedule. Skipped if the producer is the noop one.
func scheduleBillingCycle(ctx context.Context, producer queue.Producer, log *zap.Logger) {
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		log.Warn("billing cycle: load IST failed, using UTC", zap.Error(err))
		loc = time.UTC
	}
	for {
		now := time.Now().In(loc)
		next := time.Date(now.Year(), now.Month(), now.Day(), 2, 0, 0, 0, loc)
		if !next.After(now) {
			next = next.Add(24 * time.Hour)
		}
		wait := time.Until(next)
		log.Info("billing cycle: next tick scheduled",
			zap.Time("at", next), zap.Duration("in", wait))
		select {
		case <-ctx.Done():
			return
		case <-time.After(wait):
			if err := producer.Publish(ctx, queue.ChannelBillingCycle, map[string]interface{}{}); err != nil {
				log.Warn("billing cycle: publish tick failed", zap.Error(err))
			}
		}
	}
}

func logHandler(name string, log *zap.Logger) queue.Handler {
	return func(_ context.Context, msg *queue.Message) error {
		log.Info("worker received message", zap.String("channel", name), zap.String("payload", msg.Payload))
		return nil
	}
}

// buildFrontendURL combines a reference URL's origin with a fresh path + query.
// Used to derive the invite accept URL from the password-reset base URL — they
// share the same frontend origin.
func buildFrontendURL(refURL, path string, params map[string]string) string {
	u, err := url.Parse(refURL)
	if err != nil || u.Scheme == "" {
		return fmt.Sprintf("http://localhost:3000%s?%s", path, encodeParams(params))
	}
	u.Path = path
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	u.Fragment = ""
	return u.String()
}

func appendQuery(rawURL, key, value string) string {
	u, err := url.Parse(rawURL)
	if err != nil || u.Scheme == "" {
		sep := "?"
		if strings.Contains(rawURL, "?") {
			sep = "&"
		}
		return rawURL + sep + url.QueryEscape(key) + "=" + url.QueryEscape(value)
	}
	q := u.Query()
	q.Set(key, value)
	u.RawQuery = q.Encode()
	return u.String()
}

func encodeParams(p map[string]string) string {
	v := url.Values{}
	for k, val := range p {
		v.Set(k, val)
	}
	return v.Encode()
}
