-- RSS 订阅源：每行一个 feed，自带抓取调度与条件请求缓存。
-- feed_url 全局唯一；friend_link_id 为 NULL 表示无归属，随友链级联删除。
CREATE TABLE IF NOT EXISTS friend_rss (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    friend_link_id  INTEGER REFERENCES friend_link(id) ON DELETE CASCADE,
    name            TEXT    NOT NULL,
    feed_url        TEXT    NOT NULL UNIQUE,
    site_url        TEXT    NOT NULL DEFAULT '',
    feed_type       TEXT    NOT NULL DEFAULT '',
    status          TEXT    NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'paused', 'dead')),
    error_count     INTEGER NOT NULL DEFAULT 0,
    last_error      TEXT    NOT NULL DEFAULT '',
    etag            TEXT    NOT NULL DEFAULT '',
    last_modified   TEXT    NOT NULL DEFAULT '',
    fetch_interval  INTEGER NOT NULL DEFAULT 0,
    next_fetch_at   INTEGER NOT NULL DEFAULT 0,
    last_fetched_at INTEGER NOT NULL DEFAULT 0,
    last_success_at INTEGER NOT NULL DEFAULT 0,
    created_at      INTEGER NOT NULL DEFAULT 0,
    updated_at      INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_friend_rss_due ON friend_rss(status, next_fetch_at);
CREATE INDEX IF NOT EXISTS idx_friend_rss_friend_link ON friend_rss(friend_link_id);
CREATE INDEX IF NOT EXISTS idx_friend_rss_updated ON friend_rss(updated_at DESC, id DESC);
