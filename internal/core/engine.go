package core

import (
	"context"
	"time"
)

type engine struct {
	llm        LLM
	store      Store
	basePrompt string
}

func NewEngine(llm LLM, store Store, basePrompt string) Engine {
	return &engine{llm: llm, store: store, basePrompt: basePrompt}
}

func (e *engine) Handle(ctx context.Context, inmessage InboundMessage) (OutboundMessage, error) {
	message := Message{Role: RoleUser, Text: inmessage.Text, CreatedAt: time.Now()}
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
	resp, err := e.llm.Complete(ctx, CompletionRequest{System: e.basePrompt, Messages: history})
	if err != nil {
		return OutboundMessage{}, err
	}

	//Store AI response
	err = e.store.AppendMessage(ctx, inmessage.SessionID, Message{Role: RoleAssistant, Text: resp.Text, CreatedAt: time.Now()})
	if err != nil {
		return OutboundMessage{}, err
	}

	//Return response
	return OutboundMessage{resp.Text, inmessage.SessionID}, nil
}
