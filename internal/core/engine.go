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
		resp, err = e.llm.Complete(ctx, CompletionRequest{System: e.systemPrompt(), Messages: history, Tools: e.tools.Specs()})
		if err != nil {
			return OutboundMessage{}, err
		}

		if len(resp.ToolCalls) == 0 { // <- если toolcalls нету, просто сохраняем сообщение и возвращаем OutboundMessage. Но кто у нас использует Handle()? Он сам должен получить OutboundMessage и сделать LLM.Complete()? Не совсем понятно.
			//Store AI response
			err = e.store.AppendMessage(ctx, inmessage.SessionID, Message{Role: RoleAssistant, Text: resp.Text, CreatedAt: time.Now()})
			if err != nil {
				return OutboundMessage{}, err
			}
			return OutboundMessage{resp.Text, inmessage.SessionID}, nil
		} // <- если вызов тулзы есть, (тут кстати непонятно, LLM может вызвать только одну тулзу или несколько за сообщение?) мы сохраняем сообщение а затем <|читай ниже|>.
		err = e.store.AppendMessage(ctx, inmessage.SessionID, Message{
			Role:      RoleAssistant,
			Text:      resp.Text,      // может быть пустым — норм
			ToolCalls: resp.ToolCalls, // ← вот они
			CreatedAt: time.Now(),
		})
		if err != nil {
			return OutboundMessage{}, err
		}
		for _, call := range resp.ToolCalls { //парсим все вызовы
			var result string                  // сохраняем переменную для результата тулзы, да?
			tool, ok := e.tools.Get(call.Name) //вызываем??? не совсем понятно здесь. не, мы получаем имя тулзы, да? не совсем понятно тут.
			if !ok {                           // если ok не true(не совсем понятно, что такое ok. Типо, подтверждение вызова тулзы или подтверждение сущестования тулзы?), возвращаем в result строку
				result = "Error: unknown tool " + call.Name
			} else {
				output, err := tool.Invoke(ctx, call.Args) // а здесь мы уже вызываем тулзу, да? есть контекст и аргументы. а где мы аргументы парсим?
				if err != nil {
					result = "Error: " + err.Error()
				} else {
					result = output
				}
			}
			err = e.store.AppendMessage(ctx, inmessage.SessionID, Message{Role: RoleTool, Text: result, CreatedAt: time.Now(), ToolCallID: call.ID}) //ну и добавляем вызов тулзы в историю, не совсем понятно откуда брать ToolCallID
			if err != nil {
				return OutboundMessage{Text: "Error", SessionID: inmessage.SessionID}, err
			}
		}
	}

	//Return response
	return OutboundMessage{Text: resp.Text, SessionID: inmessage.SessionID}, nil
}

func (e *engine) systemPrompt() string {
	sysPrompt := e.basePrompt
	sysPrompt = sysPrompt + "\n\n - Время: " + time.Now().Format("Monday, 2 January 2006, 15:04")
	//sysPrompt = sysPrompt + "\n\n - Заметки"
	return sysPrompt
}
