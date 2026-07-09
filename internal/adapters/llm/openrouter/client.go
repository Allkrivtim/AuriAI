package openrouter

import (
	"context"
	"errors"
	"os"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

func MessageAI(message string) (string, error) {
	client := openai.NewClient(
		option.WithBaseURL("https://openrouter.ai/api/v1"),
		option.WithAPIKey(os.Getenv("OPENROUTER_API_KEY")),
	)

	resp, err := client.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
		Model: "google/gemma-4-31b-it:free",
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(message),
		},
	})

	if err != nil {
		return "Error", err
	}
	if len(resp.Choices) == 0 {
		return "AI вернул пустой ответ (нет choices)", errors.New("AI вернул пустой ответ (нет choices)")
	}
	return resp.Choices[0].Message.Content, err
}
