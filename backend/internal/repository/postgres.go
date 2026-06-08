package repository

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type DBConfig struct {
	DSN            string
	MaxOpenConns   int
	MaxIdleConns   int
	ConnMaxLifeSec int
	LogLevel       string
}

// OpenPostgres opens a GORM connection over the pgx driver and tunes the pool.
func OpenPostgres(ctx context.Context, cfg DBConfig, log *zap.Logger) (*gorm.DB, error) {
	level := gormlogger.Warn
	switch cfg.LogLevel {
	case "silent":
		level = gormlogger.Silent
	case "error":
		level = gormlogger.Error
	case "info":
		level = gormlogger.Info
	}

	db, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{
		Logger:                                   gormlogger.Default.LogMode(level),
		DisableForeignKeyConstraintWhenMigrating: true,
		PrepareStmt:                              true,
	})
	if err != nil {
		return nil, fmt.Errorf("gorm open: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("underlying sql.DB: %w", err)
	}

	if cfg.MaxOpenConns <= 0 {
		cfg.MaxOpenConns = 50
	}
	if cfg.MaxIdleConns <= 0 {
		cfg.MaxIdleConns = 25
	}
	if cfg.ConnMaxLifeSec <= 0 {
		cfg.ConnMaxLifeSec = 300
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifeSec) * time.Second)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(pingCtx); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("postgres ping: %w", err)
	}

	log.Info("PostgreSQL connected (GORM)",
		zap.Int("maxOpen", cfg.MaxOpenConns),
		zap.Int("maxIdle", cfg.MaxIdleConns),
	)
	return db, nil
}

// AutoMigrate is a thin pass-through. Only use in dev or for module-local helper
// tables; production schema should live in migrations/postgres/.
func AutoMigrate(db *gorm.DB, models ...interface{}) error {
	if len(models) == 0 {
		return nil
	}
	return db.AutoMigrate(models...)
}
