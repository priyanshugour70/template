package queue

import (
	"context"
	"encoding/json"
)

type Message struct {
	ID      string
	Stream  string
	Payload string
}

type Handler func(ctx context.Context, msg *Message) error

type Consumer interface {
	Consume(ctx context.Context, stream string, handler Handler) error
}

func DecodePayload(payload string, v interface{}) error {
	return json.Unmarshal([]byte(payload), v)
}
