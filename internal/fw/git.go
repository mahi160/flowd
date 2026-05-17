package fw

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var extLang = map[string]string{
	".go": "Go", ".rs": "Rust", ".py": "Python", ".js": "JavaScript",
	".ts": "TypeScript", ".tsx": "TypeScript", ".jsx": "JavaScript",
	".rb": "Ruby", ".java": "Java", ".kt": "Kotlin", ".swift": "Swift",
	".c": "C", ".h": "C", ".cpp": "C++", ".cc": "C++", ".hpp": "C++",
	".cs": "C#", ".php": "PHP", ".sh": "Shell", ".zsh": "Shell", ".bash": "Shell",
	".lua": "Lua", ".vim": "Vim", ".md": "Markdown", ".yaml": "YAML", ".yml": "YAML",
	".toml": "TOML", ".json": "JSON", ".html": "HTML", ".css": "CSS", ".scss": "SCSS",
	".sql": "SQL", ".tf": "Terraform", ".svelte": "Svelte",
	".vue": "Vue", ".dart": "Dart", ".ex": "Elixir", ".exs": "Elixir", ".erl": "Erlang",
	".scala": "Scala", ".clj": "Clojure", ".hs": "Haskell", ".ml": "OCaml",
}

// knownLangs is the set of language names we consider meaningful.
// Derived from extLang values. Anything not in this set is noise.
var knownLangs = func() map[string]bool {
	m := map[string]bool{}
	for _, v := range extLang {
		m[v] = true
	}
	return m
}()

// isKnownLang reports whether a language name is one we recognise (not a noise
// extension like "lock", "svg", "db", etc.).
func isKnownLang(lang string) bool {
	return knownLangs[lang]
}

func langOf(ext string) string {
	if l, ok := extLang[ext]; ok {
		return l
	}
	if ext == "" {
		return "other"
	}
	return strings.TrimPrefix(ext, ".")
}

// LangFromCommand maps a runtime process name to a language.
func LangFromCommand(cmd string) string {
	switch cmd {
	case "node", "bun":
		return "JavaScript"
	case "deno":
		return "TypeScript"
	case "python", "python3":
		return "Python"
	case "go":
		return "Go"
	case "cargo":
		return "Rust"
	case "ruby":
		return "Ruby"
	case "java":
		return "Java"
	case "php":
		return "PHP"
	case "elixir":
		return "Elixir"
	}
	return ""
}

// ScanLangs returns file-extension counts for source files in dir (maxdepth 2).
// Uses filepath.WalkDir — no external process, works on all platforms.
func ScanLangs(dir string) map[string]int {
	counts := map[string]int{}
	depth := strings.Count(filepath.Clean(dir), string(os.PathSeparator))

	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		// Skip hidden dirs and common noise dirs.
		name := d.Name()
		if d.IsDir() {
			if name == "node_modules" || name == "vendor" || name == ".git" ||
				(len(name) > 0 && name[0] == '.') {
				return filepath.SkipDir
			}
			// Limit depth to 2 below the root.
			cur := strings.Count(filepath.Clean(path), string(os.PathSeparator))
			if cur-depth > 2 {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(name))
		if ext == "" {
			return nil
		}
		lang := langOf(ext)
		if lang != "other" {
			counts[lang]++
		}
		return nil
	})
	return counts
}

// runGit runs a git command in dir, returns combined output and error.
// This is the single canonical git runner for the whole package.
func runGit(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// git runs a git command and returns only an error (discards stdout/stderr
// on success). Wraps runGit for callers that don't need the output.
func git(ctx context.Context, dir string, args ...string) error {
	out, err := runGit(ctx, dir, args...)
	if err != nil {
		return fmt.Errorf("git %v: %w\n%s", args, err, out)
	}
	return nil
}

// gitTimeout is the max time we'll wait for any single git metadata query.
// Prevents hangs on network filesystems or slow git hooks.
const gitTimeout = 5 * time.Second

func RepoRoot(cwd string) string {
	ctx, cancel := context.WithTimeout(context.Background(), gitTimeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, "git", "-C", cwd, "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func RepoName(cwd string) string {
	root := RepoRoot(cwd)
	if root == "" {
		return ""
	}
	return filepath.Base(root)
}

func CurrentBranch(repo string) string {
	ctx, cancel := context.WithTimeout(context.Background(), gitTimeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, "git", "-C", repo, "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

type FileStat struct {
	Added, Removed int
}

// parseNumstat merges git --numstat output into a FileStat map.
func parseNumstat(s string, files map[string]FileStat) {
	for _, ln := range strings.Split(strings.TrimSpace(s), "\n") {
		if ln == "" {
			continue
		}
		parts := strings.Fields(ln)
		if len(parts) < 3 {
			continue
		}
		a, _ := strconv.Atoi(parts[0]) // "-" → 0 for binary files
		r, _ := strconv.Atoi(parts[1])
		f := parts[2]
		cur := files[f]
		cur.Added += a
		cur.Removed += r
		files[f] = cur
	}
}

// DiffStat returns per-file added/removed lines for commits in [since, until].
// Only committed work is counted — uncommitted diffs are intentionally excluded
// to prevent double-counting across block boundaries.
func DiffStat(repo, sinceISO, untilISO string) map[string]FileStat {
	ctx, cancel := context.WithTimeout(context.Background(), gitTimeout)
	defer cancel()
	files := map[string]FileStat{}
	if out, err := exec.CommandContext(ctx, "git", "-C", repo, "log",
		"--since="+sinceISO, "--until="+untilISO,
		"--pretty=tformat:", "--numstat").Output(); err == nil {
		parseNumstat(string(out), files)
	}
	return files
}
