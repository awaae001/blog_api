-- 友链标签：tag 只服务于友链分组，name 不区分大小写地唯一。
CREATE TABLE IF NOT EXISTS friend_tag(
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT    NOT NULL COLLATE NOCASE UNIQUE,
    color      TEXT    NOT NULL DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    updated_at INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_friend_tag_sort ON friend_tag(sort_order, id);

CREATE TRIGGER IF NOT EXISTS trg_friend_tag_updated_at
AFTER UPDATE ON friend_tag
FOR EACH ROW
BEGIN
  UPDATE friend_tag SET updated_at = strftime('%s','now') WHERE id = OLD.id;
END;

-- 友链标签多对多。复合主键天然去重，不需要自增 id。
CREATE TABLE IF NOT EXISTS friend_link_tag (
  friend_link_id INTEGER NOT NULL REFERENCES friend_link(id) ON DELETE CASCADE,
  tag_id         INTEGER NOT NULL REFERENCES friend_tag(id) ON DELETE CASCADE,
  PRIMARY KEY (friend_link_id, tag_id)
);

CREATE INDEX IF NOT EXISTS idx_friend_link_tag_tag ON friend_link_tag(tag_id, friend_link_id);
