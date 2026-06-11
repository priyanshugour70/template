package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	App       App
	Server    Server
	Postgres  Postgres
	Redis     Redis
	SMTP      SMTP
	CORS      CORS
	Auth      Auth
	Assets    Assets
	Developer Developer
}

type App struct {
	Env string `mapstructure:"env"`
	// BaseDomain is the apex that frontends are hosted under, e.g. "lssgoo.com".
	// Used to construct tenant-aware URLs (password reset link, handoff redirect,
	// CORS wildcard) and to recognise subdomains on incoming requests.
	BaseDomain string `mapstructure:"base_domain"`
	// FrontendScheme is "http" (dev) or "https" (prod) — used when the backend
	// composes a frontend URL it cannot infer from the request.
	FrontendScheme string `mapstructure:"frontend_scheme"`
	// FrontendPort is appended to the host when constructing frontend URLs in
	// dev. Leave empty in prod (the scheme implies 80/443). Example: "3000".
	FrontendPort string `mapstructure:"frontend_port"`
}

type Server struct {
	Port           int    `mapstructure:"port"`
	ReadTimeout    int    `mapstructure:"read_timeout_sec"`
	WriteTimeout   int    `mapstructure:"write_timeout_sec"`
	IdleTimeoutSec int    `mapstructure:"idle_timeout_sec"`
	MaxHeaderBytes int    `mapstructure:"max_header_bytes"`
	Mode           string `mapstructure:"mode"`
}

