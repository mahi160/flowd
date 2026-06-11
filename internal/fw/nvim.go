package fw

import (
	_ "embed"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// nvimStateTTL is the maximum age of a plugin state file before we consider
// it stale. nvim writes on every BufEnter (≤ a few seconds), so 15 s gives
// ample headroom while still catching a stale file from a crashed session.
const nvimStateTTL = 15 * time.Second

// nvimState is the JSON payload written by plugin/flowd.lua.
type nvimState struct {
	PID      int    `json:"pid"`
	Cwd      string `json:"cwd"`
	File     string `json:"file"`
	Filetype string `json:"filetype"`
	Project  string `json:"project"`
	Ts       int64  `json:"ts"`
}

// nvimStateDir returns the directory where the plugin writes state files:
// $XDG_DATA_HOME/flowd/nvim  (defaults to ~/.local/share/flowd/nvim).
func nvimStateDir() string {
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		home, _ := os.UserHomeDir()
		dataHome = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(dataHome, "flowd", "nvim")
}

// nvimConfigDir returns the neovim user config directory:
// $XDG_CONFIG_HOME/nvim  (defaults to ~/.config/nvim).
func nvimConfigDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "nvim")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "nvim")
}

// ReadNvimState scans the nvim state directory for a fresh state file whose
// cwd matches the given directory. We match by cwd (not by PID) because
// tmux's #{pane_pid} is the shell's PID, which differs from nvim's own PID
// (what the plugin uses to name the file).
//
// Returns nil (not an error) when:
//   - the plugin is not installed (no files in the dir),
//   - no file matches the cwd, or
//   - the matching file is stale.
//
// The caller should treat a nil return as "fall back to git-diff inference."
func ReadNvimState(cwd string) *nvimState {
	if cwd == "" {
		return nil
	}
	dir := nvimStateDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil // plugin absent or dir gone — not an error
	}
	now := time.Now()
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		var s nvimState
		if json.Unmarshal(data, &s) != nil {
			continue
		}
		if s.Cwd != cwd {
			continue
		}
		age := now.Sub(time.Unix(s.Ts, 0))
		if age > nvimStateTTL || age < -5*time.Second {
			continue // stale — nvim crashed without VimLeave, or clock skew
		}
		return &s
	}
	return nil
}

// NvimFiletype returns the filetype for a pane running nvim whose cwd is the
// given directory, or "" when the plugin is absent, stale, or not matching.
func NvimFiletype(cwd string) string {
	s := ReadNvimState(cwd)
	if s == nil {
		return ""
	}
	return s.Filetype
}

// ── Plugin installation ───────────────────────────────────────────────────────

// nvimPluginSource is the content of plugin/flowd.lua, embedded at build time
// so the binary is self-contained and can install the plugin on any machine.
// The file at lua/flowd.lua is a synced copy of plugin/flowd.lua — see the
// `sync-plugin` target in the Makefile.
//
//go:embed lua/flowd.lua
var nvimPluginSource string

// InstallNvimPlugin writes the bundled flowd.lua to
// ~/.config/nvim/plugin/flowd.lua (creating the directory if needed).
// Returns the path it was written to.
func InstallNvimPlugin() (string, error) {
	pluginDir := filepath.Join(nvimConfigDir(), "plugin")
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return "", err
	}
	dest := filepath.Join(pluginDir, "flowd.lua")
	if err := os.WriteFile(dest, []byte(nvimPluginSource), 0644); err != nil {
		return "", err
	}
	return dest, nil
}

// NvimPluginInstalled reports whether flowd.lua is present in the neovim
// plugin directory (quick check — doesn't verify version or contents).
func NvimPluginInstalled() bool {
	dest := filepath.Join(nvimConfigDir(), "plugin", "flowd.lua")
	_, err := os.Stat(dest)
	return err == nil
}
