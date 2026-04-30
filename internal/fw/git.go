package fw

import (
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

var extLang = map[string]string{
	".go": "Go", ".rs": "Rust", ".py": "Python", ".js": "JavaScript",
	".ts": "TypeScript", ".tsx": "TypeScript", ".jsx": "JavaScript",
	".rb": "Ruby", ".java": "Java", ".kt": "Kotlin", ".swift": "Swift",
	".c": "C", ".h": "C", ".cpp": "C++", ".cc": "C++", ".hpp": "C++",
	".cs": "C#", ".php": "PHP", ".sh": "Shell", ".zsh": "Shell", ".bash": "Shell",
	".lua": "Lua", ".vim": "Vim", ".md": "Markdown", ".yaml": "YAML", ".yml": "YAML",
	".toml": "TOML", ".json": "JSON", ".html": "HTML", ".css": "CSS", ".scss": "SCSS",
	".sql": "SQL", ".dockerfile": "Docker", ".tf": "Terraform", ".svelte": "Svelte",
	".vue": "Vue", ".dart": "Dart", ".ex": "Elixir", ".exs": "Elixir", ".erl": "Erlang",
	".scala": "Scala", ".clj": "Clojure", ".hs": "Haskell", ".ml": "OCaml",
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

func RepoRoot(cwd string) string {
	out, err := exec.Command("git", "-C", cwd, "rev-parse", "--show-toplevel").Output()
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
	out, err := exec.Command("git", "-C", repo, "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

type FileStat struct {
	Added, Removed int
}

// DiffStat returns per-file added/removed lines (deduped) for committed work
// in [start..end] plus current uncommitted diff. Files appearing in both
// sources are merged (lines summed, file counted once).
func DiffStat(repo, sinceISO, untilISO string) map[string]FileStat {
	files := map[string]FileStat{}
	merge := func(s string) {
		for _, ln := range strings.Split(strings.TrimSpace(s), "\n") {
			if ln == "" {
				continue
			}
			parts := strings.Fields(ln)
			if len(parts) < 3 {
				continue
			}
			a, _ := strconv.Atoi(parts[0]) // "-" → 0 (binary file)
			r, _ := strconv.Atoi(parts[1])
			f := parts[2]
			cur := files[f]
			cur.Added += a
			cur.Removed += r
			files[f] = cur
		}
	}
	if out, err := exec.Command("git", "-C", repo, "log",
		"--since="+sinceISO, "--until="+untilISO,
		"--pretty=tformat:", "--numstat").Output(); err == nil {
		merge(string(out))
	}
	if out, err := exec.Command("git", "-C", repo, "diff", "--numstat").Output(); err == nil {
		merge(string(out))
	}
	return files
}
