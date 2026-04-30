package fw

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	PollIntervalSec  int      `yaml:"poll_interval_sec"`
	FocusBlockMin    int      `yaml:"focus_block_min"`    // focused minutes that make one block
	IdleThresholdSec int      `yaml:"idle_threshold_sec"`
	RepoPath         string   `yaml:"repo_path"`
	GitRemote        string   `yaml:"git_remote"` // blank = local-only (no push)
	Branch           string   `yaml:"branch"`
	WatchDirs        []string `yaml:"watch_dirs"`

	MachineName string `yaml:"machine_name"` // subfolder name inside repo (default: hostname)

	AIEnabled bool   `yaml:"ai_enabled"`
	AICommand string `yaml:"ai_command"` // any CLI reading stdin → stdout. Run via `sh -c`.
	AIPrompt  string `yaml:"ai_prompt"`  // prepended to the block summary on stdin.
}

// DBPath derives the SQLite path from RepoPath and MachineName.
// The DB always lives at <repo>/<machine>/flowd.db.
func (c *Config) DBPath() string {
	return filepath.Join(expandHome(c.RepoPath), c.MachineName, "flowd.db")
}

func DefaultConfig() *Config {
	home, _ := os.UserHomeDir()
	hostname, _ := os.Hostname()
	machine := strings.SplitN(hostname, ".", 2)[0]
	return &Config{
		PollIntervalSec:  3,
		FocusBlockMin:    30,
		IdleThresholdSec: 120,
		RepoPath:         filepath.Join(home, "flowd-private"),
		Branch:           "main",
		MachineName:      machine,
		WatchDirs:        []string{home},
		AIEnabled:        true,
		AICommand:        "pi -p --model haiku",
		AIPrompt:         "Summarize this coding session (30 focused minutes) in 2 short sentences. Focus on what was accomplished and any patterns. Be concise.",
	}
}

func DefaultConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "flowd", "config.yaml")
}

func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	cfg.RepoPath = expandHome(cfg.RepoPath)
	for i, d := range cfg.WatchDirs {
		cfg.WatchDirs[i] = expandHome(d)
	}
	// back-compat: if MachineName was not set, default is already hostname from DefaultConfig.
	if cfg.MachineName == "" {
		cfg.MachineName = DefaultConfig().MachineName
	}
	return cfg, nil
}

func WriteConfig(path string, cfg *Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0640)
}

func expandHome(p string) string {
	if len(p) >= 2 && p[:2] == "~/" {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, p[2:])
	}
	return p
}
