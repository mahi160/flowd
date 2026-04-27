package fw

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func ask(r *bufio.Reader, prompt, def string) string {
	if def != "" {
		fmt.Printf("  %s [%s]: ", prompt, def)
	} else {
		fmt.Printf("  %s: ", prompt)
	}
	line, _ := r.ReadString('\n')
	line = strings.TrimSpace(line)
	if line == "" {
		return def
	}
	return line
}

func askBool(r *bufio.Reader, prompt string, def bool) bool {
	d := "n"
	if def {
		d = "y"
	}
	return strings.ToLower(ask(r, prompt+" (y/n)", d)) == "y"
}

func askInt(r *bufio.Reader, prompt string, def int) int {
	n, err := strconv.Atoi(ask(r, prompt, strconv.Itoa(def)))
	if err != nil {
		return def
	}
	return n
}

// RunInitWizard walks the user through config questions.
func RunInitWizard() (*Config, error) {
	cfg := DefaultConfig()
	r := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Println("  flowd init — Press Enter to accept [default].")
	fmt.Println()

	cfg.DBPath = ask(r, "DB path", cfg.DBPath)
	cfg.PollIntervalSec = askInt(r, "tmux poll interval (seconds)", cfg.PollIntervalSec)
	cfg.SummaryIntervalMin = askInt(r, "summary block interval (minutes)", cfg.SummaryIntervalMin)
	cfg.MinFocusMin = askInt(r, "min focused minutes per block to push", cfg.MinFocusMin)

	fmt.Println()
	fmt.Println("  ── Journal repo ──")
	fmt.Println("  Private repo where flowd writes summaries (NOT your project repo).")
	cfg.PushDB = askBool(r, "sync summaries to a private journal repo", cfg.PushDB)
	if cfg.PushDB {
		cfg.RepoPath = ask(r, "local journal repo path", cfg.RepoPath)
		cfg.Branch = ask(r, "branch", cfg.Branch)
		cfg.GitRemote = ask(r, "remote URL (blank to skip)", cfg.GitRemote)
	}

	fmt.Println()
	fmt.Println("  ── Watch dirs ──")
	fmt.Println("  Comma-separated. Only panes whose cwd is under these paths are tracked.")
	cfg.WatchDirs = splitCSV(ask(r, "watch dirs", strings.Join(cfg.WatchDirs, ", ")))

	return cfg, nil
}

func splitCSV(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func SetupRepo(repoPath, remote, branch string) error {
	repoPath = expandHome(repoPath)
	if err := os.MkdirAll(repoPath, 0750); err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(repoPath, ".git")); os.IsNotExist(err) {
		if out, err := gitIn(repoPath, "init", "-b", branch); err != nil {
			return fmt.Errorf("git init: %w\n%s", err, out)
		}
		fmt.Printf("  git init → %s\n", repoPath)
	}
	if remote == "" {
		return nil
	}
	out, getErr := gitIn(repoPath, "remote", "get-url", "origin")
	existing := strings.TrimSpace(out)
	if getErr != nil {
		if out, err := gitIn(repoPath, "remote", "add", "origin", remote); err != nil {
			return fmt.Errorf("git remote add: %w\n%s", err, out)
		}
		fmt.Printf("  remote origin → %s\n", remote)
	} else if existing != remote {
		if out, err := gitIn(repoPath, "remote", "set-url", "origin", remote); err != nil {
			return fmt.Errorf("git remote set-url: %w\n%s", err, out)
		}
		fmt.Printf("  remote origin updated → %s\n", remote)
	}
	return nil
}

func AskTmuxAutostart() bool {
	return askBool(bufio.NewReader(os.Stdin), "start fw automatically when tmux starts", false)
}

func SetupTmuxAutostart() error {
	home, _ := os.UserHomeDir()
	conf := filepath.Join(home, ".tmux.conf")
	const marker = "fw start"

	existing, err := os.ReadFile(conf)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if strings.Contains(string(existing), marker) {
		fmt.Println("  tmux autostart already configured")
		return nil
	}

	f, err := os.OpenFile(conf, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	line := "\n# flowd — start activity tracker with tmux\nrun-shell \"fw start &> /tmp/flowd.log &\"\n"
	if _, err := f.WriteString(line); err != nil {
		return err
	}
	fmt.Println("  added run-shell line to ~/.tmux.conf")
	_ = exec.Command("tmux", "source-file", conf).Run()
	return nil
}

func gitIn(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err
}
