# Dashboard as a static single-file embed

`fw dashboard` builds a self-contained HTML file (the compiled Svelte app
with all data baked in as a `window.__FLOWD_DATA__` JSON blob) and opens it
in the browser. The file is written to `$TMPDIR/flowd.html` and is a
snapshot of that moment's DB state.

The data for all six periods (today, yesterday, week, month, year, all) is
embedded in one payload so every tab works immediately without network
round-trips.

**Considered alternative:** run an HTTP server from the daemon (e.g.
`fw dashboard --serve :8080`) with lazy per-period JSON endpoints and
auto-refresh via SSE or polling.

Rejected because:

1. **Port management complexity.** A long-running server needs port
   allocation, conflict detection, and a way to stop it. The daemon already
   manages one background process; a second server adds operational surface.
2. **"Live" adds little value here.** The underlying data only changes every
   30 minutes when a new block lands. Refreshing a snapshot is no worse than
   a live feed for this cadence.
3. **Local-first design.** A static file works offline, opens with `file://`,
   can be shared or archived, and requires no firewall rules.
4. **Vite `vite-plugin-singlefile` already exists in the build.** The
   pipeline (Svelte → single HTML → `go:embed`) was already working before
   this refactor. Changing it would be pure cost.
