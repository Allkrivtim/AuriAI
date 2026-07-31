package core

import (
	"context"
	"time"
)

type engine struct {
	llm         LLM
	store       Store
	memoryStore MemoryStore
	basePrompt  string
	tools       ToolRegistry
}

func NewEngine(llm LLM, store Store, memoryStore MemoryStore, basePrompt string, tools ToolRegistry) Engine {
	return &engine{llm: llm, store: store, memoryStore: memoryStore, basePrompt: basePrompt, tools: tools}
}

func (e *engine) Handle(ctx context.Context, inputMessage InboundMessage) (OutboundMessage, error) {
	message := Message{Role: RoleUser, Text: inputMessage.Text, CreatedAt: time.Now()}
	//Store message
	err := e.store.AppendMessage(ctx, inputMessage.SessionID, message)
	if err != nil {
		return OutboundMessage{}, err
	}
	var resp CompletionResponse

	for step := 0; step < 10; step++ {
		//Get history
		history, err := e.store.History(ctx, inputMessage.SessionID, 50)
		if err != nil {
			return OutboundMessage{}, err
		}

		//Create response to LLM provider
		resp, err = e.llm.Complete(ctx, CompletionRequest{System: e.systemPrompt(), Messages: history, Tools: e.tools.Specs()})
		if err != nil {
			return OutboundMessage{}, err
		}

		if len(resp.ToolCalls) == 0 {
			//Store AI response
			err = e.store.AppendMessage(ctx, inputMessage.SessionID, Message{Role: RoleAssistant, Text: resp.Text, CreatedAt: time.Now()})
			if err != nil {
				return OutboundMessage{}, err
			}
			return OutboundMessage{resp.Text, inputMessage.SessionID}, nil
		}
		err = e.store.AppendMessage(ctx, inputMessage.SessionID, Message{
			Role:      RoleAssistant,
			Text:      resp.Text,
			ToolCalls: resp.ToolCalls,
			CreatedAt: time.Now(),
		})
		if err != nil {
			return OutboundMessage{}, err
		}
		for _, call := range resp.ToolCalls {
			var result string
			tool, ok := e.tools.Get(call.Name)
			if !ok {
				result = "Error: unknown tool " + call.Name
			} else {
				output, err := tool.Invoke(ctx, call.Args)
				if err != nil {
					result = "Error: " + err.Error()
				} else {
					result = output
				}
			}
			err = e.store.AppendMessage(ctx, inputMessage.SessionID, Message{Role: RoleTool, Text: result, CreatedAt: time.Now(), ToolCallID: call.ID}) //ну и добавляем вызов тулзы в историю, не совсем понятно откуда брать ToolCallID
			if err != nil {
				return OutboundMessage{Text: "Error", SessionID: inputMessage.SessionID}, err
			}
		}
	}

	//Return response
	return OutboundMessage{Text: resp.Text, SessionID: inputMessage.SessionID}, nil
}

func (e *engine) systemPrompt() string {
	sysPrompt := e.basePrompt
	sysPrompt = sysPrompt + "\n\n - Время: " + time.Now().Format("Monday, 2 January 2006, 15:04")

	notesRaw, err := e.memoryStore.Notes(context.Background())
	if err != nil {
		return sysPrompt + "Не удалось обработать заметки."
	}
	var notes string
	for _, note := range notesRaw {
		if note.Pin == 0 {
			continue
		}
		text := "\n   - " + note.Name + ":" + note.Text
		notes = notes + text
	}
	sysPrompt = sysPrompt + "\n\n - Заметки:" + string(notes)
	return sysPrompt
}
