package openrouter

import (
	"AssistantAI/internal/core"
	"context"
	"errors"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// NewClient создаёт LLM-провайдер поверх OpenRouter (Chat Completions API).
// apiKey и model приходят снаружи (из main/env) — адаптер сам в окружение не лезет.
func NewClient(apiKey string, model string) *Client {
	return &Client{
		api: openai.NewClient(
			option.WithBaseURL("https://openrouter.ai/api/v1"),
			option.WithAPIKey(apiKey),
		),
		model: model,
	}
}

// Complete реализует core.LLM: маппит историю ядра в формат openai-go,
// вызывает Chat Completions и возвращает текст ответа.
func (c *Client) Complete(ctx context.Context, request core.CompletionRequest) (core.CompletionResponse, error) {
	var msgs []openai.ChatCompletionMessageParamUnion

	// system prompt первым, если задан
	if request.System != "" {
		msgs = append(msgs, openai.SystemMessage(request.System))
	}

	// история диалога
	for _, m := range request.Messages {
		switch m.Role {
		case core.RoleUser:
			msgs = append(msgs, openai.UserMessage(m.Text))
		case core.RoleAssistant:
			msgs = append(msgs, openai.AssistantMessage(m.Text))
		case core.RoleSystem:
			msgs = append(msgs, openai.SystemMessage(m.Text))
		case core.RoleTool:
			// пропускаем — появится на этапе тулзов
		}
	}

	completion, err := c.api.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:    c.model,
		Messages: msgs,
	})
	if err != nil {
		return core.CompletionResponse{}, err
	}
	if len(completion.Choices) == 0 {
		return core.CompletionResponse{}, errors.New("openrouter: empty response, no choices")
	}

	return core.CompletionResponse{Text: completion.Choices[0].Message.Content}, nil
}

// Проверка на этапе компиляции, что *Client удовлетворяет интерфейсу core.LLM.
var _ core.LLM = (*Client)(nil)
