package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/your-org/your-service/internal/cache"
	"github.com/your-org/your-service/internal/config"
	"github.com/your-org/your-service/internal/modules/audit"
	"github.com/your-org/your-service/internal/modules/rbac"
	"github.com/your-org/your-service/internal/modules/subscription"
	"github.com/your-org/your-service/internal/pkg/logger"
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

	// ── Module services the worker needs (read-only or write-only) ─────────
	auditM := audit.New(db, log)
	rbacM := rbac.New(db, log, cacheSvc, producer)
	subM := subscription.New(db, log, cacheSvc, producer)

	consumer := queue.NewRedisConsumer(queueRdb)

	// audit.events → write to audit_log
	go runConsumer(ctx, consumer, queue.ChannelAudit, auditM.Service.HandlerFunc(), log)

	// permission cache invalidations
	go runConsumer(ctx, consumer, queue.ChannelPermissionInvalidate,
		permissionInvalidateHandler(rbacM.Service, log), log)

	// subscription cache invalidations
	go runConsumer(ctx, consumer, queue.ChannelSubscriptionInvalidate,
		subscriptionInvalidateHandler(subM.Service, log), log)

	// invite + password-reset email channels — wire to your mailer here. For now
	// the worker only logs them so events aren't lost.
	go runConsumer(ctx, consumer, queue.ChannelInviteEmail, logHandler("invite.email", log), log)
	go runConsumer(ctx, consumer, queue.ChannelPasswordResetEmail, logHandler("password_reset.email", log), log)
	go runConsumer(ctx, consumer, queue.ChannelUserWelcomeEmail, logHandler("user.welcome.email", log), log)
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
		// If a (user, org) pair is provided, invalidate that key directly.
		if payload.UserID != "" && payload.OrgID != "" {
			uid, err1 := uuid.Parse(payload.UserID)
			oid, err2 := uuid.Parse(payload.OrgID)
			if err1 == nil && err2 == nil {
				svc.InvalidateUserOrgCache(ctx, uid, oid)
			}
		}
		// Broader invalidations (role change, membership change) are best-effort:
		// the cache entries will expire naturally within TTL. A full sweep
		// implementation would query membership_roles and delete each cache key.
		return nil
	}
}

func subscriptionInvalidateHandler(svc *subscription.Service, log *zap.Logger) queue.Handler {
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

func logHandler(name string, log *zap.Logger) queue.Handler {
	return func(_ context.Context, msg *queue.Message) error {
		log.Info("worker received message", zap.String("channel", name), zap.String("payload", msg.Payload))
		return nil
	}
}
