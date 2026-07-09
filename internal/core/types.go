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
	Role      Role
	Text      string
	CreatedAt time.Time
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
