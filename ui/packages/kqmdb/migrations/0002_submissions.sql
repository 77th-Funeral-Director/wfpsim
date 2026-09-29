ALTER TABLE simulations ADD COLUMN source TEXT NOT NULL DEFAULT 'upstream';

CREATE TABLE submissions (
    id TEXT PRIMARY KEY NOT NULL,
    request_hash TEXT NOT NULL,
    source_url TEXT NOT NULL,
    document TEXT NOT NULL CHECK (json_valid(document)),
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','approved','rejected')),
    submitted_at INTEGER NOT NULL,
    reviewed_at INTEGER,
    reviewed_by TEXT,
    reason TEXT NOT NULL DEFAULT '',
    published_id TEXT,
    previous_document TEXT
);
CREATE INDEX submissions_queue ON submissions(status, submitted_at, id);
