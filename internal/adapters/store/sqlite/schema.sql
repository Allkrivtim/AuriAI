CREATE TABLE IF NOT EXISTS messages (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT    NOT NULL,
    role       TEXT    NOT NULL,
    text       TEXT    NOT NULL,
    created_at   TIMESTAMP NOT NULL,
    tool_calls   TEXT,              -- JSON, nullable
    tool_call_id TEXT               -- nullable
);

CREATE INDEX IF NOT EXISTS idx_messages_session
    ON messages (session_id, created_at);