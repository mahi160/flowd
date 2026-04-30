package db

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
  key_count       INTEGER NOT NULL DEFAULT 0,
  switches        INTEGER NOT NULL DEFAULT 0,
  tools           TEXT NOT NULL DEFAULT '',
  summary         TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_blocks_start ON blocks(start_ts);

CREATE TABLE IF NOT EXISTS schema_version (
  version INTEGER PRIMARY KEY
);
`

func (d *DB) migrate() error {
	_, err := d.Exec(schema)
	return err
}
