CREATE TABLE IF NOT EXISTS images (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT,
    url TEXT NOT NULL,
    local_path INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'normal' CHECK (status IN (
        'normal',
        'pause',
        'pending'
    ))
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_iamge_repo_url ON images (url);