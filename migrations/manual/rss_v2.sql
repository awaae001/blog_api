-- 旧版 RSS 表升级到 v2，手动执行：
--   sqlite3 data/blog.db < migrations/manual/rss_v2.sql
-- 先备份数据库，并在服务停止时执行。
--
-- 变化：
--   * friend_rss.rss_url -> feed_url，全局唯一；同一 URL 的重复行只保留 id 最小的
--   * friend_link_id = -1 或指向不存在的友链 -> NULL
--   * status/is_died/times -> status(active/paused/dead) + error_count
--   * friend_rss_post.rss_id -> feed_id，time -> published_at，新增 guid（旧数据用 link 填充）
--   * 清理同源「标题 + 发布时间」相同的重复文章（换域名/改固定链接导致），保留最后抓到的一条
--   * next_fetch_at = 0：服务启动后调度器会把所有源立刻抓一遍，顺带补齐 etag 等字段

PRAGMA foreign_keys = OFF;
BEGIN IMMEDIATE;

CREATE TABLE friend_rss_v2 (
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

INSERT INTO friend_rss_v2 (id, friend_link_id, name, feed_url, status, error_count, created_at, updated_at)
SELECT
    f.id,
    CASE WHEN EXISTS (SELECT 1 FROM friend_link l WHERE l.id = f.friend_link_id) THEN f.friend_link_id END,
    f.name,
    f.rss_url,
    CASE
        WHEN f.status = 'pause' THEN 'paused'
        WHEN f.is_died THEN 'dead'
        ELSE 'active'
    END,
    f.times,
    f.updated_at,
    f.updated_at
FROM friend_rss f
WHERE f.id IN (SELECT MIN(id) FROM friend_rss GROUP BY rss_url);

-- REFERENCES 直接写最终表名，改名后无需再修外键。
CREATE TABLE friend_rss_post_v2 (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    feed_id      INTEGER NOT NULL REFERENCES friend_rss(id) ON DELETE CASCADE,
    guid         TEXT    NOT NULL,
    title        TEXT    NOT NULL,
    link         TEXT    NOT NULL,
    description  TEXT    NOT NULL DEFAULT '',
    author       TEXT    NOT NULL DEFAULT '',
    published_at INTEGER NOT NULL,
    fetched_at   INTEGER NOT NULL,
    UNIQUE (feed_id, guid),
    UNIQUE (link)
);

-- 重复 feed 下的文章改挂到保留行；遗留库里 link 重复的文章只留第一条。
INSERT OR IGNORE INTO friend_rss_post_v2 (id, feed_id, guid, title, link, description, author, published_at, fetched_at)
SELECT
    p.id,
    (SELECT MIN(f2.id) FROM friend_rss f2 WHERE f2.rss_url = f.rss_url),
    p.link,
    p.title,
    p.link,
    p.description,
    COALESCE(p.author, ''),
    MAX(p.time, 0),
    MAX(p.time, 0)
FROM friend_rss_post p
JOIN friend_rss f ON f.id = p.rss_id
ORDER BY p.id;

DROP TABLE friend_rss_post;
DROP TABLE friend_rss;
ALTER TABLE friend_rss_v2 RENAME TO friend_rss;
ALTER TABLE friend_rss_post_v2 RENAME TO friend_rss_post;

CREATE INDEX idx_friend_rss_due ON friend_rss(status, next_fetch_at);
CREATE INDEX idx_friend_rss_friend_link ON friend_rss(friend_link_id);
CREATE INDEX idx_friend_rss_updated ON friend_rss(updated_at DESC, id DESC);
DELETE FROM friend_rss_post
WHERE title <> '' AND published_at > 0
  AND id NOT IN (
    SELECT MAX(id) FROM friend_rss_post
    WHERE title <> '' AND published_at > 0
    GROUP BY feed_id, title, published_at
  );
CREATE UNIQUE INDEX idx_friend_rss_post_feed_title_time
    ON friend_rss_post(feed_id, title, published_at) WHERE title <> '' AND published_at > 0;
CREATE INDEX idx_friend_rss_post_feed_published ON friend_rss_post(feed_id, published_at DESC);
CREATE INDEX idx_friend_rss_post_published ON friend_rss_post(published_at DESC);

-- 005 里针对旧列建的索引已随旧表删除，这里只是兜底。
DROP INDEX IF EXISTS idx_friend_rss_link_status_died_updated_id;
DROP INDEX IF EXISTS idx_friend_rss_status_died_updated_id;

COMMIT;
PRAGMA foreign_keys = ON;
