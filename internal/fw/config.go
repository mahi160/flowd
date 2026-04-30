package fw

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	PollIntervalSec    int      `yaml:"poll_interval_sec"`
	SummaryIntervalMin int      `yaml:"summary_interval_min"`
	MinFocusMin        int      `yaml:"min_focus_min"`
	PushDB             bool     `yaml:"push_db"`
	RepoPath           string   `yaml:"repo_path"`
	GitRemote          string   `yaml:"git_remote"`
	Branch             string   `yaml:"branch"`
	WatchDirs          []string `yaml:"watch_dirs"`
	DBPath             string   `yaml:"db_path"`
}

func DefaultConfig() *Config {
	home, _ := os.UserHomeDir()
	return &Config{
		PollIntervalSec:    3,
		SummaryIntervalMin: 30,
		MinFocusMin:        15,
		PushDB:             false,
		RepoPath:           filepath.Join(home, "flowd-private"),
		Branch:             "main",
		DBPath:             filepath.Join(home, ".local", "share", "flowd", "flowd.db"),
		WatchDirs:          []string{home},
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
	cfg.DBPath = expandHome(cfg.DBPath)
	for i, d := range cfg.WatchDirs {
		cfg.WatchDirs[i] = expandHome(d)
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
