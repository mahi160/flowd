package initwizard

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mahi/flowd/internal/config"
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
	defStr := "n"
	if def {
		defStr = "y"
	}
	ans := ask(r, prompt+" (y/n)", defStr)
	return strings.ToLower(ans) == "y"
}

func askInt(r *bufio.Reader, prompt string, def int) int {
	ans := ask(r, prompt, strconv.Itoa(def))
	n, err := strconv.Atoi(ans)
	if err != nil {
		return def
	}
	return n
}

// Run walks the user through config questions and returns a populated Config.
func Run() (*config.Config, error) {
	cfg := config.Defaults()
	r := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Println("  flowd init — answer questions to create your config")
	fmt.Println("  Press Enter to accept [default].")
	fmt.Println()

	fmt.Println("  ── Database ──────────────────────────────────────")
	cfg.DBPath = ask(r, "DB path", cfg.DBPath)

	fmt.Println()
	fmt.Println("  ── Polling ───────────────────────────────────────")
	cfg.PollIntervalSec = askInt(r, "tmux poll interval (seconds)", cfg.PollIntervalSec)
	cfg.SummaryIntervalMin = askInt(r, "summary block interval (minutes)", cfg.SummaryIntervalMin)

	fmt.Println()
	fmt.Println("  ── Keys ──────────────────────────────────────────")
	cfg.TrackKeys = askBool(r, "track key counts (aggregated, no raw keys)", cfg.TrackKeys)

	fmt.Println()
	fmt.Println("  ── Git sync ──────────────────────────────────────")
	cfg.PushDB = askBool(r, "sync logs to a private git repo", cfg.PushDB)
	if cfg.PushDB {
		cfg.RepoPath = ask(r, "local repo path", cfg.RepoPath)
		cfg.Branch = ask(r, "branch", cfg.Branch)
		fmt.Println("  Git remote URL (e.g. git@github.com:you/flowd-private.git)")
		fmt.Println("  Leave blank to skip — you can add it later with: git remote add origin <url>")
		cfg.GitRemote = ask(r, "remote URL", cfg.GitRemote)
	}

	fmt.Println()
	fmt.Println("  ── Project directories ───────────────────────────")
	fmt.Println("  Flowd tracks panes whose cwd is under these paths.")
	fmt.Println("  Comma-separated. Example: ~/code, ~/work, ~/personal")
	defaultDirs := strings.Join(cfg.WatchDirs, ", ")
	rawDirs := ask(r, "watch dirs", defaultDirs)
	cfg.WatchDirs = splitDirs(rawDirs)

	fmt.Println()
	fmt.Println("  ── AI summaries ──────────────────────────────────")
	fmt.Println("  Leave blank to skip. Examples: claude prompt / gemini chat")
	cfg.AICommand = ask(r, "AI command", cfg.AICommand)

	return cfg, nil
}

// SetupRepo initialises the local repo directory and wires the remote if provided.
// Safe to call even if the repo already exists.
func SetupRepo(repoPath, remote, branch string) error {
	repoPath = expandHome(repoPath)

	if err := os.MkdirAll(repoPath, 0750); err != nil {
		return fmt.Errorf("create repo dir: %w", err)
	}

	gitDir := filepath.Join(repoPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		if out, err := gitIn(repoPath, "init", "-b", branch); err != nil {
			return fmt.Errorf("git init: %w\n%s", err, out)
		}
		fmt.Printf("  git init → %s\n", repoPath)
	}

	if remote == "" {
		return nil
	}

	// check if origin already set
	out, getErr := gitIn(repoPath, "remote", "get-url", "origin")
	existing := strings.TrimSpace(out)
	if getErr != nil {
		// origin does not exist yet
		if out, err := gitIn(repoPath, "remote", "add", "origin", remote); err != nil {
			return fmt.Errorf("git remote add: %w\n%s", err, out)
		}
		fmt.Printf("  remote origin → %s\n", remote)
	} else if existing != remote {
		if out, err := gitIn(repoPath, "remote", "set-url", "origin", remote); err != nil {
			return fmt.Errorf("git remote set-url: %w\n%s", err, out)
		}
		fmt.Printf("  remote origin updated → %s\n", remote)
	} else {
		fmt.Printf("  remote origin already set → %s\n", existing)
	}

	return nil
}

func splitDirs(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		d := strings.TrimSpace(part)
		if d != "" {
			out = append(out, d)
		}
	}
	return out
}

func gitIn(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func expandHome(path string) string {
	if len(path) >= 2 && path[:2] == "~/" {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}
