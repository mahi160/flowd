package fw

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"
)

func openTestDB(t *testing.T) *DB {
	t.Helper()
	f, err := os.CreateTemp("", "fw-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })
	d, err := OpenDB(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Close() })
	return d
}

func seedEvents(t *testing.T, d *DB, start time.Time, pollSec int, count int, meta PaneMeta) {
	t.Helper()
	raw, _ := json.Marshal(meta)
	for i := 0; i < count; i++ {
		ts := start.Add(time.Duration(i*pollSec) * time.Second)
		if _, err := d.Exec(
			`INSERT INTO events (ts, type, value, meta) VALUES (?, ?, ?, ?)`,
			ts.UTC(), EvActive, meta.Pane, string(raw),
		); err != nil {
			t.Fatal(err)
		}
	}
}

func TestBuildBlockFocusedMin(t *testing.T) {
	d := openTestDB(t)
	start := time.Now().UTC().Truncate(time.Minute)
	end := start.Add(30 * time.Minute)
	const pollSec = 3
	// 600 ticks × 3s = 1800s = 30 min
	seedEvents(t, d, start, pollSec, 600, PaneMeta{
		Session: "main", Window: "editor", Pane: "0",
		Command: "nvim", Category: "editor", Cwd: "/tmp/proj", Repo: "proj",
	})

	b, err := BuildBlock(context.Background(), d, start, end, pollSec, false)
	if err != nil {
		t.Fatal(err)
	}
	if b.FocusedMin != 30 {
		t.Errorf("FocusedMin: got %d, want 30", b.FocusedMin)
	}
}

func TestBuildBlockByTool(t *testing.T) {
	d := openTestDB(t)
	start := time.Now().UTC().Truncate(time.Minute)
	end := start.Add(30 * time.Minute)
	const pollSec = 60

	seedEvents(t, d, start, pollSec, 20, PaneMeta{
		Session: "s", Window: "editor", Category: "editor", Pane: "0", Cwd: "/tmp",
	})
	seedEvents(t, d, start.Add(20*time.Minute), pollSec, 10, PaneMeta{
		Session: "s", Window: "git", Category: "git", Pane: "1", Cwd: "/tmp",
	})

	b, err := BuildBlock(context.Background(), d, start, end, pollSec, false)
	if err != nil {
		t.Fatal(err)
	}
	if b.ByTool["editor"] == 0 {
		t.Error("expected editor time > 0")
	}
	if b.ByTool["git"] == 0 {
		t.Error("expected git time > 0")
	}
	if b.ByTool["editor"] <= b.ByTool["git"] {
		t.Errorf("editor (%d) should exceed git (%d)", b.ByTool["editor"], b.ByTool["git"])
	}
}

func TestBuildBlockContextSwitches(t *testing.T) {
	d := openTestDB(t)
	start := time.Now().UTC().Truncate(time.Minute)
	end := start.Add(30 * time.Minute)

	meta, _ := json.Marshal(PaneMeta{Session: "s1", Window: "w", Pane: "0", Cwd: "/tmp"})
	for i, sess := range []string{"proj-a", "proj-b", "proj-a"} {
		d.Exec(`INSERT INTO events (ts, type, value, meta) VALUES (?, ?, ?, ?)`,
			start.Add(time.Duration(i)*time.Minute).UTC(), EvSessionChange, sess, string(meta))
	}

	b, err := BuildBlock(context.Background(), d, start, end, 3, false)
	if err != nil {
		t.Fatal(err)
	}
	if b.Switches != 3 {
		t.Errorf("Switches: got %d, want 3", b.Switches)
	}
}

// TestBuildBlockAIPaneNoPhantomLang verifies that when an event carries
// Category:"ai" (as set by the proctree resolver relabelling an interpreter
// pane), no phantom runtime language (JavaScript/Python/…) leaks into
// Block.Languages.
// TestBuildBlockNvimPluginLanguage verifies that when PaneMeta carries a
// NvimFiletype (from the flowd.lua plugin), the block's Languages map reflects
// the plugin-provided language rather than falling back to git-diff.
func TestBuildBlockNvimPluginLanguage(t *testing.T) {
	d := openTestDB(t)
	start := time.Now().UTC().Truncate(time.Minute)
	end := start.Add(30 * time.Minute)
	const pollSec = 60

	// 20 ticks × 60 s = 20 min in nvim editing a Go file (plugin reports ft=go).
	seedNvimEvents(t, d, start, pollSec, 20, PaneMeta{
		Session: "main", Window: "editor", Pane: "0",
		Command: "nvim", Category: "editor",
		Cwd: "/tmp/proj", Repo: "proj",
		NvimFiletype: "go",
	})

	b, err := BuildBlock(context.Background(), d, start, end, pollSec, false)
	if err != nil {
		t.Fatal(err)
	}
	if b.Languages["Go"] == 0 {
		t.Errorf("expected Go in Languages; got %v", b.Languages)
	}
	// No other language should appear from the nvim path.
	for lang, min := range b.Languages {
		if lang != "Go" && min > 0 {
			t.Errorf("unexpected language %q (%d min)", lang, min)
		}
	}
}

