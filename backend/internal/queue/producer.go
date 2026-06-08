package queue

import (
	"context"
	"encoding/json"
)

// Producer publishes events to a Redis Pub/Sub channel.
type Producer interface {
	Publish(ctx context.Context, channel string, payload interface{}) error
}

func EncodePayload(v interface{}) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
