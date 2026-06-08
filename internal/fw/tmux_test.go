package fw

import (
	"testing"
)

func TestParsePane(t *testing.T) {
	// 7-field format: session|window|pane|paneid|panepid|command|cwd
	raw := "main|editor|0|%0|1234|nvim|/home/user/project\n"
	p, err := parsePane(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Session != "main" {
		t.Errorf("Session: got %q, want %q", p.Session, "main")
	}
	if p.Window != "editor" {
		t.Errorf("Window: got %q, want %q", p.Window, "editor")
	}
	if p.Pane != "0" {
		t.Errorf("Pane: got %q, want %q", p.Pane, "0")
	}
	if p.PaneID != "%0" {
		t.Errorf("PaneID: got %q, want %q", p.PaneID, "%0")
	}
	if p.PanePID != 1234 {
		t.Errorf("PanePID: got %d, want 1234", p.PanePID)
	}
	if p.Command != "nvim" {
		t.Errorf("Command: got %q, want %q", p.Command, "nvim")
	}
	if p.Cwd != "/home/user/project" {
		t.Errorf("Cwd: got %q, want %q", p.Cwd, "/home/user/project")
	}
}

func TestParsePanePathWithPipes(t *testing.T) {
	// cwd is the last (7th) field and should not be further split.
	raw := "sess|win|1|%3|5678|zsh|/tmp/work\n"
	p, err := parsePane(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Cwd != "/tmp/work" {
		t.Errorf("Cwd: got %q, want %q", p.Cwd, "/tmp/work")
	}
	if p.PanePID != 5678 {
		t.Errorf("PanePID: got %d, want 5678", p.PanePID)
	}
}

func TestParsePaneTooFewFields(t *testing.T) {
	// Needs 7 fields; 6 is now too few.
	_, err := parsePane("main|editor|0|%0|1234|nvim") // missing cwd
	if err == nil {
		t.Error("expected error for truncated output, got nil")
	}
}

func TestParsePanePIDZeroOnBadInput(t *testing.T) {
	// If pane_pid field is not numeric, PanePID should be 0 (not a fatal error).
	raw := "main|editor|0|%0|notanumber|nvim|/tmp\n"
	p, err := parsePane(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.PanePID != 0 {
		t.Errorf("PanePID: got %d, want 0 for non-numeric input", p.PanePID)
	}
}
