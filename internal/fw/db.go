package fw

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

const schema = `
CREATE TABLE IF NOT EXISTS state (
  key   TEXT PRIMARY KEY,
  value TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS events (
  id    INTEGER PRIMARY KEY AUTOINCREMENT,
  ts    DATETIME NOT NULL,
  type  TEXT NOT NULL,
  value TEXT NOT NULL DEFAULT '',
  meta  TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_events_ts   ON events(ts);
CREATE INDEX IF NOT EXISTS idx_events_type ON events(type);

CREATE TABLE IF NOT EXISTS blocks (
  id              INTEGER PRIMARY KEY AUTOINCREMENT,
  start_ts        DATETIME NOT NULL UNIQUE,
  end_ts          DATETIME NOT NULL,
  project         TEXT NOT NULL DEFAULT '',
  repo            TEXT NOT NULL DEFAULT '',
  focused_minutes INTEGER NOT NULL DEFAULT 0,
  switches        INTEGER NOT NULL DEFAULT 0,
  data            TEXT NOT NULL DEFAULT '',
  summary         TEXT NOT NULL DEFAULT '',
  ai_summary      TEXT NOT NULL DEFAULT ''
);
CREATE TABLE IF NOT EXISTS ai_sessions_raw (
  id           INTEGER PRIMARY KEY AUTOINCREMENT,
  tool         TEXT NOT NULL,
  project      TEXT NOT NULL,
  session_id   TEXT NOT NULL,
  ts           DATETIME NOT NULL,
  tokens_read  INTEGER NOT NULL,
  tokens_write INTEGER NOT NULL,
  tokens_cache INTEGER NOT NULL,
  cost         REAL NOT NULL,
  tools_called INTEGER NOT NULL,
  files_changed INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS ai_sessions_watermark (
  tool   TEXT PRIMARY KEY,
  offset INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_blocks_start ON blocks(start_ts);
`

type DB struct{ *sql.DB }

// GetState reads a value from the key-value state table.
// Returns "" if the key does not exist.
func (d *DB) GetState(key string) string {
	var v string
	d.QueryRow(`SELECT value FROM state WHERE key=?`, key).Scan(&v)
	return v
}

// SetState upserts a value in the key-value state table.
func (d *DB) SetState(key, value string) error {
	_, err := d.Exec(`INSERT INTO state(key,value) VALUES(?,?) ON CONFLICT(key) DO UPDATE SET value=excluded.value`, key, value)
	return err
}

func OpenDB(path string) (*DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}
	conn, err := sql.Open("sqlite3", path+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, err
	}
	conn.SetMaxOpenConns(1)
	if _, err := conn.Exec(schema); err != nil {
		conn.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	// migrate older DBs: add columns if missing (errors ignored if already present).
	_, _ = conn.Exec(`ALTER TABLE blocks ADD COLUMN data TEXT NOT NULL DEFAULT ''`)
	_, _ = conn.Exec(`ALTER TABLE blocks ADD COLUMN ai_summary TEXT NOT NULL DEFAULT ''`)
	// remove duplicate blocks keeping the latest id per start_ts
	_, _ = conn.Exec(`DELETE FROM blocks WHERE id NOT IN (
		SELECT MAX(id) FROM blocks GROUP BY start_ts
	)`)
	return &DB{conn}, nil
}
