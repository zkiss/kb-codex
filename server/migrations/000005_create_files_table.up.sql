CREATE TABLE IF NOT EXISTS files (
    kb_id INTEGER NOT NULL REFERENCES knowledge_bases(id) ON DELETE CASCADE,
    file_name TEXT NOT NULL,
    lookup_name TEXT NOT NULL,
    mime_type TEXT NOT NULL,
    content BYTEA NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (kb_id, lookup_name)
);
