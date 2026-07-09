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

func (e *engine) Handle(ctx context.Context, inmessage InboundMessage) (OutboundMessage, error) {
	message := Message{Role: RoleUser, Text: inmessage.Text}
	//Store message
	err := e.store.AppendMessage(ctx, inmessage.SessionID, message)
	if err != nil {
		return OutboundMessage{}, err
	}

	//Get history
	history, err := e.store.History(ctx, inmessage.SessionID, 50)
	if err != nil {
		return OutboundMessage{}, err
	}

	//Create response to LLM provider
	resp, err := e.llm.Complete(ctx, CompletionRequest{System: "", Messages: history})
	if err != nil {
		return OutboundMessage{}, err
	}

	//Return responce
	return OutboundMessage{resp.Text, inmessage.SessionID}, nil
}
