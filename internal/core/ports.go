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
	Close() error
}

type Tool interface {
	Spec() ToolSpec
	Invoke(ctx context.Context, args string) (string, error)
}

type ToolRegistry interface {
	Register(t Tool)
	Get(name string) (Tool, bool)
	Specs() []ToolSpec
}

type MemoryStore interface {
	AppendMemory(ctx context.Context, memory Memory) error
	Memories(ctx context.Context) ([]Memory, error)
	DeleteMemory(ctx context.Context, memory Memory) error
	AppendNote(ctx context.Context, note Note) error
	Notes(ctx context.Context) ([]Note, error)
	DeleteNote(ctx context.Context, note Note) error
}
