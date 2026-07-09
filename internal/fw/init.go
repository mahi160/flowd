package fw

import (
	"bufio"
	"context"
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
	fmt.Println("  Create an empty private repo on GitHub first, then paste the URL.")
	fmt.Println("  Leave blank for local-only (no GitHub contribution graph).")
	fmt.Println()
	fmt.Println("  e.g. git@github.com:yourname/journal.git")
	fmt.Println()
	cfg.GitRemote = ask(r, "git remote URL (blank = local only)", "")
	cfg.RepoPath = ask(r, "local repo path", cfg.RepoPath)

	fmt.Println()
	fmt.Println("  ── Machine name ──")
	fmt.Printf("  Used as the folder inside the repo: %s/<machine>/2026-04.md\n", cfg.RepoPath)
	fmt.Println()
	cfg.MachineName = ask(r, "machine name", cfg.MachineName)

	// ── AI summaries ────────────────────────────────────────────────────────
	fmt.Println()
	fmt.Println("  ── AI summaries (optional) ──")
	fmt.Println("  flowd can pipe each focus block + your daily standup through an")
	fmt.Println("  AI CLI for auto-generated notes. Any stdin→stdout command works.")
	fmt.Println()

	if cmd, enabled := pickAICommand(r); enabled {
		cfg.AIEnabled = true
		cfg.AICommand = cmd
		fmt.Printf("  ai_command: %s\n", cmd)
	} else {
		cfg.AIEnabled = false
		cfg.AICommand = ""
		fmt.Println("  AI summaries disabled. Enable later in ~/.config/flowd/config.yaml.")
	}

	// ── neovim plugin ────────────────────────────────────────────────────────
	if _, err := exec.LookPath("nvim"); err == nil {
		fmt.Println()
		fmt.Println("  ── neovim plugin (optional) ──")
		fmt.Println("  Gives flowd your exact filetype on every buffer switch, before")
		fmt.Println("  any git commit lands. Works with or without a plugin manager.")
		fmt.Println()

		if NvimPluginInstalled() {
			fmt.Printf("  already installed: %s/plugin/flowd.lua\n", nvimConfigDir())
		} else if askBool(r, "install flowd.lua into ~/.config/nvim/plugin/", true) {
			if dest, err := InstallNvimPlugin(); err != nil {
				fmt.Printf("  warn: could not install plugin: %v\n", err)
			} else {
				fmt.Printf("  installed → %s\n", dest)
			}
		}
	}

	return cfg, nil
}

// aiPreset is a known AI CLI that flowd can auto-detect.
type aiPreset struct {
	name    string // display name
	binary  string // executable to look for in PATH
	command string // ai_command value to write to config
}

var aiPresets = []aiPreset{
	{name: "pi", binary: "pi", command: "pi -p --model haiku"},
	{name: "claude", binary: "claude", command: "claude --print"},
	{name: "gemini", binary: "gemini", command: "gemini -p --model gemini-2.0-flash-lite"},
	{name: "llm", binary: "llm", command: "llm -m gpt-4o-mini"},
	{name: "aider", binary: "aider", command: "aider --no-pretty"},
}

// pickAICommand auto-detects installed AI CLIs, shows a numbered list, and
// returns the chosen command + whether AI is enabled.
func pickAICommand(r *bufio.Reader) (cmd string, enabled bool) {
	// Find installed presets.
	var found []aiPreset
	for _, p := range aiPresets {
		if _, err := exec.LookPath(p.binary); err == nil {
			found = append(found, p)
		}
	}

	if len(found) == 0 {
		fmt.Println("  No supported AI CLI detected in PATH.")
		fmt.Println("  Supported: pi, claude, gemini, llm, aider (or any stdin→stdout tool).")
		fmt.Println()
		custom := ask(r, "ai_command (blank to skip)", "")
		if custom == "" {
			return "", false
		}
		return custom, true
	}

	fmt.Println("  Detected on your system:")
	for i, p := range found {
		fmt.Printf("    %d) %-10s →  %s\n", i+1, p.name, p.command)
	}
	fmt.Println("    0) skip — no AI")
	fmt.Println()

	defaultChoice := "1" // first detected tool
	choiceStr := ask(r, "choice", defaultChoice)
	n, err := strconv.Atoi(strings.TrimSpace(choiceStr))
	if err != nil {
		// Treat non-numeric input as a custom command.
		if strings.TrimSpace(choiceStr) != "" {
			return choiceStr, true
		}
		return "", false
	}
	if n == 0 || n > len(found) {
		return "", false
	}
	return found[n-1].command, true
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
				if err != nil {
					return fmt.Errorf("git clone failed: %s\n\nFix your remote URL or SSH key, then re-run fw init", strings.TrimSpace(string(out)))
				}
				fmt.Printf("  git clone → %s\n", repoPath)
				alreadyRepo = true
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
	return askBool(bufio.NewReader(os.Stdin), "start fw automatically when tmux starts", true)
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

	line := "\n# flowd — start activity tracker with tmux\nrun-shell \"fw start\"\n"
	if _, err := f.WriteString(line); err != nil {
		return err
	}
	fmt.Println("  added run-shell line to ~/.tmux.conf")
	_ = exec.Command("tmux", "source-file", conf).Run()
	return nil
}
