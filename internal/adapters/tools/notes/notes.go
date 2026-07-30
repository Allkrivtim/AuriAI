package memory

import (
	"AuriAI/internal/core"
	"context"
	"encoding/json"
	"time"
)

func (t *Tool) Spec() core.ToolSpec {
	return core.ToolSpec{
		Name:        "notes_control",
		Description: "Store notes",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"action": map[string]any{
					"type":        "string",
					"enum":        []string{"save", "get", "delete"}, // ← ограничили варианты
					"description": "What to do: save a new note, get all, or delete one",
				},
				"name":         map[string]any{"type": "string", "description": "Note name/key"},
				"text":         map[string]any{"type": "string", "description": "Note content (for save)"},
				"expire_at":    map[string]any{"type": "string", "description": "Note time to live in minutes/key"},
				"notification": map[string]any{"type": "string", "description": "Note notification is enabled/key"},
			},
			"required": []string{"action"},
		},
	}
}

func (t *Tool) Invoke(ctx context.Context, args string) (string, error) {
	var input struct {
		Action       string `json:"action"`
		Name         string `json:"name"`
		Text         string `json:"text"`
		ExpireAt     uint   `json:"expire_at"`
		Notification uint   `json:"notification"`
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
		err := t.store.AppendNote(ctx, core.Note{Name: input.Name, Text: input.Text, ExpireIn: 1, Notification: input.Notification, CreatedAt: time.Now()})
		if err != nil {
			return "", err
		}
		result = "Note successfully saved."
	case "get":
		memories, err := t.store.Notes(ctx)
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
		err := t.store.DeleteNote(ctx, core.Note{Name: input.Name})
		if err != nil {
			return "", err
		}
		return "Note successfully deleted.", nil
	default:
		return "Error: unknown action '" + input.Action + "'. Use 'save', 'get' or 'delete'.", nil
	}

	return result, nil
}
