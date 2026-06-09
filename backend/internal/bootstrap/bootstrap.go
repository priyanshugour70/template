// Package bootstrap is the composition root. Wire every module, client, and
// middleware here. This is the only package that should import concrete module
// and client packages together.
package bootstrap

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/your-org/your-service/internal/cache"
	"github.com/your-org/your-service/internal/config"
	"github.com/your-org/your-service/internal/health"
	"github.com/your-org/your-service/internal/middleware"
	"github.com/your-org/your-service/internal/modules/apikey"
	"github.com/your-org/your-service/internal/modules/audit"
	"github.com/your-org/your-service/internal/modules/auth"
	"github.com/your-org/your-service/internal/modules/department"
	"github.com/your-org/your-service/internal/modules/group"
	"github.com/your-org/your-service/internal/modules/webhook"
	"github.com/your-org/your-service/internal/modules/notification"
	"github.com/your-org/your-service/internal/modules/rbac"
	"github.com/your-org/your-service/internal/modules/subscription"
	"github.com/your-org/your-service/internal/modules/tenant"
	"github.com/your-org/your-service/internal/modules/user"
	"github.com/your-org/your-service/internal/pkg/logger"
	pkgmodel "github.com/your-org/your-service/internal/pkg/model"
	"github.com/your-org/your-service/internal/pkg/response"
	"github.com/your-org/your-service/internal/queue"
	"github.com/your-org/your-service/internal/repository"
)

const Version = "0.1.0"

func parseAllowedOrigins(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if o := strings.TrimSpace(p); o != "" {
			out = append(out, o)
		}
	}
	return out
}

// API is the bag of handles built by BootstrapAPI. Closed in main on shutdown.
type API struct {
	Config       *config.Config
	Log          *zap.Logger
	Router       *gin.Engine
	DB           *gorm.DB
	Redis        *redis.Client
	Cache        cache.Cache
	Producer     queue.Producer
	QueueRedis   *redis.Client
	TenantSvc    *tenant.Service
	UserSvc      *user.Service
	RBACSvc      *rbac.Service
	SubSvc       *subscription.Service
	AuditSvc     *audit.Service
	AuthSvc      *auth.Service
	NotifSvc     *notification.Service
	DeptSvc      *department.Service
	GroupSvc     *group.Service
	APIKeySvc    *apikey.Service
	WebhookSvc   *webhook.Service
}

func BootstrapAPI(ctx context.Context, cfg *config.Config, log *zap.Logger) (*API, error) {
	var err error
	if log == nil {
		log, err = logger.New(cfg.App.Env)
		if err != nil {
			return nil, err
		}
	}

	gin.SetMode(cfg.Server.Mode)
	router := gin.New()
	router.MaxMultipartMemory = 50 << 20

	origins := parseAllowedOrigins(cfg.CORS.AllowedOrigins)
	router.Use(cors.New(cors.Config{
		AllowOrigins: origins,
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{
			"Origin", "Content-Type", "Accept", "Authorization", "X-Api-Key",
			middleware.HeaderCorrelationID, middleware.HeaderRequestID,
		},
		AllowCredentials: true,
		ExposeHeaders:    []string{middleware.HeaderCorrelationID, middleware.HeaderRequestID},
	}))
	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.Recovery(log), middleware.CorrelationID(log), middleware.Logger(log))

	// ── PostgreSQL ─────────────────────────────────────────────────────────
	var db *gorm.DB
	if cfg.Postgres.Host != "" {
		logLevel := "warn"
		if cfg.App.Env == "development" {
			logLevel = "info"
		}
		db, err = repository.OpenPostgres(ctx, repository.DBConfig{
			DSN:            cfg.Postgres.DSN(),
			MaxOpenConns:   cfg.Postgres.MaxOpenConns,
			MaxIdleConns:   cfg.Postgres.MaxIdleConns,
			ConnMaxLifeSec: cfg.Postgres.ConnMaxLife,
			LogLevel:       logLevel,
		}, log)
		if err != nil {
			log.Warn("PostgreSQL unavailable; API starting in degraded mode", zap.Error(err))
			db = nil
		}
	} else {
		log.Warn("POSTGRES_HOST not set; database features disabled")
	}

	// ── Redis ─────────────────────────────────────────────────────────────
	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.Addr,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})
	pingCtx, pingCancel := context.WithTimeout(ctx, 5*time.Second)
	if err := rdb.Ping(pingCtx).Err(); err != nil {
		log.Warn("Redis unavailable, cache and queue will no-op", zap.Error(err))
		_ = rdb.Close()
		rdb = nil
	}
	pingCancel()

	var cacheSvc cache.Cache
	var producer queue.Producer
	var queueRdb *redis.Client
	if rdb != nil {
		cacheSvc = cache.NewRedisCache(rdb)
		queueRdb = newRedisClient(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.QueueDB)
		queuePingCtx, queuePingCancel := context.WithTimeout(ctx, 5*time.Second)
		queuePingErr := queueRdb.Ping(queuePingCtx).Err()
		queuePingCancel()
		if queuePingErr != nil {
			log.Warn("Redis queue DB unavailable, pub/sub will no-op", zap.Error(queuePingErr))
			_ = queueRdb.Close()
			queueRdb = nil
			producer = &queue.NoopProducer{}
		} else {
			producer = queue.NewRedisProducer(queueRdb)
			log.Info("redis pub/sub producer ready", zap.Int("queue_db", cfg.Redis.QueueDB))
		}
	} else {
		cacheSvc = cache.NewMemoryCache()
		producer = &queue.NoopProducer{}
	}

	// ── Health + Swagger ──────────────────────────────────────────────────
	healthChecker := health.NewChecker(Version, cfg.App.Env, db, rdb)
	healthChecker.Register(router)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler,
		ginSwagger.DefaultModelsExpandDepth(-1),
	))
	router.GET("/", func(c *gin.Context) { c.Redirect(http.StatusFound, "/swagger/index.html") })

	api := router.Group("/api/v1")

	out := &API{
		Config:     cfg,
		Log:        log,
		Router:     router,
		DB:         db,
		Redis:      rdb,
		Cache:      cacheSvc,
		Producer:   producer,
		QueueRedis: queueRdb,
	}

	if db != nil {
		// Install GORM callbacks that auto-populate created_by/updated_by/deleted_by
		// from request context for every model.
		if err := pkgmodel.RegisterCallbacks(db); err != nil {
			log.Warn("model callbacks registration failed", zap.Error(err))
		}
		registerModules(api, db, log, cfg, producer, cacheSvc, out)
	} else {
		log.Warn("skipping module registration — no database connection")
		router.NoRoute(func(c *gin.Context) {
			if strings.HasPrefix(c.Request.URL.Path, "/api/v1") {
				response.Fail(c, 503, "SERVICE_UNAVAILABLE",
					"API requires a database connection. Set POSTGRES_* and ensure PostgreSQL is reachable, then restart.",
					nil)
				return
			}
			c.JSON(404, gin.H{"success": false, "error": gin.H{"code": "NOT_FOUND", "message": "not found"}})
		})
	}

	return out, nil
}

