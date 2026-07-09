package core

import (
	"context"
)

type engine struct {
	llm   LLM
	store Store
}

func NewEngine(llm LLM, store Store) Engine {
	return &engine{llm: llm, store: store}
}

func (e *engine) Handle(ctx context.Context, message InboundMessage) (OutboundMessage, error) {

	return OutboundMessage{}, nil
}
