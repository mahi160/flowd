# Process-tree walk for AI CLI detection

When a tmux pane runs an AI coding CLI like `pi` or `claude-code`, the tool
runs inside a Node.js (or Python) interpreter. `tmux`'s
`#{pane_current_command}` returns the interpreter name (`node`, `python`),
not the actual tool — causing flowd to (a) attribute the time to `node`
instead of `pi` in `ByTool`, and (b) infer phantom JavaScript/Python
"language" minutes from a pane that never touched source code.

We decided to walk the OS process tree from the pane's PID whenever the
foreground command is a known interpreter (`node`, `python`, `bun`, `deno`,
`ruby`, …). We look for a descendant whose executable basename matches a
known AI CLI (`pi`, `claude`, `aider`, `codex`, `opencode`, `llm`, …).
When found, we override the pane's command label and set category `ai` before
writing the event. Results are cached per pane-PID with a 5 s TTL to keep
the common path cheap.

On Linux the tree is built from `/proc/<pid>/stat` + `/proc/<pid>/cmdline`;
on macOS from one `ps -axo pid=,ppid=,comm=` call.

**Considered alternative:** ignore the tmux label entirely for AI time and
derive it solely from the JSONL transcript timestamps (which are already
ingested by `ai_sessions`). Rejected because JSONL gives token/cost data
but not wall-clock pane-occupancy time, and `ByTool` would silently
under-count AI tool usage.
