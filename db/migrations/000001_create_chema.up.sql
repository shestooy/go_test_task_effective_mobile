CREATE TABLE IF NOT EXISTS songs(
    id SERIAL PRIMARY KEY,
    group_name VARCHAR(255) NOT NULL,
    song    VARCHAR(255) NOT NULL,
    release_date TEXT,
    text TEXT,
    link VARCHAR(255) NOT NULL,
    UNIQUE (group_name, song)
);

CREATE INDEX IF NOT EXISTS idx_group ON songs(group_name);
CREATE INDEX IF NOT EXISTS idx_song ON songs(song);
CREATE INDEX IF NOT EXISTS idx_release_date ON songs(release_date);