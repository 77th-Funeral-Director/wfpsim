CREATE TABLE simulations (
    id TEXT PRIMARY KEY NOT NULL,
    document TEXT NOT NULL CHECK (json_valid(document)),
    create_date INTEGER NOT NULL,
    dps REAL NOT NULL,
    duration REAL NOT NULL,
    imported_at INTEGER NOT NULL,
    seen_run TEXT NOT NULL,
    visible INTEGER NOT NULL DEFAULT 1
);
CREATE INDEX simulations_created ON simulations(create_date, id);
CREATE INDEX simulations_dps ON simulations(dps, id);
CREATE INDEX simulations_duration ON simulations(duration, id);

CREATE TABLE sync_state (
    key TEXT PRIMARY KEY NOT NULL,
    value TEXT NOT NULL
);
