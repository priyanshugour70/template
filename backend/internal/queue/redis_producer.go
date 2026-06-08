package queue

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type RedisProducer struct {
	client *redis.Client
}

func NewRedisProducer(client *redis.Client) *RedisProducer {
	return &RedisProducer{client: client}
}

func (p *RedisProducer) Publish(ctx context.Context, channel string, payload interface{}) error {
	data, err := EncodePayload(payload)
	if err != nil {
		return err
	}
	return p.client.Publish(ctx, channel, data).Err()
}
