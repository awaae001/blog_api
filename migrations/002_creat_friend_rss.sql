CREATE TABLE IF NOT EXISTS friend_rss (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    friend_link_id INTEGER NOT NULL,
    rss_url TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'survival' CHECK ( status IN (
    'survival',
    'timeout',
    'error',
    'not_found'
  )),
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
)