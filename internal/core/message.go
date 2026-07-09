package core

import "AssistantAI/internal/adapters/llm/openrouter"

func AskAI(message string) (string, error) {
	resp, err := openrouter.MessageAI(message)
	return resp, err
}
