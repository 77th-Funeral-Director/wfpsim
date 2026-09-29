CREATE TABLE upgrade_history (
    id TEXT PRIMARY KEY NOT NULL,
    target_table TEXT NOT NULL CHECK (target_table IN ('simulations','submissions')),
    target_id TEXT NOT NULL,
    engine_hash TEXT NOT NULL,
    previous_row TEXT NOT NULL CHECK (json_valid(previous_row)),
    new_document TEXT NOT NULL CHECK (json_valid(new_document)),
    created_at INTEGER NOT NULL
);
CREATE INDEX upgrade_history_target ON upgrade_history(target_table, target_id, created_at);
