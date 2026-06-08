# neovim plugin

The `plugin/flowd.lua` plugin gives flowd accurate, real-time language
attribution for neovim panes — before any git commit lands.

Without the plugin, flowd infers the language from `git diff` of changed
files (which is good but lags until you commit) or by scanning file
extensions in the cwd. With the plugin, every buffer switch is immediately
visible to the daemon.

## How it works

The plugin writes a small JSON file on every `BufEnter`, `BufWritePost`, and
`DirChanged`:

```
~/.local/share/flowd/nvim/<pid>.json
```

The file contains:

```json
{
  "pid":      12345,
  "cwd":      "/home/user/myproject",
  "file":     "/home/user/myproject/main.go",
  "filetype": "go",
  "project":  "myproject",
  "ts":       1718000000
}
```

The daemon reads this file when it polls a pane running `nvim`. If the file
is fresh (written within the last 15 s), the filetype is used for language
attribution. If the file is absent or stale, the daemon falls back silently
to git-diff inference — **the plugin is fully optional**.

On `VimLeave`, the file is deleted so stale data from a crashed session
doesn't persist.

## Installation

### Automatic (recommended)

The plugin is bundled inside the `fw` binary. After installing flowd, run:

```sh
fw setup-nvim
```

This writes `flowd.lua` to `~/.config/nvim/plugin/flowd.lua` and no further
configuration is needed — neovim auto-sources all files in that directory.
You can also install it during `fw init` when neovim is detected.

To update the plugin after upgrading `fw`, re-run `fw setup-nvim`.

### lazy.nvim

If you prefer to manage it via lazy.nvim, point it at the installed file:

```lua
{
  dir    = vim.fn.stdpath("config") .. "/plugin",  -- after fw setup-nvim
  name   = "flowd",
  -- setup() is called automatically; this line is optional:
  config = function() require("flowd").setup() end,
}
```

### Manual (no plugin manager, no `fw setup-nvim`)

Copy `plugin/flowd.lua` from the flowd repo to `~/.config/nvim/plugin/flowd.lua`.
No `require()` call needed — the file auto-runs on source.

### Verification

After installing, open a file in nvim and run:

```sh
cat ~/.local/share/flowd/nvim/$(pgrep -n nvim).json
```

You should see a JSON object with the current buffer's filetype.

## Configuration

`setup()` accepts an options table (currently no options; reserved for future
use):

```lua
require("flowd").setup({})  -- same as require("flowd").setup()
```

## Supported filetypes → Languages mapping

The daemon maps neovim filetypes to canonical language names using
`langFromFiletype` in `internal/fw/git.go`. Currently mapped:

| nvim filetype | Language |
|---------------|----------|
| `go` | Go |
| `python` | Python |
| `javascript`, `javascriptreact` | JavaScript |
| `typescript`, `typescriptreact` | TypeScript |
| `rust` | Rust |
| `lua` | Lua |
| `ruby` | Ruby |
| `java` | Java |
| `kotlin` | Kotlin |
| `swift` | Swift |
| `c` | C |
| `cpp` | C++ |
| `cs` | C# |
| `php` | PHP |
| `elixir` | Elixir |
| `haskell` | Haskell |
| `scala` | Scala |
| `html` | HTML |
| `css`, `scss`, `sass`, `less` | CSS |
| `svelte` | Svelte |
| `vue` | Vue |
| `sh`, `bash`, `zsh` | Shell |
| `sql` | SQL |
| `markdown`, `mdx` | Markdown |
| `yaml`, `toml`, `json`, `jsonc` | Config |
| `dockerfile` | Dockerfile |

Unmapped filetypes are silently ignored (the git-diff path still runs).
