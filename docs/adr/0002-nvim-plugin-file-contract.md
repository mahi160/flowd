# neovim integration via plugin-written JSON file

To get per-buffer language data from neovim (filetype before a git commit
lands), flowd uses a small Lua plugin (`plugin/flowd.lua`) that writes a
JSON state file to `~/.local/share/flowd/nvim/<pid>.json` on every
`BufEnter`/`DirChanged`.

The daemon reads that file when it polls a pane whose command is `nvim`.
If the file is absent or older than 15 s, it degrades gracefully to git-diff
inference — the plugin is optional.

**Considered alternative:** discover each neovim instance's RPC socket
(written to `$XDG_RUNTIME_DIR/nvim.<pid>/0` or similar) and query
`nvim --remote-expr 'expand("%") . "|" . &filetype'` without a plugin.
Rejected because: (a) socket discovery is fragile — the path varies between
platforms and neovim versions; (b) matching the right socket to the right
tmux pane when multiple nvim instances are open requires additional PID
correlation; (c) a `--remote-expr` call spawns a child process on every
3-second poll tick; (d) the file-based approach gives richer data (project
name, cwd, timestamp) with a single `ReadFile` per tick.

The file path and JSON schema (`pid`, `cwd`, `file`, `filetype`, `project`,
`ts`) are now a public contract between the plugin and the daemon. Breaking
changes require a version bump in both.
