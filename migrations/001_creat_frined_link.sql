-- 创建友链表
CREATE TABLE IF NOT EXISTS friend_link (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  website_url TEXT NOT NULL,
  website_name TEXT NOT NULL,
  website_icon_url TEXT ,
  description TEXT NOT NULL,
  email TEXT,
  times INTEGER NOT NULL DEFAULT 0,
  status TEXT NOT NULL DEFAULT 'survival' CHECK ( status IN (
    'survival',
    'timeout',
    'error',
    'died',
    'pending'
  )),
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_friend_link_status ON friend_link (status);