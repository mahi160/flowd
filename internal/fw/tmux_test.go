package fw

import (
	"testing"
)

func TestParsePane(t *testing.T) {
	raw := "main|editor|0|%0|nvim|/home/user/project\n"
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
	if p.Command != "nvim" {
		t.Errorf("Command: got %q, want %q", p.Command, "nvim")
	}
	if p.Cwd != "/home/user/project" {
		t.Errorf("Cwd: got %q, want %q", p.Cwd, "/home/user/project")
	}
}

func TestParsePanePathWithPipes(t *testing.T) {
	// cwd should not be split even if it contained something unusual; only 6 fields
	raw := "sess|win|1|%3|zsh|/tmp/work\n"
	p, err := parsePane(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Cwd != "/tmp/work" {
		t.Errorf("Cwd: got %q, want %q", p.Cwd, "/tmp/work")
	}
}

func TestParsePaneTooFewFields(t *testing.T) {
	_, err := parsePane("main|editor|0")
	if err == nil {
		t.Error("expected error for truncated output, got nil")
	}
}
