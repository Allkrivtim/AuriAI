package sqlite

import (
	"AssistantAI/internal/core"
	"context"
	"database/sql"

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
	_, err := s.dot.Exec(s.db, "append-message", SessionID, string(m.Role), m.Text, m.CreatedAt)
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
		var m core.Message
		var r string
		if err := rows.Scan(&r, &m.Text, &m.CreatedAt); err != nil {
			return nil, err
		}
		m.Role = core.Role(r)
		msgs = append(msgs, m)
	}
	return msgs, rows.Err()
}

var _ core.Store = (*Store)(nil)
