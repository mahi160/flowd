package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	PollIntervalSec    int      `yaml:"poll_interval_sec"`
	SummaryIntervalMin int      `yaml:"summary_interval_min"`
	TrackKeys          bool     `yaml:"track_keys"`
	TrackRawKeys       bool     `yaml:"track_raw_keys"`
	PushDB             bool     `yaml:"push_db"`
	RepoPath           string   `yaml:"repo_path"`
	Branch             string   `yaml:"branch"`
	AICommand          string   `yaml:"ai_command"`
	ExcludePaths       []string `yaml:"exclude_paths"`
	DBPath             string   `yaml:"db_path"`
}

func Defaults() *Config {
	home, _ := os.UserHomeDir()
	return &Config{
		PollIntervalSec:    3,
		SummaryIntervalMin: 30,
		TrackKeys:          true,
		TrackRawKeys:       false,
		PushDB:             false,
		RepoPath:           filepath.Join(home, "flowd-private"),
		Branch:             "main",
		DBPath:             filepath.Join(home, ".local", "share", "flowd", "flowd.db"),
	}
}

func Load(path string) (*Config, error) {
	cfg := Defaults()

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

	// expand ~ in paths
	cfg.RepoPath = expandHome(cfg.RepoPath)
	cfg.DBPath = expandHome(cfg.DBPath)

	return cfg, nil
}

func DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "flowd", "config.yaml")
}

func Write(path string, cfg *Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	return os.WriteFile(path, data, 0640)
}

func expandHome(path string) string {
	if len(path) >= 2 && path[:2] == "~/" {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}
