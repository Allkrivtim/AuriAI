package openai

import (
	"AuriAI/internal/core"
	"context"
	"errors"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

func NewClient(apiKey string, model string, url string) *Client {
	return &Client{
		api: openai.NewClient(
			option.WithBaseURL(url),
			option.WithAPIKey(apiKey),
		),
		model: model,
	}
}

// конвертер спеков тулзов в формат openai
func toOpenAITools(specs []core.ToolSpec) []openai.ChatCompletionToolParam {
	tools := make([]openai.ChatCompletionToolParam, 0, len(specs))
	for _, s := range specs {
		tools = append(tools, openai.ChatCompletionToolParam{
			Function: openai.FunctionDefinitionParam{
				Name:        s.Name,
				Description: openai.String(s.Description),
				Parameters:  openai.FunctionParameters(s.Parameters),
			},
		})
	}
	return tools
}

func (c *Client) Complete(ctx context.Context, request core.CompletionRequest) (core.CompletionResponse, error) {
	var msgs []openai.ChatCompletionMessageParamUnion

	if request.System != "" {
		msgs = append(msgs, openai.SystemMessage(request.System))
	}

	for _, m := range request.Messages {
		switch m.Role {
		case core.RoleUser:
			msgs = append(msgs, openai.UserMessage(m.Text))
		case core.RoleSystem:
			msgs = append(msgs, openai.SystemMessage(m.Text))
		case core.RoleTool:
			msgs = append(msgs, openai.ToolMessage(m.Text, m.ToolCallID))
		case core.RoleAssistant:
			if len(m.ToolCalls) == 0 {
				msgs = append(msgs, openai.AssistantMessage(m.Text))
			} else {
				calls := make([]openai.ChatCompletionMessageToolCallParam, 0, len(m.ToolCalls))
				for _, tc := range m.ToolCalls {
					calls = append(calls, openai.ChatCompletionMessageToolCallParam{
						ID: tc.ID,
						Function: openai.ChatCompletionMessageToolCallFunctionParam{
							Name:      tc.Name,
							Arguments: tc.Args,
						},
					})
				}
				msgs = append(msgs, openai.ChatCompletionMessageParamUnion{
					OfAssistant: &openai.ChatCompletionAssistantMessageParam{
						ToolCalls: calls,
					},
				})
			}
		}
	}

	params := openai.ChatCompletionNewParams{
		Model:    c.model,
		Messages: msgs,
	}
	if len(request.Tools) > 0 {
		params.Tools = toOpenAITools(request.Tools)
	}

	completion, err := c.api.Chat.Completions.New(ctx, params)
	if err != nil {
		return core.CompletionResponse{}, err
	}
	if len(completion.Choices) == 0 {
		return core.CompletionResponse{}, errors.New("openai: empty response, no choices")
	}

	msg := completion.Choices[0].Message

	var toolCalls []core.ToolCall
	for _, tc := range msg.ToolCalls {
		toolCalls = append(toolCalls, core.ToolCall{
			ID:   tc.ID,
			Name: tc.Function.Name,
			Args: tc.Function.Arguments,
		})
	}

	return core.CompletionResponse{
		Text:      msg.Content,
		ToolCalls: toolCalls,
	}, nil
}

var _ core.LLM = (*Client)(nil)
