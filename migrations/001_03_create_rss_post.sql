CREATE TABLE IF NOT EXISTS friend_rss_post (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    rss_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    author TEXT,
    link TEXT NOT NULL,
    description TEXT NOT NULL,
    time INTEGER NOT NULL,

    -- 级联
    FOREIGN KEY (rss_id) REFERENCES friend_rss(id) ON DELETE CASCADE
)