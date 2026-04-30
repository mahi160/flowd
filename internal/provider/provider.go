package provider

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// Summarize passes prompt to an external AI command via stdin and returns output.
// command is a shell string like "claude prompt" or "gemini chat".
// Returns empty string if command is empty (no-op).
func Summarize(ctx context.Context, command, prompt string) (string, error) {
	if command == "" {
		return "", nil
	}

	parts := strings.Fields(command)
	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	cmd.Stdin = bytes.NewBufferString(prompt)

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("ai command %q: %w", command, err)
	}
	return strings.TrimSpace(string(out)), nil
}
