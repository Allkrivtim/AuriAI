package sqlite

import (
	"AuriAI/internal/core"
	"context"
	"database/sql"
	"encoding/json"

	_ "embed"

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
	_, err := s.dot.Exec(s.db, "append-message", SessionID, string(m.Role), m.Text, m.CreatedAt, toolCalls, m.ToolCallID)
	return err
}

func (s *Store) History(ctx context.Context, sessionID string, limit int) ([]core.Message, error) {
	rows, err := s.dot.Query(s.db, "history", sessionID, limit)
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

var _ core.Store = (*Store)(nil)
