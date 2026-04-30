## CONFIG.md

# Example config.yaml

```yaml
poll_interval_sec: 3
summary_interval_min: 30
track_keys: true
track_raw_keys: false
push_db: true
repo_path: ~/flowd-private
branch: main
ai_command: claude-code summarize --stdin
exclude_paths:
  - ~/.ssh
  - ~/Downloads
```

## Supported AI Commands

Examples:

```bash
claude prompt
gemini chat
codex run
pi summarize
python myscript.py
```

---

## Suggested First Prompt To Agent

Read all docs. Build Flowd in TASKS.md order.
Keep code minimal, modular, production-grade.
After each phase:

1. explain changes
2. run tests
3. update TASKS.md
4. commit changes
