package health

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	ms "github.com/your-org/your-service/internal/meilisearch"
	apperrors "github.com/your-org/your-service/internal/pkg/errors"
	"github.com/your-org/your-service/internal/pkg/response"
)

type Status string

const (
	StatusUp       Status = "up"
	StatusDown     Status = "down"
	StatusDegraded Status = "degraded"
)

type ComponentHealth struct {
	Status  Status                 `json:"status"`
	Latency string                 `json:"latency,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

type Response struct {
	Status     Status                     `json:"status"`
	Timestamp  string                     `json:"timestamp"`
	Uptime     string                     `json:"uptime"`
	Version    string                     `json:"version"`
	Env        string                     `json:"env"`
	Components map[string]ComponentHealth `json:"components"`
	System     SystemInfo                 `json:"system"`
}

type SystemInfo struct {
	Hostname    string `json:"hostname"`
	GoVersion   string `json:"goVersion"`
	OS          string `json:"os"`
	Arch        string `json:"arch"`
	NumCPU      int    `json:"numCpu"`
	Goroutines  int    `json:"goroutines"`
	MemoryAlloc string `json:"memoryAlloc"`
	MemoryTotal string `json:"memoryTotal"`
	MemorySys   string `json:"memorySys"`
	GCCycles    uint32 `json:"gcCycles"`
}

type Checker struct {
	startTime   time.Time
	version     string
	env         string
	db          *gorm.DB
	redis       *redis.Client
	meiliClient *ms.Client
}

func NewChecker(version, env string, db *gorm.DB, rdb *redis.Client, meiliClient *ms.Client) *Checker {
	return &Checker{
		startTime:   time.Now(),
		version:     version,
		env:         env,
		db:          db,
		redis:       rdb,
		meiliClient: meiliClient,
	}
}

func (h *Checker) Register(router *gin.Engine) {
	v1 := router.Group("/api/v1")
	v1.GET("/health", h.Health)
	v1.GET("/health/live", h.Live)
	v1.GET("/health/ready", h.Ready)

	router.GET("/health", h.Health)
	router.GET("/health/live", h.Live)
	router.GET("/health/ready", h.Ready)
}

// Health returns the full health report.
//
// @Summary  Comprehensive service health
// @Tags     health
// @Produce  json
// @Success  200 {object} Response
// @Failure  503 {object} Response
// @Router   /health [get]
func (h *Checker) Health(c *gin.Context) {
	components := make(map[string]ComponentHealth)
	components["mariadb"] = h.checkMariaDB(c.Request.Context())
	components["redis"] = h.checkRedis(c.Request.Context())
	components["meilisearch"] = h.checkMeilisearch(c.Request.Context())

	overall := StatusUp
	for _, comp := range components {
		if comp.Status == StatusDown {
			overall = StatusDown
			break
		}
		if comp.Status == StatusDegraded {
			overall = StatusDegraded
		}
	}

	status := http.StatusOK
	if overall == StatusDown {
		status = http.StatusServiceUnavailable
	}

	payload := Response{
		Status:     overall,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		Uptime:     time.Since(h.startTime).Round(time.Second).String(),
		Version:    h.version,
		Env:        h.env,
		Components: components,
		System:     h.systemInfo(),
	}

	if status == http.StatusOK {
		response.OK(c, payload)
		return
	}
	response.Fail(c, status, apperrors.CodeServiceUnavailable, "service health is down",
		map[string]interface{}{"health": payload})
}

// Live always returns 200 if the process is up.
//
// @Summary  Liveness probe
// @Tags     health
// @Produce  json
// @Success  200 {object} map[string]string
// @Router   /health/live [get]
func (h *Checker) Live(c *gin.Context) {
	response.OK(c, gin.H{
		"status":    "up",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// Ready returns 200 only when MariaDB and Redis are reachable.
//
// @Summary  Readiness probe
// @Tags     health
// @Produce  json
// @Success  200 {object} map[string]string
// @Failure  503 {object} map[string]string
// @Router   /health/ready [get]
func (h *Checker) Ready(c *gin.Context) {
	ctx := c.Request.Context()
	dbHealth := h.checkMariaDB(ctx)
	redisHealth := h.checkRedis(ctx)

	if dbHealth.Status == StatusDown {
		response.Fail(c, http.StatusServiceUnavailable, apperrors.CodeServiceUnavailable, "service not ready",
			map[string]interface{}{"status": "not_ready", "reason": "mariadb down"})
		return
	}
	if redisHealth.Status == StatusDown {
		response.Fail(c, http.StatusServiceUnavailable, apperrors.CodeServiceUnavailable, "service not ready",
			map[string]interface{}{"status": "not_ready", "reason": "redis down"})
		return
	}
	response.OK(c, gin.H{"status": "ready", "timestamp": time.Now().UTC().Format(time.RFC3339)})
}

func (h *Checker) checkMariaDB(ctx context.Context) ComponentHealth {
	if h.db == nil {
		return ComponentHealth{Status: StatusDown, Error: "not configured"}
	}
	sqlDB, err := h.db.DB()
	if err != nil {
		return ComponentHealth{Status: StatusDown, Error: fmt.Sprintf("get sql.DB: %v", err)}
	}

	start := time.Now()
	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(pingCtx); err != nil {
		return ComponentHealth{Status: StatusDown, Error: fmt.Sprintf("ping: %v", err)}
	}
	latency := time.Since(start)

	stats := sqlDB.Stats()
	var version string
	row := sqlDB.QueryRowContext(pingCtx, "SELECT VERSION()")
	_ = row.Scan(&version)

	details := map[string]interface{}{
		"version":      version,
		"openConns":    stats.OpenConnections,
		"inUse":        stats.InUse,
		"idle":         stats.Idle,
		"maxOpen":      stats.MaxOpenConnections,
		"waitCount":    stats.WaitCount,
		"waitDuration": stats.WaitDuration.String(),
	}

	status := StatusUp
	if latency > 500*time.Millisecond {
		status = StatusDegraded
	}
	return ComponentHealth{Status: status, Latency: latency.Round(time.Microsecond).String(), Details: details}
}

func (h *Checker) checkRedis(ctx context.Context) ComponentHealth {
	if h.redis == nil {
		return ComponentHealth{Status: StatusDown, Error: "not configured"}
	}
	start := time.Now()
	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	pong, err := h.redis.Ping(pingCtx).Result()
	if err != nil {
		return ComponentHealth{Status: StatusDown, Error: fmt.Sprintf("ping: %v", err)}
	}
	latency := time.Since(start)

	details := map[string]interface{}{"pong": pong}
	poolStats := h.redis.PoolStats()
	details["pool"] = map[string]interface{}{
		"totalConns": poolStats.TotalConns,
		"idleConns":  poolStats.IdleConns,
		"staleConns": poolStats.StaleConns,
		"hits":       poolStats.Hits,
		"misses":     poolStats.Misses,
	}

	status := StatusUp
	if latency > 100*time.Millisecond {
		status = StatusDegraded
	}
	return ComponentHealth{Status: status, Latency: latency.Round(time.Microsecond).String(), Details: details}
}

func (h *Checker) checkMeilisearch(ctx context.Context) ComponentHealth {
	if h.meiliClient == nil {
		return ComponentHealth{Status: StatusUp, Details: map[string]interface{}{"enabled": false}}
	}
	start := time.Now()
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := h.meiliClient.Ping(pingCtx); err != nil {
		return ComponentHealth{Status: StatusDegraded, Error: fmt.Sprintf("ping: %v", err)}
	}
	latency := time.Since(start)

	st := StatusUp
	if latency > 500*time.Millisecond {
		st = StatusDegraded
	}
	return ComponentHealth{Status: st, Latency: latency.Round(time.Microsecond).String()}
}

func (h *Checker) systemInfo() SystemInfo {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	hostname, _ := os.Hostname()

	return SystemInfo{
		Hostname:    hostname,
		GoVersion:   runtime.Version(),
		OS:          runtime.GOOS,
		Arch:        runtime.GOARCH,
		NumCPU:      runtime.NumCPU(),
		Goroutines:  runtime.NumGoroutine(),
		MemoryAlloc: formatBytes(m.Alloc),
		MemoryTotal: formatBytes(m.TotalAlloc),
		MemorySys:   formatBytes(m.Sys),
		GCCycles:    m.NumGC,
	}
}

func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
