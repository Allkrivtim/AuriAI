package sqlite

import (
	"AuriAI/internal/core"
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"

	"github.com/qustavo/dotsql"
	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

//go:embed queries.sql
var queriesSQL string

func NewStore(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(schemaSQL); err != nil {
		return nil, err
	}

	dot, err := dotsql.LoadFromString(queriesSQL)
	if err != nil {
		return nil, err
	}

	return &Store{db: db, dot: dot}, nil
}

func (s *Store) Close() error {
	err := s.db.Close()
	return err
}

func (s *Store) AppendMessage(ctx context.Context, SessionID string, m core.Message) error {
	var toolCalls any // nil → SQLite запишет NULL
	if len(m.ToolCalls) > 0 {
		b, err := json.Marshal(m.ToolCalls)
		if err != nil {
			return err
		}
		toolCalls = string(b)
	}
	_, err := s.dot.ExecContext(ctx, s.db, "append-message", SessionID, string(m.Role), m.Text, m.CreatedAt, toolCalls, m.ToolCallID)
	return err
}

func (s *Store) AppendMemory(ctx context.Context, m core.Memory) error {
	_, err := s.dot.ExecContext(ctx, s.db, "append-memory", m.Name, m.Text, m.CreatedAt)
	return err
}

func (s *Store) AppendNote(ctx context.Context, n core.Note) error {
	if n.Notification > 1 {
		n.Notification = 1
	}
	_, err := s.dot.ExecContext(ctx, s.db, "append-note", n.Name, n.Text, n.CreatedAt, n.ExpireIn, n.Notification, n.Pin)
	return err
}

func (s *Store) History(ctx context.Context, sessionID string, limit int) ([]core.Message, error) {
	rows, err := s.dot.QueryContext(ctx, s.db, "history", sessionID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []core.Message
	for rows.Next() {
		var message core.Message
		var role string
		var toolCallsJSON sql.NullString
		var toolCallID sql.NullString
		if err := rows.Scan(&role, &message.Text, &message.CreatedAt, &toolCallsJSON, &toolCallID); err != nil {
			return nil, err
		}
		message.Role = core.Role(role)
		if toolCallsJSON.Valid && toolCallsJSON.String != "" {
			if err := json.Unmarshal([]byte(toolCallsJSON.String), &message.ToolCalls); err != nil {
				return nil, err
			}
		}
		if toolCallID.Valid {
			message.ToolCallID = toolCallID.String
		}

		msgs = append(msgs, message)
	}
	return msgs, rows.Err()
}

func (s *Store) Memories(ctx context.Context) ([]core.Memory, error) {
	rows, err := s.dot.QueryContext(ctx, s.db, "memories")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memories []core.Memory
	for rows.Next() {
		var memory core.Memory
		if err := rows.Scan(&memory.Name, &memory.Text, &memory.CreatedAt); err != nil {
			return nil, err
		}
		memories = append(memories, memory)
	}
	return memories, rows.Err()
}

func (s *Store) Notes(ctx context.Context) ([]core.Note, error) {
	rows, err := s.dot.QueryContext(ctx, s.db, "notes")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []core.Note
	for rows.Next() {
		var note core.Note
		if err := rows.Scan(&note.Name, &note.Text, &note.CreatedAt, &note.ExpireIn, &note.Notification, &note.Pin); err != nil {
			return nil, err
		}
		if note.Notification > 1 {
			note.Notification = 1
		}
		notes = append(notes, note)
	}
	return notes, rows.Err()
}

var _ core.MemoryStore = (*Store)(nil)
var _ core.Store = (*Store)(nil)

func (s *Store) DeleteMemory(ctx context.Context, m core.Memory) error {
	_, err := s.dot.ExecContext(ctx, s.db, "delete-memory", m.Name)
	return err
}

func (s *Store) DeleteNote(ctx context.Context, n core.Note) error {
	_, err := s.dot.ExecContext(ctx, s.db, "delete-note", n.Name)
	return err
}
