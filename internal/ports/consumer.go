package ports

import (
	"context"
)

type MessageBroker interface {
	Consumer
	Producer
}

type Consumer interface {
	Consume(ctx context.Context, handler func(ctx context.Context, message []byte) error) error
}

type Producer interface {
	Send(key string, message []byte) error
}