// TestBuildBlockNvimFallback verifies that when NvimFiletype is empty (plugin
// absent), the block builder does not error and still builds correctly.
func TestBuildBlockNvimFallback(t *testing.T) {
	d := openTestDB(t)
	start := time.Now().UTC().Truncate(time.Minute)
	end := start.Add(30 * time.Minute)
	const pollSec = 60

	seedEvents(t, d, start, pollSec, 20, PaneMeta{
		Session: "main", Window: "editor", Pane: "0",
		Command: "nvim", Category: "editor",
		Cwd: "/tmp/proj", Repo: "",
		NvimFiletype: "", // plugin absent
	})

	b, err := BuildBlock(context.Background(), d, start, end, pollSec, false)
	if err != nil {
		t.Fatalf("fallback should not error: %v", err)
	}
	if b.FocusedMin == 0 {
		t.Error("expected non-zero FocusedMin even without nvim plugin")
	}
}

// seedNvimEvents is like seedEvents but stores the full PaneMeta including
// NvimFiletype by encoding the whole struct as meta JSON.
func seedNvimEvents(t *testing.T, d *DB, start time.Time, pollSec int, count int, meta PaneMeta) {
	t.Helper()
	raw, _ := json.Marshal(meta)
	for i := 0; i < count; i++ {
		ts := start.Add(time.Duration(i*pollSec) * time.Second)
		if _, err := d.Exec(
			`INSERT INTO events (ts, type, value, meta) VALUES (?, ?, ?, ?)`,
			ts.UTC(), EvActive, meta.Pane, string(raw),
		); err != nil {
			t.Fatal(err)
		}
	}
}

func TestBuildBlockAIPaneNoPhantomLang(t *testing.T) {
	d := openTestDB(t)
	start := time.Now().UTC().Truncate(time.Minute)
	end := start.Add(30 * time.Minute)
	const pollSec = 3

	// Simulate 300 ticks where the pane command is "node" but Category="ai"
	// (what track.resolveCommand produces after proctree resolution).
	seedEvents(t, d, start, pollSec, 300, PaneMeta{
		Session:  "main",
		Window:   "ai",
		Pane:     "0",
		Command:  "pi",   // relabelled by proctree resolver
		Category: "ai",  // ← key: NOT "runtime"
		Cwd:      "/tmp/proj",
		Repo:     "proj",
	})

	b, err := BuildBlock(context.Background(), d, start, end, pollSec, false)
	if err != nil {
		t.Fatal(err)
	}

	// Time should appear under "pi", not "node".
	if b.ByTool["pi"] == 0 {
		t.Error("expected pi in ByTool")
	}
	if b.ByTool["node"] != 0 {
		t.Errorf("node should not appear in ByTool after relabelling, got %d min", b.ByTool["node"])
	}

	// No phantom JS/Python language should appear since Category!="runtime".
	if min := b.Languages["JavaScript"]; min != 0 {
		t.Errorf("phantom JavaScript in Languages: %d min", min)
	}
	if min := b.Languages["Python"]; min != 0 {
		t.Errorf("phantom Python in Languages: %d min", min)
	}
}

func TestTopKey(t *testing.T) {
	m := map[string]int{"a": 1, "b": 5, "c": 3}
	if got := topKey(m); got != "b" {
		t.Errorf("topKey: got %q, want %q", got, "b")
	}
}

func TestTopLine(t *testing.T) {
	m := map[string]int{"Go": 20, "Python": 5, "Rust": 10}
	line := topLine(m, "min")
	if line == "" {
		t.Error("expected non-empty topLine")
	}
	// Go should appear first (highest)
	if line[:2] != "Go" {
		t.Errorf("expected Go first in %q", line)
	}
}

func TestRender(t *testing.T) {
	loc := time.Local
	start := time.Date(2026, 1, 1, 14, 0, 0, 0, loc)
	end := time.Date(2026, 1, 1, 14, 30, 0, 0, loc)
	b := &Block{
		StartTS:    start,
		EndTS:      end,
		FocusedMin: 25,
		Switches:   2,
		Repo:       "myrepo",
		Branch:     "main",
		ByTool:     map[string]int{"editor": 20},
		ByProject:  map[string]int{"myrepo": 25},
		Languages:  map[string]int{"Go": 15},
	}
	s := render(b)
	for _, want := range []string{"14:00", "14:30", "25 min", "myrepo", "main", "Go"} {
		if !containsStr(s, want) {
			t.Errorf("render output missing %q:\n%s", want, s)
		}
	}
}

func containsStr(s, sub string) bool {
	return len(sub) > 0 && len(s) >= len(sub) && func() bool {
		for i := 0; i <= len(s)-len(sub); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	}()
}
