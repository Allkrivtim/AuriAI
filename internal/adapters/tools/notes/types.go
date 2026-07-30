package memory

import "AuriAI/internal/core"

type Tool struct {
	store core.MemoryStore
}

func NewTool(store core.MemoryStore) *Tool {
	return &Tool{store: store}
}
