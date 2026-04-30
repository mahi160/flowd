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

	fmt.Println("  ── Journal repo ──")
	fmt.Println("  Private repo where flowd stores summaries AND the SQLite DB.")
	fmt.Println("  Provide a remote URL to keep it synced; leave blank for local-only.")
	cfg.GitRemote = ask(r, "private repo remote URL (blank = local only)", cfg.GitRemote)
	cfg.RepoPath = ask(r, "local repo path", cfg.RepoPath)
	cfg.Branch = ask(r, "branch", cfg.Branch)
	cfg.DBPath = filepath.Join(expandHome(cfg.RepoPath), "flowd.db")

	fmt.Println()
	fmt.Println("  ── Polling ──")
	cfg.PollIntervalSec = askInt(r, "tmux poll interval (seconds)", cfg.PollIntervalSec)
	cfg.SummaryIntervalMin = askInt(r, "summary block interval (minutes)", cfg.SummaryIntervalMin)
	cfg.MinFocusMin = askInt(r, "min focused minutes per block to push", cfg.MinFocusMin)
	cfg.IdleThresholdSec = askInt(r, "idle threshold (sec) — pause tracking after no input", cfg.IdleThresholdSec)

	fmt.Println()
	fmt.Println("  ── AI summaries (optional) ──")
	fmt.Println("  Pipe each block's summary through any CLI AI tool (claude, codex,")
	fmt.Println("  pi, opencode, llm, etc). The tool must read stdin and print to stdout.")
	cfg.AIEnabled = askBool(r, "enable AI summaries", cfg.AIEnabled)
	if cfg.AIEnabled {
		fmt.Println("  Examples: 'claude --print', 'codex exec', 'llm -m claude-3-5-sonnet'")
		cfg.AICommand = ask(r, "AI command", cfg.AICommand)
		cfg.AIPrompt = ask(r, "AI prompt", cfg.AIPrompt)
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

// SetupRepo prepares the journal repo. If a remote is given and the path
// doesn't exist yet, it tries `git clone`. Otherwise it `git init`s in
// place and adds the remote (if any). Always writes a .gitignore that
// excludes SQLite work files.
func SetupRepo(repoPath, remote, branch string) error {
	repoPath = expandHome(repoPath)

	alreadyRepo := false
	if _, err := os.Stat(filepath.Join(repoPath, ".git")); err == nil {
		alreadyRepo = true
	}

	if !alreadyRepo {
		if remote != "" {
			// Try clone. Falls back to init on failure (e.g., empty remote).
			if _, err := os.Stat(repoPath); os.IsNotExist(err) || isEmptyDir(repoPath) {
				out, err := exec.Command("git", "clone", "-b", branch, remote, repoPath).CombinedOutput()
				if err == nil {
					fmt.Printf("  git clone → %s\n", repoPath)
					alreadyRepo = true
				} else {
					fmt.Printf("  clone failed (%s) — initializing fresh repo\n", strings.TrimSpace(string(out)))
				}
			}
		}
		if !alreadyRepo {
			if err := os.MkdirAll(repoPath, 0750); err != nil {
				return err
			}
			if out, err := gitIn(repoPath, "init", "-b", branch); err != nil {
				return fmt.Errorf("git init: %w\n%s", err, out)
			}
			fmt.Printf("  git init → %s\n", repoPath)
		}
	}

	if err := writeGitignore(repoPath); err != nil {
		return err
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

func isEmptyDir(p string) bool {
	entries, err := os.ReadDir(p)
	return err == nil && len(entries) == 0
}

func writeGitignore(repoPath string) error {
	const body = "# flowd — SQLite work files\nflowd.db-wal\nflowd.db-shm\n"
	path := filepath.Join(repoPath, ".gitignore")
	existing, _ := os.ReadFile(path)
	if strings.Contains(string(existing), "flowd.db-wal") {
		return nil
	}
	merged := string(existing)
	if merged != "" && !strings.HasSuffix(merged, "\n") {
		merged += "\n"
	}
	merged += body
	return os.WriteFile(path, []byte(merged), 0640)
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
