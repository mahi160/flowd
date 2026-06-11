package fw

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// aiCLIArgv is the set of known AI-CLI executable base-names we recognise
// when walking a process tree. Lower-case; compared after filepath.Base +
// lower-casing each argv[0].
var aiCLIArgv = map[string]bool{
	"pi":       true,
	"claude":   true,
	"aider":    true,
	"codex":    true,
	"opencode": true,
	"llm":      true,
	"sgpt":     true,
	"tgpt":     true,
	"gemini":   true,
	"copilot":  true,
}

// interpreters is the set of process names that warrant a tree walk.
// When tmux reports one of these as pane_current_command, the real tool
// may be a child process.
var interpreters = map[string]bool{
	"node":    true,
	"bun":     true,
	"deno":    true,
	"python":  true,
	"python3": true,
	"ruby":    true,
	"perl":    true,
}

// IsInterpreter reports whether cmd is a known interpreter whose real
// underlying tool should be looked up via ResolveAICLI.
func IsInterpreter(cmd string) bool {
	return interpreters[strings.ToLower(cmd)]
}

// procEntry holds the minimal info we need from a single OS process.
type procEntry struct {
	PID  int
	PPID int
	// exe is the base-name of the executable (argv[0]).
	exe string
}

// ResolveAICLI walks the process tree rooted at panePID and returns the
// name of the first descendant whose executable matches a known AI CLI
// (e.g. "pi", "claude"). Returns "" if none is found or on any error.
//
// On Linux the tree is built from /proc; on macOS a single `ps` call
// builds it. The call is intentionally cheap enough for every poll tick
// when the pane command is a known interpreter.
func ResolveAICLI(panePID int) string {
	if panePID <= 0 {
		return ""
	}
	var tree map[int]procEntry
	var err error
	if runtime.GOOS == "linux" {
		tree, err = procTreeLinux()
	} else {
		tree, err = procTreePS()
	}
	if err != nil || len(tree) == 0 {
		return ""
	}
	return bfsAICLI(panePID, tree)
}

// bfsAICLI performs a breadth-first search starting at rootPID over the
// child relationships in tree, returning the first AI-CLI exe name found.
func bfsAICLI(rootPID int, tree map[int]procEntry) string {
	// Build pid→children index.
	children := make(map[int][]int, len(tree))
	for pid, e := range tree {
		children[e.PPID] = append(children[e.PPID], pid)
	}

	queue := []int{rootPID}
	seen := map[int]bool{rootPID: true}
	for len(queue) > 0 {
		pid := queue[0]
		queue = queue[1:]

		if e, ok := tree[pid]; ok && aiCLIArgv[e.exe] {
			return e.exe
		}
		for _, child := range children[pid] {
			if !seen[child] {
				seen[child] = true
				queue = append(queue, child)
			}
		}
	}
	return ""
}

// ── Linux: /proc ──────────────────────────────────────────────────────────────

func procTreeLinux() (map[int]procEntry, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}
	tree := make(map[int]procEntry, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}
		ppid := ppidFromStat(pid)
		exe := exeFromCmdline(pid)
		if exe == "" {
			continue
		}
		tree[pid] = procEntry{PID: pid, PPID: ppid, exe: exe}
	}
	return tree, nil
}

// ppidFromStat extracts the PPID from /proc/<pid>/stat.
// Format: "pid (comm) state ppid ..." — comm can contain spaces, so we
// search for the last ')' before parsing the remaining fields.
func ppidFromStat(pid int) int {
	data, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/stat")
	if err != nil {
		return 0
	}
	return parsePPIDFromStat(string(data))
}

func parsePPIDFromStat(stat string) int {
	idx := strings.LastIndex(stat, ")")
	if idx < 0 {
		return 0
	}
	rest := strings.TrimLeft(stat[idx+1:], " ")
	// remaining fields: state ppid ...
	parts := strings.Fields(rest)
	if len(parts) < 2 {
		return 0
	}
	ppid, _ := strconv.Atoi(parts[1])
	return ppid
}

// exeFromCmdline extracts the base-name of argv[0] from /proc/<pid>/cmdline
// (NUL-separated).
func exeFromCmdline(pid int) string {
	data, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/cmdline")
	if err != nil || len(data) == 0 {
		return ""
	}
	// First NUL-terminated field is argv[0].
	arg0 := string(bytes.SplitN(data, []byte{0}, 2)[0])
	return strings.ToLower(filepath.Base(arg0))
}

// ── macOS / generic: ps ───────────────────────────────────────────────────────

func procTreePS() (map[int]procEntry, error) {
	// -axo: all procs, custom output.
	// "pid ppid comm" gives the exe name without arguments, avoiding
	// whitespace ambiguity in the command column.
	out, err := exec.Command("ps", "-axo", "pid=,ppid=,comm=").Output()
	if err != nil {
		return nil, err
	}
	return parsePS(string(out)), nil
}

func parsePS(output string) map[int]procEntry {
	lines := strings.Split(output, "\n")
	tree := make(map[int]procEntry, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		pid, err := strconv.Atoi(fields[0])
		if err != nil {
			continue
		}
		ppid, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}
		exe := strings.ToLower(filepath.Base(fields[2]))
		tree[pid] = procEntry{PID: pid, PPID: ppid, exe: exe}
	}
	return tree
}
