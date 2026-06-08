package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/your-org/your-service/internal/config"
	"github.com/your-org/your-service/internal/pkg/logger"
	"github.com/your-org/your-service/internal/queue"
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

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.QueueDB,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatal("redis ping failed", zap.Error(err))
	}
	defer rdb.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	consumer := queue.NewRedisConsumer(rdb)

	// Register one consumer per channel. Replace noopHandler with your real handlers,
	// or call into internal/jobs/* for scheduled work.
	for _, ch := range queue.DefaultChannels() {
		channel := ch
		go runConsumer(ctx, consumer, channel, noopHandler, log)
	}

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

func noopHandler(_ context.Context, _ *queue.Message) error { return nil }
