# Rules for Coding Agents

## Read First

Read PRD.md and TASKS.md before coding.

## Tech Constraints

- Language: Go
- DB: SQLite
- CLI: Cobra or stdlib
- OS target first: Linux/macOS
- Minimal dependencies

## Architecture Rules

- Keep modules small and replaceable
- Prefer interfaces where needed, not everywhere
- No premature abstractions
- Keep state explicit
- Fail gracefully if tmux unavailable

## Performance Rules

- Prefer tmux hooks over polling
- Poll fallback every 3-5 sec max
- Batch DB writes
- Summaries every 30 min only

## Privacy Rules

- Never store raw keys by default
- Never send data externally
- AI calls only through user-configured command

## Delivery Rules

- Build milestone by milestone
- After each milestone: test, document, commit
- Avoid speculative features
