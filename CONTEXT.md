# flowd

A background daemon that tracks focused coding time inside tmux, writes a journal
entry every 30 focused minutes, and commits it to a private git repo so GitHub's
contribution graph reflects when work actually happened. Local-first: data lives in
SQLite and a git repo the user owns.

## Language

**Event**:
A single sampled observation of the active tmux pane (taken every poll interval),
or a state-change marker (cwd change, session change). The raw, append-only signal.
_Avoid_: tick, sample, datapoint

**Block**:
A summary unit built from events covering 30 focused minutes of work. One block
becomes one journal entry and one git commit.
_Avoid_: session, period, chunk

**Focused minute**:
A minute of activity that counts toward a block. Idle time (pane quiet beyond the
idle threshold, or lid closed) does not count.
_Avoid_: active minute, tracked time

**Context switch**:
A change of tmux session (project/context) within a block. Counted per block.
_Avoid_: switch (ambiguous), jump

**Journal**:
The private git repo (and its monthly markdown files) where block summaries are
committed. One subfolder per machine.
_Avoid_: log, diary, report

**AI session**:
A conversation with an AI coding CLI (pi, claude-code), reconstructed by reading
that tool's JSONL transcript files — not by watching tmux. Distinct from a tmux
session.
_Avoid_: chat, conversation

**Machine**:
A named device whose work is tracked into its own journal subfolder. Lets one
journal repo aggregate multiple computers.
_Avoid_: host, node, device

**Watch dir**:
A configured directory; only tmux panes whose cwd is inside a watch dir are tracked.
_Avoid_: project root, tracked path

**Standup**:
An AI-generated today/yesterday summary of work done per project, built from git
commit messages and flowd-tracked time. Shown on the dashboard and as a roll-up
header in the journal. Distinct from the per-block AI summary.
_Avoid_: recap, summary (overloaded), daily report

**Roll-up**:
The per-day header in the monthly journal file: total focused minutes, block count,
top repo, language totals, and the standup text. Sits above the per-block entries
for that day and is regenerated each time a new block lands.
_Avoid_: summary (overloaded), aggregate

**Interpreter pane**:
A tmux pane whose `pane_current_command` is a language runtime (`node`, `python`, …)
rather than the actual tool. flowd walks the process tree from the pane PID to
resolve the real foreground command (e.g. `pi`, `claude`).
_Avoid_: runtime pane, node pane
