CREATE TABLE IF NOT EXISTS images (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT,
    url TEXT NOT NULL UNIQUE,
    local_path TEXT,
    is_local INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'normal' CHECK (status IN (
        'normal',
        'pause',
        'broken'
        'pending'
    ))
);