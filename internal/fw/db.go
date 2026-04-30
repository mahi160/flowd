package fw

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

const schema = `
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
  start_ts        DATETIME NOT NULL,
  end_ts          DATETIME NOT NULL,
  project         TEXT NOT NULL DEFAULT '',
  repo            TEXT NOT NULL DEFAULT '',
  focused_minutes INTEGER NOT NULL DEFAULT 0,
  switches        INTEGER NOT NULL DEFAULT 0,
  data            TEXT NOT NULL DEFAULT '',
  summary         TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_blocks_start ON blocks(start_ts);
`

type DB struct{ *sql.DB }

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
	return &DB{conn}, nil
}
