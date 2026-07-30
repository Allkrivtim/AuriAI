-- name: append-message
INSERT INTO messages (session_id, role, text, created_at, tool_calls, tool_call_id)
VALUES (?, ?, ?, ?, ?, ?);

-- name: append-memory
INSERT OR REPLACE INTO memories (name, text, created_at)
VALUES (?, ?, ?);

-- name: append-note
INSERT INTO notes (name, text, created_at, expire_in, notification)
VALUES (?, ?, ?, ?, ?);

-- name: history
SELECT role, text, created_at, tool_calls, tool_call_id
FROM (
         SELECT * FROM messages WHERE session_id = ? ORDER BY created_at DESC LIMIT ?
     ) ORDER BY created_at ASC;

-- name: memories
SELECT name, text, created_at
FROM memories ORDER BY created_at ASC;

-- name: notes
SELECT name, text, created_at, expire_in, notification, pin
FROM notes ORDER BY created_at ASC;

-- name: delete-memory
DELETE FROM memories WHERE name = ?;

-- name: delete-note
DELETE FROM notes WHERE name = ?;