type Postgres struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	Database     string `mapstructure:"database"`
	SSLMode      string `mapstructure:"sslmode"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	ConnMaxLife  int    `mapstructure:"conn_max_life_sec"`
}

// DSN builds a key=value PostgreSQL connection string consumable by pgx and lib/pq.
func (p Postgres) DSN() string {
	if p.Port <= 0 {
		p.Port = 5432
	}
	if p.Database == "" {
		p.Database = "app_db"
	}
	if p.SSLMode == "" {
		p.SSLMode = "disable"
	}
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=10 TimeZone=Asia/Kolkata",
		p.Host, p.Port, p.User, p.Password, p.Database, p.SSLMode,
	)
}

type Redis struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	QueueDB  int    `mapstructure:"queue_db"`
}

type CORS struct {
	// AllowedOrigins is the legacy comma-separated explicit allow-list. Still
	// honoured but the recommended way is now AllowedApex (wildcards).
	AllowedOrigins string `mapstructure:"allowed_origins"`
	// AllowedApex is the bare apex domain (e.g. "lssgoo.com"). Both the apex
	// and every subdomain (`*.<apex>`) are allowed over https. Empty disables
	// wildcard matching and falls back to AllowedOrigins only.
	AllowedApex string `mapstructure:"allowed_apex"`
}

// Auth holds only the essentials for a JWT-based auth flow.
// Add policy knobs (lockout, invite expiry, RBAC thresholds) only when the
// real module that needs them is added.
type Auth struct {
	JWTSecret            string `mapstructure:"jwt_secret"`
	AccessTokenMinutes   int    `mapstructure:"access_token_minutes"`
	RefreshTokenDays     int    `mapstructure:"refresh_token_days"`
	PasswordResetBaseURL string `mapstructure:"password_reset_base_url"`
	// HandoffTTLSeconds is how long a single-use SSO handoff token is valid.
	// Used by the apex → tenant subdomain login redirect. Keep short.
	HandoffTTLSeconds int `mapstructure:"handoff_ttl_seconds"`
}

type SMTP struct {
	Host      string `mapstructure:"host"`
	Port      int    `mapstructure:"port"`
	Username  string `mapstructure:"username"`
	Password  string `mapstructure:"password"`
	FromEmail string `mapstructure:"from_email"`
	FromName  string `mapstructure:"from_name"`
}

type Assets struct {
	S3Region          string `mapstructure:"s3_region"`
	S3Bucket          string `mapstructure:"s3_bucket"`
	S3KeyPrefix       string `mapstructure:"s3_key_prefix"`
	S3PublicBaseURL   string `mapstructure:"s3_public_base_url"`
	PresignTTLSeconds int    `mapstructure:"presign_ttl_seconds"`
}

type Developer struct {
	APIKeyPepper string `mapstructure:"api_key_pepper"`
}

// loadDotenvFiles loads env files in a stable order:
//  1. ENV_FILE (absolute path or path relative to cwd) — used in Docker/systemd when a single file is mounted.
//  2. .env from module root then cwd.
//  3. If APP_ENV=staging|production, overload .env.staging or .env.prod from module root then cwd.
//  4. .env.local from module root then cwd (local overrides).
//
// joho/godotenv does not override variables already set in the environment, so
// `APP_ENV=staging make run-api` keeps staging even if .env contains APP_ENV=development.
func loadDotenvFiles() {
	wd, _ := os.Getwd()
	if wd == "" {
		wd = "."
	}
	modRoot := ""
	for d := wd; d != ""; {
		if _, err := os.Stat(filepath.Join(d, "go.mod")); err == nil {
			modRoot = d
			break
		}
		parent := filepath.Dir(d)
		if parent == d {
			break
		}
		d = parent
	}

	load := func(path string) { _ = godotenv.Load(path) }
	overload := func(path string) { _ = godotenv.Overload(path) }

	if ef := strings.TrimSpace(os.Getenv("ENV_FILE")); ef != "" {
		p := ef
		if !filepath.IsAbs(p) {
			p = filepath.Join(wd, ef)
		}
		overload(p)
	} else {
		if modRoot != "" {
			load(filepath.Join(modRoot, ".env"))
		}
		load(filepath.Join(wd, ".env"))

		env := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
		switch env {
		case "staging", "stage":
			if modRoot != "" {
				overload(filepath.Join(modRoot, ".env.staging"))
			}
			overload(filepath.Join(wd, ".env.staging"))
		case "production", "prod":
			if modRoot != "" {
				overload(filepath.Join(modRoot, ".env.prod"))
			}
			overload(filepath.Join(wd, ".env.prod"))
		}
	}

	if modRoot != "" {
		overload(filepath.Join(modRoot, ".env.local"))
	}
	overload(filepath.Join(wd, ".env.local"))
}

// Load returns application config and enforces production validation rules.
func Load() (*Config, error) {
	loadDotenvFiles()

	v := viper.New()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	v.SetDefault("app.env", "development")
	v.SetDefault("app.base_domain", "lvh.me")
	v.SetDefault("app.frontend_scheme", "http")
	v.SetDefault("app.frontend_port", "3000")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.read_timeout_sec", 30)
	v.SetDefault("server.write_timeout_sec", 30)
	v.SetDefault("server.idle_timeout_sec", 120)
	v.SetDefault("server.max_header_bytes", 1048576)
	v.SetDefault("server.mode", "debug")

	v.SetDefault("postgres.host", "localhost")
	v.SetDefault("postgres.port", 5432)
	v.SetDefault("postgres.user", "postgres")
	v.SetDefault("postgres.password", "postgres")
	v.SetDefault("postgres.database", "app_db")
	v.SetDefault("postgres.sslmode", "disable")
	v.SetDefault("postgres.max_open_conns", 50)
	v.SetDefault("postgres.max_idle_conns", 25)
	v.SetDefault("postgres.conn_max_life_sec", 300)

	v.SetDefault("redis.addr", "localhost:6379")
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 2)
	v.SetDefault("redis.queue_db", 3)

	v.SetDefault("cors.allowed_origins", "http://localhost:3000,http://127.0.0.1:3000,http://lvh.me:3000")
	v.SetDefault("cors.allowed_apex", "")

	v.SetDefault("smtp.host", "")
	v.SetDefault("smtp.port", 587)
	v.SetDefault("smtp.username", "")
	v.SetDefault("smtp.password", "")
	v.SetDefault("smtp.from_email", "noreply@example.com")
	v.SetDefault("smtp.from_name", "App")

	v.SetDefault("auth.jwt_secret", devJWTPlaceholder)
	v.SetDefault("auth.access_token_minutes", 15)
	v.SetDefault("auth.refresh_token_days", 7)
	v.SetDefault("auth.password_reset_base_url", "")
	v.SetDefault("auth.handoff_ttl_seconds", 60)

	v.SetDefault("assets.s3_region", "")
	v.SetDefault("assets.s3_bucket", "")
	v.SetDefault("assets.s3_key_prefix", "uploads/")
	v.SetDefault("assets.s3_public_base_url", "")
	v.SetDefault("assets.presign_ttl_seconds", 900)

	v.SetDefault("developer.api_key_pepper", "")

	var c Config
	if err := v.Unmarshal(&c); err != nil {
		return nil, err
	}

	if c.Server.Port <= 0 {
		c.Server.Port = 8080
	}
	if c.Server.IdleTimeoutSec <= 0 {
		c.Server.IdleTimeoutSec = 120
	}
	if c.Server.MaxHeaderBytes <= 0 {
		c.Server.MaxHeaderBytes = 1 << 20
	}

	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &c, nil
}

const devJWTPlaceholder = "dev-change-me-in-production-min-32-chars-xx"

func countAllowedCORSOrigins(raw string) int {
	if strings.TrimSpace(raw) == "" {
		return 0
	}
	n := 0
	for _, p := range strings.Split(raw, ",") {
		if strings.TrimSpace(p) != "" {
			n++
		}
	}
	return n
}

// Validate enforces production-safety rules after defaults and env overrides are applied.
func (c *Config) Validate() error {
	env := strings.ToLower(strings.TrimSpace(c.App.Env))
	isProd := env == "production" || env == "prod"
	if !isProd {
		return nil
	}
	// In production we require either an explicit allow-list or a configured
	// apex domain (the apex covers `*.<apex>` via wildcard match).
	if countAllowedCORSOrigins(c.CORS.AllowedOrigins) == 0 && strings.TrimSpace(c.CORS.AllowedApex) == "" {
		return fmt.Errorf("config: APP_ENV=%q requires either CORS_ALLOWED_ORIGINS or CORS_ALLOWED_APEX to be set", c.App.Env)
	}
	// Reject wildcard origins in production — a single "*" cancels every CSRF
	// and origin-pinning benefit the explicit allow-list provides.
	for _, p := range strings.Split(c.CORS.AllowedOrigins, ",") {
		if strings.TrimSpace(p) == "*" {
			return fmt.Errorf("config: APP_ENV=%q forbids the wildcard \"*\" in CORS_ALLOWED_ORIGINS", c.App.Env)
		}
	}
	secret := strings.TrimSpace(c.Auth.JWTSecret)
	if len(secret) < 32 {
		return fmt.Errorf("config: AUTH_JWT_SECRET must be at least 32 characters in production")
	}
	if secret == devJWTPlaceholder {
		return fmt.Errorf("config: AUTH_JWT_SECRET must not use the development default in production")
	}
	return nil
}
