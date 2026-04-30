package initwizard

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mahi/flowd/internal/config"
)

func ask(r *bufio.Reader, prompt, def string) string {
	if def != "" {
		fmt.Printf("  %s [%s]: ", prompt, def)
	} else {
		fmt.Printf("  %s: ", prompt)
	}
	line, _ := r.ReadString('\n')
	line = strings.TrimSpace(line)
	if line == "" {
		return def
	}
	return line
}

func askBool(r *bufio.Reader, prompt string, def bool) bool {
	defStr := "n"
	if def {
		defStr = "y"
	}
	ans := ask(r, prompt+" (y/n)", defStr)
	return strings.ToLower(ans) == "y"
}

func askInt(r *bufio.Reader, prompt string, def int) int {
	ans := ask(r, prompt, strconv.Itoa(def))
	n, err := strconv.Atoi(ans)
	if err != nil {
		return def
	}
	return n
}

// Run walks the user through config questions and returns a populated Config.
func Run() (*config.Config, error) {
	cfg := config.Defaults()
	r := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Println("  flowd init — answer questions to create your config")
	fmt.Println("  Press Enter to accept [default].")
	fmt.Println()

	fmt.Println("  ── Database ──────────────────────────────────────")
	cfg.DBPath = ask(r, "DB path", cfg.DBPath)

	fmt.Println()
	fmt.Println("  ── Polling ───────────────────────────────────────")
	cfg.PollIntervalSec = askInt(r, "tmux poll interval (seconds)", cfg.PollIntervalSec)
	cfg.SummaryIntervalMin = askInt(r, "summary block interval (minutes)", cfg.SummaryIntervalMin)

	fmt.Println()
	fmt.Println("  ── Keys ──────────────────────────────────────────")
	cfg.TrackKeys = askBool(r, "track key counts (aggregated, no raw keys)", cfg.TrackKeys)

	fmt.Println()
	fmt.Println("  ── Git sync ──────────────────────────────────────")
	cfg.PushDB = askBool(r, "sync logs to a private git repo", cfg.PushDB)
	if cfg.PushDB {
		cfg.RepoPath = ask(r, "repo path", cfg.RepoPath)
		cfg.Branch = ask(r, "branch", cfg.Branch)
	}

	fmt.Println()
	fmt.Println("  ── AI summaries ──────────────────────────────────")
	fmt.Println("  Leave blank to skip. Examples: claude prompt / gemini chat")
	cfg.AICommand = ask(r, "AI command", cfg.AICommand)

	return cfg, nil
}
