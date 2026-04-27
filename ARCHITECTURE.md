# System Design

## Runtime Flow

1. `fw start`
2. daemon loads config
3. collectors gather tmux state
4. events written to SQLite
5. scheduler creates 30-min blocks
6. summarizer generates markdown log
7. sync pushes repo

## Components

```text
cmd/fw
internal/config
internal/db
internal/collector/tmux
internal/collector/process
internal/session
internal/summarizer
internal/provider
internal/sync
internal/report
```

## Data Model

```sql
CREATE TABLE events (
  id INTEGER PRIMARY KEY,
  ts DATETIME,
  type TEXT,
  value TEXT,
  meta TEXT
);

CREATE TABLE blocks (
  id INTEGER PRIMARY KEY,
  start_ts DATETIME,
  end_ts DATETIME,
  project TEXT,
  repo TEXT,
  focused_minutes INTEGER,
  key_count INTEGER,
  switches INTEGER,
  tools TEXT,
  summary TEXT
);
```
