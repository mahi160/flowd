package fw

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

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
CREATE INDEX IF NOT EXISTS idx_blocks_start ON blocks(start_ts);

CREATE TABLE IF NOT EXISTS ai_sessions_raw (
  id            INTEGER PRIMARY KEY AUTOINCREMENT,
  tool          TEXT NOT NULL,
  project       TEXT NOT NULL,
  session_id    TEXT NOT NULL,
  model         TEXT NOT NULL DEFAULT '',
  ts            DATETIME NOT NULL,
  tokens_read   INTEGER NOT NULL,
  tokens_write  INTEGER NOT NULL,
  tokens_cache  INTEGER NOT NULL,
  cost          REAL NOT NULL,
  tools_called  INTEGER NOT NULL,
  files_changed INTEGER NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_ai_sessions_unique
  ON ai_sessions_raw(tool, session_id, ts);

CREATE TABLE IF NOT EXISTS ai_sessions_watermark (
  tool   TEXT PRIMARY KEY,
  offset INTEGER NOT NULL DEFAULT 0
);
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
	_, err := d.Exec(`INSERT INTO state(key,value) VALUES(?,?)
	  ON CONFLICT(key) DO UPDATE SET value=excluded.value`, key, value)
	return err
}

// PruneEvents deletes raw events older than keepDays. Blocks are already
// persisted as summaries so old events are just disk overhead.
func (d *DB) PruneEvents(ctx context.Context, keepDays int) error {
	cutoff := time.Now().AddDate(0, 0, -keepDays).UTC()
	res, err := d.ExecContext(ctx,
		`DELETE FROM events WHERE ts < ?`, cutoff)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n > 0 {
		_, _ = d.ExecContext(ctx, "PRAGMA wal_checkpoint(TRUNCATE)")
	}
	return nil
}

// QueryStreak returns the current consecutive-day streak by checking all
// historical blocks in the DB, not just a single period's worth.
func (d *DB) QueryStreak(ctx context.Context) int {
	rows, err := d.QueryContext(ctx,
		`SELECT date(start_ts, 'localtime') as day
		 FROM blocks
		 WHERE focused_minutes > 0
		 GROUP BY day
		 ORDER BY day DESC`)
	if err != nil {
		return 0
	}
	defer rows.Close()

	streak := 0
	expected := time.Now().Local().Format("2006-01-02")
	for rows.Next() {
		var day string
		if err := rows.Scan(&day); err != nil {
			break
		}
		if day != expected {
			// Allow today to be missing (the day isn't over yet);
			// but only for the very first iteration.
			if streak == 0 {
				yesterday := time.Now().Local().AddDate(0, 0, -1).Format("2006-01-02")
				if day != yesterday {
					break
				}
				// expected will be set correctly by the assignment below; no need to set here.
			} else {
				break
			}
		}
		streak++
		t, _ := time.Parse("2006-01-02", day)
		expected = t.AddDate(0, 0, -1).Format("2006-01-02")
	}
	return streak
}

// QueryMonthStats returns block count, active day count, and total focus
// minutes for the current calendar month — used to generate the README.
func (d *DB) QueryMonthStats(ctx context.Context) (blocks, days, focusMin int) {
	now := time.Now().Local() // use local time so month boundary matches the user's calendar
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).UTC()
	end := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location()).UTC()

	d.QueryRowContext(ctx,
		`SELECT COUNT(*), COALESCE(SUM(focused_minutes), 0) FROM blocks
		 WHERE start_ts >= ? AND start_ts < ?`,
		start, end).Scan(&blocks, &focusMin)

	d.QueryRowContext(ctx,
		`SELECT COUNT(DISTINCT date(start_ts)) FROM blocks
		 WHERE start_ts >= ? AND start_ts < ?`,
		start, end).Scan(&days)
	return
}

func OpenDB(path string) (*DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}
	// _loc=UTC ensures time.Time values are scanned in UTC.
	dsn := path + "?_journal_mode=WAL&_busy_timeout=5000&_loc=UTC"
	conn, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	conn.SetMaxOpenConns(1)
	if _, err := conn.Exec(schema); err != nil {
		conn.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	// ── Migrations for older DBs ──────────────────────────────────────
	// Errors from ALTER TABLE are expected when the column already exists.
	_, _ = conn.Exec(`ALTER TABLE blocks ADD COLUMN data TEXT NOT NULL DEFAULT ''`)
	_, _ = conn.Exec(`ALTER TABLE blocks ADD COLUMN ai_summary TEXT NOT NULL DEFAULT ''`)
	_, _ = conn.Exec(`ALTER TABLE ai_sessions_raw ADD COLUMN model TEXT NOT NULL DEFAULT ''`)

	// Dedup blocks — keep latest id per start_ts.
	_, _ = conn.Exec(`DELETE FROM blocks WHERE id NOT IN (
		SELECT MAX(id) FROM blocks GROUP BY start_ts
	)`)

	// Add unique index if it didn't exist yet (schema CREATE already handles
	// new DBs; this covers existing DBs that predate the index).
	_, _ = conn.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_ai_sessions_unique
	  ON ai_sessions_raw(tool, session_id, ts)`)

	return &DB{conn}, nil
}
