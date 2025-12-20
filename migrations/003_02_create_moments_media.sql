CREATE TABLE IF NOT EXISTS moments_media (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    moment_id INTEGER NOT NULL,
    name TEXT,
    media_url TEXT NOT NULL,
    media_type TEXT NOT NULL DEFAULT 'image' CHECK ( media_type IN (
        'image',
        'video'
    )),
    is_deleted INTEGER NOT NULL DEFAULT 0,

    FOREIGN KEY (moment_id) REFERENCES moments(id) ON DELETE CASCADE
);

-- 索引
CREATE INDEX IF NOT EXISTS idx_moments_media_moment_id ON moments_media (moment_id);