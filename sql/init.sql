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
    metadata TEXT
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
    link TEXT,
    author TEXT,
    source TEXT,
    timestamp DATETIME,
    ts TEXT,
    layout TEXT
);

CREATE TABLE IF NOT EXISTS sources (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    owner_id TEXT,
    name TEXT,
    url TEXT,
    feed_url TEXT,
    categories TEXT,
    disable_fetch BOOLEAN
);

CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT,
    created DATETIME,
    password_hash BLOB,
    is_admin BOOLEAN
);