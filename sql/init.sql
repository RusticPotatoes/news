CREATE TABLE IF NOT EXISTS edition (
    id TEXT PRIMARY KEY,
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
    user_id TEXT,
    insertion_timestamp DATETIME,
    payload TEXT
);

CREATE TABLE IF NOT EXISTS articles (
    id TEXT PRIMARY KEY,
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
    id TEXT PRIMARY KEY,
    owner_id TEXT,
    name TEXT,
    url TEXT,
    feed_url TEXT,
    categories TEXT,
    disable_fetch BOOLEAN
);

CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    name TEXT,
    created DATETIME,
    password_hash BLOB,
    is_admin BOOLEAN
);