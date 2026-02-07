-- Media table to track local files and their status
CREATE TABLE IF NOT EXISTS media (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    path TEXT UNIQUE NOT NULL,
    hash TEXT NOT NULL,
    is_posted BOOLEAN DEFAULT FALSE,
    ignore BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Posts table to track carousel posts and scheduling
CREATE TABLE IF NOT EXISTS posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    caption TEXT,
    scheduled_at TIMESTAMP,
    published_at TIMESTAMP,
    ig_container_id TEXT,
    status TEXT CHECK(status IN ('draft', 'scheduled', 'publishing', 'published', 'failed')) DEFAULT 'draft',
    engagement_likes INTEGER DEFAULT 0,
    engagement_comments INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    alt_text TEXT,
    location_id TEXT,
    user_tags TEXT,
    share_to_feed INTEGER DEFAULT 0,
    cover_url TEXT,
    thumb_offset INTEGER DEFAULT 0,
    collaborators TEXT,
    audio_name TEXT,
    post_as_story INTEGER DEFAULT 0
);

-- Junction table for carousel items
CREATE TABLE IF NOT EXISTS post_media (
    post_id INTEGER,
    media_id INTEGER,
    display_order INTEGER,
    ig_item_container_id TEXT,
    PRIMARY KEY (post_id, media_id),
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    FOREIGN KEY (media_id) REFERENCES media(id)
);
-- Settings table for app configuration
CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    value TEXT
);
