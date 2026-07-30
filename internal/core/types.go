package core

import "time"

type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
	RoleTool      Role = "tool"
)

type Session struct {
	SessionId string
	Messages  []Message
}

type Message struct {
	Role       Role
	Text       string
	CreatedAt  time.Time
	ToolCalls  []ToolCall
	ToolCallID string
}

type InboundMessage struct {
	SessionID string
	Provider  string
	Text      string
}

type OutboundMessage struct {
	Text      string
	SessionID string
}

type CompletionRequest struct {
	System   string
	Messages []Message
	Tools    []ToolSpec
}

type CompletionResponse struct {
	Text      string
	ToolCalls []ToolCall
}

type ToolSpec struct {
	Name        string
	Description string
	Parameters  map[string]any
}

type ToolCall struct {
	ID   string
	Name string
	Args string
}

type Memory struct {
	Name      string
	Text      string
	CreatedAt time.Time
}

type Note struct {
	Name         string
	Text         string
	ExpireIn     int
	Notification uint
	Pin          uint
	CreatedAt    time.Time
}
