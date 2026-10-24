-- RSS 文章：同一订阅源内按 guid 去重，link 全局唯一，避免聚合页重复。
CREATE TABLE IF NOT EXISTS friend_rss_post (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    feed_id      INTEGER NOT NULL REFERENCES friend_rss(id) ON DELETE CASCADE,
    guid         TEXT    NOT NULL,
    title        TEXT    NOT NULL,
    link         TEXT    NOT NULL,
    description  TEXT    NOT NULL DEFAULT '',
    author       TEXT    NOT NULL DEFAULT '',
    published_at INTEGER NOT NULL,
    fetched_at   INTEGER NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_friend_rss_post_feed_guid ON friend_rss_post(feed_id, guid);
CREATE UNIQUE INDEX IF NOT EXISTS idx_friend_rss_post_link ON friend_rss_post(link);
CREATE INDEX IF NOT EXISTS idx_friend_rss_post_feed_published ON friend_rss_post(feed_id, published_at DESC);
CREATE INDEX IF NOT EXISTS idx_friend_rss_post_published ON friend_rss_post(published_at DESC);
