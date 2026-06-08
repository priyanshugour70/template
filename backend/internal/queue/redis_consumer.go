package queue

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisConsumer struct {
	client *redis.Client
}

func NewRedisConsumer(client *redis.Client) *RedisConsumer {
	return &RedisConsumer{client: client}
}

func (rc *RedisConsumer) Consume(ctx context.Context, channel string, handler Handler) error {
	pubsub := rc.client.Subscribe(ctx, channel)
	defer func() { _ = pubsub.Close() }()

	ch := pubsub.Channel(redis.WithChannelSize(64))

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-ch:
			if !ok {
				return nil
			}
			if msg == nil {
				continue
			}
			m := &Message{ID: "", Stream: channel, Payload: msg.Payload}
			if err := handler(ctx, m); err != nil {
				// Pub/Sub has no ack; brief pause on handler errors.
				time.Sleep(500 * time.Millisecond)
			}
		}
	}
}
