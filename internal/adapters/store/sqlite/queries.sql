-- name: append-message
INSERT INTO messages (session_id, role, text, created_at, tool_calls, tool_call_id)
VALUES (?, ?, ?, ?, ?, ?);

-- name: history
SELECT role, text, created_at, tool_calls, tool_call_id
FROM (
         SELECT * FROM messages WHERE session_id = ? ORDER BY created_at DESC LIMIT ?
     ) ORDER BY created_at ASC;