CREATE TABLE IF NOT EXISTS friend_rss_post (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    friend_rss_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    link TEXT NOT NULL,
    description TEXT NOT NULL,
    time TIMESTAMP NOT NULL,
);