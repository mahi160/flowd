package fw

import (
	"testing"
)

// ── parsePPIDFromStat ─────────────────────────────────────────────────────────

func TestParsePPIDFromStat_Normal(t *testing.T) {
	// Typical /proc/pid/stat line: pid (comm) state ppid ...
	stat := "1234 (my-proc) S 5678 1234 1234 0 -1 4194304"
	if got := parsePPIDFromStat(stat); got != 5678 {
		t.Errorf("got %d, want 5678", got)
	}
}

func TestParsePPIDFromStat_CommWithSpaces(t *testing.T) {
	// comm may contain spaces and even '(' — only the last ')' is safe.
	stat := "42 (my weird (proc)) S 99 42 42 0"
	if got := parsePPIDFromStat(stat); got != 99 {
		t.Errorf("got %d, want 99", got)
	}
}

func TestParsePPIDFromStat_Empty(t *testing.T) {
	if got := parsePPIDFromStat(""); got != 0 {
		t.Errorf("got %d, want 0", got)
	}
}

func TestParsePPIDFromStat_MissingParen(t *testing.T) {
	if got := parsePPIDFromStat("1234 noparens S 5678"); got != 0 {
		t.Errorf("got %d, want 0", got)
	}
}

// ── parsePS ───────────────────────────────────────────────────────────────────

func TestParsePS_Basic(t *testing.T) {
	output := `  100   1 launchd
  200 100 zsh
  300 200 node
  301 300 pi
`
	tree := parsePS(output)
	if len(tree) != 4 {
		t.Fatalf("got %d entries, want 4", len(tree))
	}
	if tree[301].PPID != 300 {
		t.Errorf("pid 301 ppid: got %d, want 300", tree[301].PPID)
	}
	if tree[301].exe != "pi" {
		t.Errorf("pid 301 exe: got %q, want %q", tree[301].exe, "pi")
	}
}

func TestParsePS_CaseNormalized(t *testing.T) {
	output := "1 0 LaunchD\n2 1 Node\n"
	tree := parsePS(output)
	if tree[1].exe != "launchd" {
		t.Errorf("exe not lower-cased: %q", tree[1].exe)
	}
	if tree[2].exe != "node" {
		t.Errorf("exe not lower-cased: %q", tree[2].exe)
	}
}

func TestParsePS_IgnoresBadLines(t *testing.T) {
	output := "bad line\n  \n1 0 init\n"
	tree := parsePS(output)
	if len(tree) != 1 {
		t.Errorf("got %d entries, want 1", len(tree))
	}
}

// ── bfsAICLI ─────────────────────────────────────────────────────────────────

// buildTestTree constructs a synthetic procEntry map for BFS tests.
// items: [][]int{pid, ppid} plus a string exe field encoded as last slice
// element — we use a separate helper to keep it readable.
func buildTree(entries []procEntry) map[int]procEntry {
	m := make(map[int]procEntry, len(entries))
	for _, e := range entries {
		m[e.PID] = e
	}
	return m
}

func TestBfsAICLI_DirectChild(t *testing.T) {
	// shell(1) → node(2) → pi(3)
	tree := buildTree([]procEntry{
		{PID: 1, PPID: 0, exe: "zsh"},
		{PID: 2, PPID: 1, exe: "node"},
		{PID: 3, PPID: 2, exe: "pi"},
	})
	if got := bfsAICLI(1, tree); got != "pi" {
		t.Errorf("got %q, want %q", got, "pi")
	}
}

func TestBfsAICLI_RootIsAI(t *testing.T) {
	tree := buildTree([]procEntry{
		{PID: 7, PPID: 1, exe: "claude"},
	})
	if got := bfsAICLI(7, tree); got != "claude" {
		t.Errorf("got %q, want %q", got, "claude")
	}
}

func TestBfsAICLI_NotFound(t *testing.T) {
	tree := buildTree([]procEntry{
		{PID: 1, PPID: 0, exe: "zsh"},
		{PID: 2, PPID: 1, exe: "node"},
		{PID: 3, PPID: 2, exe: "webpack"},
	})
	if got := bfsAICLI(1, tree); got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestBfsAICLI_UnrelatedAIProcessIgnored(t *testing.T) {
	// pid 10 (shell we're watching) has no AI child.
	// pid 20 (different shell) does have 'aider'.
	tree := buildTree([]procEntry{
		{PID: 10, PPID: 1, exe: "bash"},
		{PID: 11, PPID: 10, exe: "node"},
		{PID: 20, PPID: 1, exe: "zsh"},
		{PID: 21, PPID: 20, exe: "aider"},
	})
	if got := bfsAICLI(10, tree); got != "" {
		t.Errorf("got %q, want empty (unrelated pane should not match)", got)
	}
}

func TestBfsAICLI_DeepNesting(t *testing.T) {
	// shell → node → npm → codex
	tree := buildTree([]procEntry{
		{PID: 1, PPID: 0, exe: "zsh"},
		{PID: 2, PPID: 1, exe: "node"},
		{PID: 3, PPID: 2, exe: "npm"},
		{PID: 4, PPID: 3, exe: "codex"},
	})
	if got := bfsAICLI(1, tree); got != "codex" {
		t.Errorf("got %q, want %q", got, "codex")
	}
}

// Zero / negative PID guards live in ResolveAICLI; verify bfsAICLI
// correctly finds nothing when the root pid is absent from the tree.
func TestBfsAICLI_PIDNotInTree(t *testing.T) {
	tree := buildTree([]procEntry{{PID: 5, PPID: 0, exe: "pi"}})
	// Starting from pid 99 which doesn't exist in tree.
	if got := bfsAICLI(99, tree); got != "" {
		t.Errorf("got %q, want empty for unknown root pid", got)
	}
}

// ── IsInterpreter ─────────────────────────────────────────────────────────────

func TestIsInterpreter(t *testing.T) {
	for _, cmd := range []string{"node", "python", "python3", "bun", "deno", "ruby"} {
		if !IsInterpreter(cmd) {
			t.Errorf("IsInterpreter(%q) = false, want true", cmd)
		}
	}
	for _, cmd := range []string{"nvim", "zsh", "git", "claude", "pi"} {
		if IsInterpreter(cmd) {
			t.Errorf("IsInterpreter(%q) = true, want false", cmd)
		}
	}
}