func newRedisClient(addr, password string, db int) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})
}

// registerModules wires every domain module in dependency order:
// tenant → user → rbac → subscription → audit → auth.
func registerModules(
	api *gin.RouterGroup,
	db *gorm.DB,
	log *zap.Logger,
	cfg *config.Config,
	producer queue.Producer,
	cacheSvc cache.Cache,
	out *API,
) {
	// Build module composition roots.
	tenantM := tenant.New(db, log, cacheSvc)
	userM := user.New(db, log, cacheSvc)
	rbacM := rbac.New(db, log, cacheSvc, producer)
	subM := subscription.New(db, log, cacheSvc, producer)
	auditM := audit.New(db, log)
	notifM := notification.New(db, log)
	// dept + group plug into rbac.Service for cache invalidation on role-binding changes.
	deptM := department.New(db, rbacM.Service, log)
	groupM := group.New(db, rbacM.Service, log)
	apikeyM := apikey.New(db, log)
	webhookM := webhook.New(db, log)
	authM := auth.New(db, tenantM.Service, userM.Service, rbacM.Service, cfg, cacheSvc, producer, log)

	out.TenantSvc = tenantM.Service
	out.UserSvc = userM.Service
	out.RBACSvc = rbacM.Service
	out.SubSvc = subM.Service
	out.AuditSvc = auditM.Service
	out.AuthSvc = authM.Service
	out.NotifSvc = notifM.Service
	out.DeptSvc = deptM.Service
	out.GroupSvc = groupM.Service
	out.APIKeySvc = apikeyM.Service
	out.WebhookSvc = webhookM.Service

	// Audit middleware on the /api/v1 group — captures every request after
	// auth has populated user/tenant/org context.
	api.Use(middleware.Audit(producer, log, middleware.DefaultAuditConfig()))

	// Build shared middleware factories.
	authMW := middleware.AuthRequired(authM.Signer)
	authOpt := middleware.AuthOptional(authM.Signer)
	permFn := rbacM.Middleware.RequirePermission

	// AuthOptional runs on the api group so audit can capture user info on
	// unauth'd paths if the client sent a token anyway.
	api.Use(authOpt)

	// Mount module routes. Note: auth-related sub-routes that require an
	// authenticated principal pass authMW into their own group inside Routes().
	tenantM.Handler.Routes(api, authMW, permFn)
	userM.Handler.WithRBAC(rbacM.Service).Routes(api, authMW, permFn)
	rbacM.Handler.Routes(api, authMW, permFn)
	subM.Handler.Routes(api, authMW, permFn)
	auditM.Handler.Routes(api, authMW, permFn)
	notifM.Handler.Routes(api, authMW, permFn)
	deptM.Handler.Routes(api, authMW, permFn)
	groupM.Handler.Routes(api, authMW, permFn)
	apikeyM.Handler.Routes(api, authMW, permFn)
	webhookM.Handler.Routes(api, authMW, permFn)
	authM.Handler.Routes(api, authMW, permFn)
}

func (a *API) Close() {
	if a.DB != nil {
		if sqlDB, err := a.DB.DB(); err == nil {
			_ = sqlDB.Close()
		}
	}
	if a.QueueRedis != nil {
		_ = a.QueueRedis.Close()
	}
	if a.Redis != nil {
		_ = a.Redis.Close()
	}
}
