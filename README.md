# Flowd

Local-first coding memory daemon for tmux users.

Flowd observes your tmux workflow (Neovim, AI CLIs, shell, lazygit), stores structured local telemetry in SQLite, generates 30-minute summaries, and syncs to your private Git repo.

CLI: `fw`

## Core Features

- tmux-focused activity tracking
- active pane / cwd / command detection
- session time + focus metrics
- key count buckets (not raw keys)
- repo/project inference
- AI summaries via your own CLI tools
- local SQLite database
- private git sync
- future reports/dashboard

## Philosophy

- Local first
- Minimal resource usage
- Useful metrics only
- No cloud dependency
- No spyware behavior

## Quick Start

```bash
go build -o fw ./cmd/fw
fw start
fw status
fw summary now
fw report today
```

---

## PRD.md

# Product Requirements Document

## Product Name

Flowd

## Goal

Create a private developer activity memory system for tmux-based coding workflows.

## Primary User

Solo developer working in tmux using:

- Neovim
- shell
- AI CLI tools
- lazygit

## Core Problems Solved

- No memory of how time was spent
- Hard to measure focus vs distraction
- Lost work context across days
- No reliable work journal
- Weak insight into coding habits

## MVP Scope

1. Detect active tmux pane/session/window
2. Detect cwd + active command
3. Log lightweight events to SQLite
4. Build 30-minute work blocks
5. Generate text summaries
6. Push logs to private Git repo
7. Show daily/weekly reports

## Non Goals

- Employee surveillance
- Raw keylogging
- Cloud SaaS
- Screenshot capture
- Browser spying
- Gamified fake productivity scores

## Success Metrics

- CPU idle overhead under 1%
- RAM under 50MB target
- Reliable 30-min summaries
- Accurate repo detection
- Daily usefulness to user
