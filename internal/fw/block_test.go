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
