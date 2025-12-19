CREATE TABLE IF NOT EXISTS moments (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  content TEXT NOT NULL,
  -- 这里原本要有一个 media ，但我懒得实现了
  status TEXT NOT NULL DEFAULT 'visible' CHECK ( status IN (
    'visible',
    'hidden',
    'deleted'
  )),
  guild_id INTEGER,
  channel_id INTEGER ,
  message_id INTEGER,
  created_at INTEGER NOT NULL DEFAULT (strftime('%s','now')),
  updated_at INTEGER NOT NULL DEFAULT (strftime('%s','now'))
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_moments_status ON moments (status);
CREATE INDEX IF NOT EXISTS idx_moments_content ON moments (content);
CREATE UNIQUE INDEX IF NOT EXISTS idx_moments_chat_message ON moments(channel_id, message_id);

-- 添加触发器，自动更新 updated_at 字段
CREATE TRIGGER IF NOT EXISTS trg_moments_updated_at
AFTER UPDATE ON moments
FOR EACH ROW
BEGIN
  UPDATE moments SET updated_at = strftime('%s','now') WHERE id = OLD.id;
END;