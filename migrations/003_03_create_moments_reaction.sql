CREATE TABLE IF NOT EXISTS moment_reactions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    moment_id INTEGER NOT NULL,
    fingerprint_id INTEGER NOT NULL,
    reaction TEXT NOT NULL DEFAULT '👍' CHECK (reaction IN (
        '👍',
        '👎',
        '❤',
        '👀',
        '💩'
    )),
    created_at INTEGER NOT NULL DEFAULT (strftime('%s','now')),

    FOREIGN KEY (moment_id) REFERENCES moments(id) ON DELETE CASCADE,
    FOREIGN KEY (fingerprint_id) REFERENCES fingerprints(id) ON DELETE CASCADE,
    UNIQUE (moment_id, fingerprint_id, reaction)
);

CREATE INDEX IF NOT EXISTS idx_moment_reaction ON moment_reactions (moment_id, reaction);
CREATE INDEX IF NOT EXISTS idx_moment_reaction_fingerprint ON moment_reactions (moment_id, fingerprint_id);
