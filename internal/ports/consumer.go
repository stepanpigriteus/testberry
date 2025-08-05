package ports

import (
	"context"
)

type MessageBroker interface {
    Consume(ctx context.Context, handler func(ctx context.Context, message []byte) error) error
}