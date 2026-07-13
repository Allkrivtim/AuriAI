package core

import (
	"context"
	"time"
)

type engine struct {
	llm        LLM
	store      Store
	basePrompt string
	tools      ToolRegistry
}

func NewEngine(llm LLM, store Store, basePrompt string, tools ToolRegistry) Engine {
	return &engine{llm: llm, store: store, basePrompt: basePrompt, tools: tools}
}

func (e *engine) Handle(ctx context.Context, inmessage InboundMessage) (OutboundMessage, error) {
	message := Message{Role: RoleUser, Text: inmessage.Text, CreatedAt: time.Now()}
	//Store message
	err := e.store.AppendMessage(ctx, inmessage.SessionID, message)
	if err != nil {
		return OutboundMessage{}, err
	}
	var resp CompletionResponse

	for step := 0; step < 10; step++ {
		//Get history
		history, err := e.store.History(ctx, inmessage.SessionID, 50)
		if err != nil {
			return OutboundMessage{}, err
		}

		//Create response to LLM provider
		resp, err = e.llm.Complete(ctx, CompletionRequest{System: e.basePrompt, Messages: history, Tools: e.tools.Specs()})
		if err != nil {
			return OutboundMessage{}, err
		}

		if len(resp.ToolCalls) == 0 {
			//Store AI response
			err = e.store.AppendMessage(ctx, inmessage.SessionID, Message{Role: RoleAssistant, Text: resp.Text, CreatedAt: time.Now()})
			if err != nil {
				return OutboundMessage{}, err
			}
			return OutboundMessage{resp.Text, inmessage.SessionID}, nil
		} else {
			err = e.store.AppendMessage(ctx, inmessage.SessionID, Message{
				Role:      RoleAssistant,
				Text:      resp.Text,      // может быть пустым — норм
				ToolCalls: resp.ToolCalls, // ← вот они
				CreatedAt: time.Now(),
			})
			if err != nil {
				return OutboundMessage{}, err
			}
			var result string
			for _, call := range resp.ToolCalls {
				tool, ok := e.tools.Get(call.Name)
				if !ok {
					result = "error: unknown tool " + call.Name
					continue
				}
				out, err := tool.Invoke(ctx, call.Args)
				if err != nil {
					result = "error: " + err.Error()
				}
				result = out
			}
		}
	}

	//Return response
	return OutboundMessage{resp.Text, inmessage.SessionID}, nil
}
