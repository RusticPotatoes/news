CREATE TABLE IF NOT EXISTS edition (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT,
    date TEXT,
    start_time DATETIME,
    end_time DATETIME,
    created DATETIME,
    sources TEXT,
    articles TEXT,
    categories TEXT,
    metadata TEXT,
    UNIQUE(name, date)
);

CREATE TABLE IF NOT EXISTS analytics (
    user_id INTEGER PRIMARY KEY AUTOINCREMENT,
    insertion_timestamp DATETIME,
    payload TEXT
);

CREATE TABLE IF NOT EXISTS articles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT,
    description TEXT,
    compressed_content BLOB,
    image_url TEXT,
    link TEXT UNIQUE,
    author TEXT,
    source_id INTEGER,
    layout_id INTEGER,
    timestamp DATETIME,
    ts TEXT,
    FOREIGN KEY(source_id) REFERENCES sources(id),
    FOREIGN KEY(layout_id) REFERENCES layouts(id)
);

CREATE TABLE IF NOT EXISTS sources (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    owner_id TEXT DEFAULT 'admin',
    name TEXT,
    url TEXT,
    feed_url TEXT,
    categories TEXT,
    disable_fetch BOOLEAN,
    last_fetch_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(owner_id, url)
);

CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE,
    created DATETIME,
    password_hash BLOB,
    is_admin BOOLEAN
);

CREATE TABLE IF NOT EXISTS layouts (
    id INTEGER PRIMARY KEY,
    size INTEGER,
    width INTEGER,
    title_size INTEGER,
    max_chars INTEGER,
    max_elements INTEGER
);

INSERT INTO layouts (id, size, width, title_size, max_chars, max_elements) VALUES
(1, 1, 2, 6, 200, 32),
(2, 2, 2, 6, 500, 32),
(3, 3, 2, 6, 2250, 32),
(4, 4, 4, 4, 2800, 45),
(5, 5, 4, 4, 4000, 45),
(6, 6, 6, 2, 3300, 60),
(0, 0, 12, 0, 0, 0);

CREATE TABLE IF NOT EXISTS feed_cache (
    URL TEXT PRIMARY KEY,
    Data BLOB,
    Expiry DATETIME
);