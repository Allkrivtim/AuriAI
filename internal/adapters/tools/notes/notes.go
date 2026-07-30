package notes

import (
	"AuriAI/internal/core"
	"AuriAI/internal/utils"
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
				"name":         map[string]any{"type": "string", "description": "Note name(unique)/key"},
				"text":         map[string]any{"type": "string", "description": "Note content (for save)"},
				"expire_at":    map[string]any{"type": "int", "description": "Note time to live in minutes. Set -1 too /key"},
				"notification": map[string]any{"type": "bool", "description": "Note notification is enabled(notification will wake assistant when note expire.)/key"},
				"pin":          map[string]any{"type": "bool", "description": "Note pin if pinned, will displayed in system prompt(use only for important notes!)"},
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
		ExpireIn     int    `json:"expire_at"`
		Notification bool   `json:"notification"`
		Pin          bool   `json:"pin"`
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
		err := t.store.AppendNote(ctx, core.Note{Name: input.Name, Text: input.Text, ExpireIn: input.ExpireIn, Notification: uint(utils.BoolToInt(input.Notification)), Pin: uint(utils.BoolToInt(input.Pin)), CreatedAt: time.Now()})
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
