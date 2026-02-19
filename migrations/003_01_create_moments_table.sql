CREATE TABLE IF NOT EXISTS moments (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  content TEXT NOT NULL,
  -- 这里原本要有一个 media ，现在请查看 003_02_create_moments_media.sql
  status TEXT NOT NULL DEFAULT 'visible' CHECK ( status IN (
    'visible',
    'hidden',
    'deleted'
  )),
  guild_id INTEGER,
  channel_id INTEGER ,
  message_id INTEGER,
  message_link TEXT,
  created_at INTEGER NOT NULL DEFAULT (strftime('%s','now')),
  updated_at INTEGER NOT NULL DEFAULT (strftime('%s','now'))
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_moments_status ON moments (status);
CREATE INDEX IF NOT EXISTS idx_moments_content ON moments (content);
-- 仅对来自频道/群消息（两者都 > 0）的记录做唯一约束。
DROP INDEX IF EXISTS idx_moments_chat_message;
CREATE UNIQUE INDEX IF NOT EXISTS idx_moments_chat_message
ON moments(channel_id, message_id)
WHERE channel_id > 0 AND message_id > 0;

-- 添加触发器，自动更新 updated_at 字段
CREATE TRIGGER IF NOT EXISTS trg_moments_updated_at
AFTER UPDATE ON moments
FOR EACH ROW
BEGIN
  UPDATE moments SET updated_at = strftime('%s','now') WHERE id = OLD.id;
END;