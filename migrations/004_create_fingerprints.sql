CREATE TABLE IF NOT EXISTS fingerprints (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    fingerprint TEXT NOT NULL,
    user_agent TEXT,
    ip TEXT,
    permissions_level TEXT NOT NULL DEFAULT 'normal' CHECK ( permissions_level IN (
        'normal',
        'friend',
        'admin'
    )),
    created_at INTEGER NOT NULL DEFAULT (strftime('%s','now'))
);

CREATE INDEX IF NOT EXISTS idx_fingerprints_fingerprint ON fingerprints (fingerprint);
