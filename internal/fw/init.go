package fw

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

// gitConfigGlobal reads a single git global config value.
func gitConfigGlobal(key string) string {
	out, err := exec.Command("git", "config", "--global", key).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// RunInitWizard walks the user through the minimal setup questions.
func RunInitWizard() (*Config, error) {
	cfg := DefaultConfig()
	r := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Println("  flowd init — Press Enter to accept [default].")

	gitName := gitConfigGlobal("user.name")
	gitEmail := gitConfigGlobal("user.email")
	fmt.Println()
	fmt.Println("  ── Git identity ──")
	fmt.Println("  Each coding block creates a git commit. For it to appear as a")
	fmt.Println("  green square on your GitHub profile, the email below must match")
	fmt.Println("  a verified address on your GitHub account.")
	fmt.Println()
	if gitName == "" && gitEmail == "" {
		fmt.Println("  ⚠  No global git identity found. Set one with:")
		fmt.Println("       git config --global user.name \"Your Name\"")
		fmt.Println("       git config --global user.email \"you@example.com\"")
		fmt.Println()
	} else {
		fmt.Printf("  name:  %s\n", gitName)
		fmt.Printf("  email: %s\n", gitEmail)
		fmt.Println()
		if !askBool(r, "Is this email verified on your GitHub account?", true) {
			fmt.Println()
			fmt.Println("  Update it with:")
			fmt.Println("    git config --global user.email \"you@example.com\"")
			fmt.Println("  Then re-run fw init.")
			fmt.Println()
			return nil, fmt.Errorf("aborted: fix git identity first")
		}
	}

	fmt.Println()
	fmt.Println("  ── Journal repo ──")
	fmt.Println("  Private git repo where flowd stores your session logs.")
	fmt.Println("  Each commit = one focus block = one green square on GitHub.")
	fmt.Println("  Add a remote to sync; leave blank for local only.")
	fmt.Println()
	cfg.GitRemote = ask(r, "git remote URL (blank = local only)", cfg.GitRemote)
	cfg.RepoPath = ask(r, "local repo path", cfg.RepoPath)

	fmt.Println()
	fmt.Println("  ── Machine name ──")
	fmt.Printf("  Used as the folder inside the repo: %s/<machine>/2026-04.md\n", cfg.RepoPath)
	fmt.Println()
	cfg.MachineName = ask(r, "machine name", cfg.MachineName)

	return cfg, nil
}

// SetupRepo prepares the journal repo. If a remote is given and the path
// doesn't exist yet, it tries git clone. Otherwise git init + remote add.
func SetupRepo(repoPath, remote, branch string) error {
	repoPath = expandHome(repoPath)
	ctx := context.Background()

	alreadyRepo := false
	if _, err := os.Stat(filepath.Join(repoPath, ".git")); err == nil {
		alreadyRepo = true
	}

	if !alreadyRepo {
		if remote != "" {
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
			if out, err := runGit(ctx, repoPath, "init", "-b", branch); err != nil {
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
	out, getErr := runGit(ctx, repoPath, "remote", "get-url", "origin")
	existing := strings.TrimSpace(out)
	if getErr != nil {
		if out, err := runGit(ctx, repoPath, "remote", "add", "origin", remote); err != nil {
			return fmt.Errorf("git remote add: %w\n%s", err, out)
		}
		fmt.Printf("  remote origin → %s\n", remote)
	} else if existing != remote {
		if out, err := runGit(ctx, repoPath, "remote", "set-url", "origin", remote); err != nil {
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

const gitignoreBlock = "# flowd — SQLite work files\nflowd.db\nflowd.db-wal\nflowd.db-shm\n"

func writeGitignore(repoPath string) error {
	path := filepath.Join(repoPath, ".gitignore")
	existing, _ := os.ReadFile(path)
	body := string(existing)

	// Check all three entries are present; if any is missing, append the block.
	needsUpdate := !strings.Contains(body, "flowd.db\n") ||
		!strings.Contains(body, "flowd.db-wal") ||
		!strings.Contains(body, "flowd.db-shm")
	if !needsUpdate {
		return nil
	}

	// Remove any partial flowd entries to avoid duplicates, then append fresh block.
	var cleaned []string
	for _, line := range strings.Split(body, "\n") {
		if strings.HasPrefix(line, "flowd.db") ||
			strings.TrimSpace(line) == "# flowd — SQLite work files" {
			continue
		}
		cleaned = append(cleaned, line)
	}
	merged := strings.TrimRight(strings.Join(cleaned, "\n"), "\n")
	if merged != "" {
		merged += "\n"
	}
	merged += gitignoreBlock
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
