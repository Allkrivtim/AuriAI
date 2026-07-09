package core

import (
	"context"
)

type Engine interface {
	Handle(ctx context.Context, message InboundMessage) (OutboundMessage, error)
}

type LLM interface {
	Complete(ctx context.Context, request CompletionRequest) (CompletionResponse, error)
}

type Store interface {
	AppendMessage(ctx context.Context, sessionID string, m Message) error
	History(ctx context.Context, sessionID string, limit int) ([]Message, error)
}
