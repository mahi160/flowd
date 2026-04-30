package fw

import (
	"testing"
)

func TestParseNumstat(t *testing.T) {
	input := "10\t2\tmain.go\n5\t0\tinternal/fw/block.go\n-\t-\tbinary.png\n"
	files := map[string]FileStat{}
	parseNumstat(input, files)

	cases := []struct {
		file string
		a, r int
	}{
		{"main.go", 10, 2},
		{"internal/fw/block.go", 5, 0},
		{"binary.png", 0, 0},
	}
	for _, c := range cases {
		s := files[c.file]
		if s.Added != c.a || s.Removed != c.r {
			t.Errorf("%s: got +%d -%d, want +%d -%d", c.file, s.Added, s.Removed, c.a, c.r)
		}
	}
}

func TestParseNumstatMerge(t *testing.T) {
	files := map[string]FileStat{}
	parseNumstat("3\t1\tfoo.go\n", files)
	parseNumstat("7\t2\tfoo.go\n", files) // same file, second source (uncommitted diff)
	s := files["foo.go"]
	if s.Added != 10 || s.Removed != 3 {
		t.Errorf("merge: got +%d -%d, want +10 -3", s.Added, s.Removed)
	}
}

func TestLangFromCommand(t *testing.T) {
	cases := map[string]string{
		"node":    "JavaScript",
		"bun":     "JavaScript",
		"deno":    "TypeScript",
		"python":  "Python",
		"python3": "Python",
		"go":      "Go",
		"cargo":   "Rust",
		"unknown": "",
	}
	for cmd, want := range cases {
		if got := LangFromCommand(cmd); got != want {
			t.Errorf("LangFromCommand(%q) = %q, want %q", cmd, got, want)
		}
	}
}

func TestLangOf(t *testing.T) {
	cases := map[string]string{
		".go":  "Go",
		".py":  "Python",
		".ts":  "TypeScript",
		".tsx": "TypeScript",
		".rs":  "Rust",
		"":     "other",
		".xyz": "xyz",
	}
	for ext, want := range cases {
		if got := langOf(ext); got != want {
			t.Errorf("langOf(%q) = %q, want %q", ext, got, want)
		}
	}
}

func TestDistributeByLines(t *testing.T) {
	stats := map[string]FileStat{
		"main.go":   {Added: 80, Removed: 20}, // 100 lines → Go
		"README.md": {Added: 10, Removed: 0},  // 10 lines → Markdown
	}
	out := distributeByLines(stats, 110)
	// Go should get 100/110 of 110 = 100 min
	if out["Go"] != 100 {
		t.Errorf("Go: got %d, want 100", out["Go"])
	}
	if out["Markdown"] != 10 {
		t.Errorf("Markdown: got %d, want 10", out["Markdown"])
	}
}

func TestDistributeByLinesEmpty(t *testing.T) {
	if out := distributeByLines(nil, 10); out != nil {
		t.Errorf("expected nil for empty stats, got %v", out)
	}
	if out := distributeByLines(map[string]FileStat{"a.go": {1, 0}}, 0); out != nil {
		t.Errorf("expected nil for zero minutes, got %v", out)
	}
}
