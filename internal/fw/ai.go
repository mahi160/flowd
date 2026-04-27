package fw

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// RunAI runs `sh -c <command>` with `prompt + "\n\n---\n\n" + body` on stdin
// and returns trimmed stdout. The 60s timeout protects the daemon from
// hanging AI commands.
//
// Examples of `command`:
//
//	claude --print
//	codex exec --model gpt-4o
//	llm -m claude-3-5-sonnet
//	pi chat
//	opencode run
func RunAI(ctx context.Context, command, prompt, body string) (string, error) {
	if command == "" {
		return "", nil
	}
	cctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(cctx, "sh", "-c", command)
	input := body
	if prompt != "" {
		input = prompt + "\n\n---\n\n" + body
	}
	cmd.Stdin = strings.NewReader(input)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("ai %q: %w", command, err)
	}
	return strings.TrimSpace(string(out)), nil
}
