CREATE TABLE IF NOT EXISTS todos (
  id         INTEGER PRIMARY KEY AUTOINCREMENT,
  title      TEXT    NOT NULL CHECK (length(title) BETWEEN 1 AND 100),
  done       INTEGER NOT NULL DEFAULT 0 CHECK (done IN (0, 1)),
  created_at TEXT    NOT NULL,
  updated_at TEXT    NOT NULL
);
