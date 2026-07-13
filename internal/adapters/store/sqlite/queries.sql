-- name: append-message
INSERT INTO messages (session_id, role, text, created_at)
VALUES (?, ?, ?, ?);

-- name: history
SELECT role, text, created_at FROM (
    SELECT role, text, created_at
    FROM messages
    WHERE session_id = ?
    ORDER BY created_at DESC
    LIMIT ?
) ORDER BY created_at ASC;