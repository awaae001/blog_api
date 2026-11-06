-- tag 表
CREATE TABLE IF NOT EXISTS tags(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    color TEXT,
    created_at INTEGER DEFAULT (strftime('%s', 'now')),
    updated_at INTEGER NOT NULL DEFAULT 0
);

CREATE TRIGGER IF NOT EXISTS trg_tag_updated_at
AFTER UPDATE ON tags
FOR EACH ROW
BEGIN
  UPDATE tags SET updated_at = strftime('%s','now') WHERE id = OLD.id;
END;
