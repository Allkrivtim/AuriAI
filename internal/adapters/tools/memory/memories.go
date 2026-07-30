package memory

import (
	"AuriAI/internal/core"
	"context"
	"encoding/json"
	"time"
)

func (t *Tool) Spec() core.ToolSpec {
	return core.ToolSpec{
		Name:        "memory_control",
		Description: "Store memories",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"action": map[string]any{
					"type":        "string",
					"enum":        []string{"save", "get", "delete"}, // ← ограничили варианты
					"description": "What to do: save a new memory, get all, or delete one",
				},
				"name": map[string]any{"type": "string", "description": "Memory name/key"},
				"text": map[string]any{"type": "string", "description": "Memory content (for save)"},
			},
			"required": []string{"action"},
		},
	}
}

func (t *Tool) Invoke(ctx context.Context, args string) (string, error) {
	var input struct {
		Action string `json:"action"`
		Name   string `json:"name"`
		Text   string `json:"text"`
	}
	if err := json.Unmarshal([]byte(args), &input); err != nil {
		return "", err
	}
	var result string

	switch input.Action {
	case "save":
		if input.Name == "" || input.Text == "" {
			return "Error: 'name' and 'text' are required for save.", nil
		}
		err := t.store.AppendMemory(ctx, core.Memory{Name: input.Name, Text: input.Text, CreatedAt: time.Now()})
		if err != nil {
			return "", err
		}
		result = "Memory successfully saved."
	case "get":
		memories, err := t.store.Memories(ctx)
		if err != nil {
			return "", err
		}
		output, err := json.Marshal(memories)
		if err != nil {
			return "", err
		}
		result = string(output)
	case "delete":
		if input.Name == "" {
			return "Error: 'name' is required for delete.", nil
		}
		err := t.store.DeleteMemory(ctx, core.Memory{Name: input.Name})
		if err != nil {
			return "", err
		}
		return "Memory successfully deleted.", nil
	default:
		return "Error: unknown action '" + input.Action + "'. Use 'save', 'get' or 'delete'.", nil
	}

	return result, nil
}
