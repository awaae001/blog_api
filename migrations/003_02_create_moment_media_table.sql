CREATE TABLE IF NOT EXISTS moment_media (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  moment_id INTEGER NOT NULL,
  media_url TEXT NOT NULL,
  is_in_local INTEGER NOT NULL DEFAULT 0,
  local_path TEXT,
  media_type TEXT NOT NULL DEFAULT 'other' CHECK ( media_type IN (
    'image',
    'video',
    'audio',
    'other'
  )),
  created_at INTEGER NOT NULL DEFAULT (strftime('%s','now')),

  -- 级联
  FOREIGN KEY (moment_id) REFERENCES moments(id) ON DELETE CASCADE
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_moment_media_moment_id ON moment_media(moment_id);
