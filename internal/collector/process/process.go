package process

import (
	"os/exec"
	"strings"
)

// RepoRoot returns the git repo root for the given path, or "" if none.
func RepoRoot(cwd string) string {
	out, err := exec.Command("git", "-C", cwd, "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// RepoName returns the basename of the git repo root.
func RepoName(cwd string) string {
	root := RepoRoot(cwd)
	if root == "" {
		return ""
	}
	parts := strings.Split(root, "/")
	return parts[len(parts)-1]
}

// ClassifyCommand maps a raw command name to a tool category.
func ClassifyCommand(cmd string) string {
	switch {
	case cmd == "nvim" || cmd == "vim" || cmd == "vi":
		return "editor"
	case cmd == "lazygit" || cmd == "git":
		return "git"
	case cmd == "claude" || cmd == "gemini" || cmd == "codex" || cmd == "aider":
		return "ai"
	case cmd == "zsh" || cmd == "bash" || cmd == "fish" || cmd == "sh":
		return "shell"
	case cmd == "node" || cmd == "python" || cmd == "python3" || cmd == "go" || cmd == "cargo":
		return "runtime"
	default:
		return "other"
	}
}
