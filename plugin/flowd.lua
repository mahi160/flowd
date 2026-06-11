-- flowd.lua — flowd neovim plugin
--
-- Writes the current buffer context to
--   ~/.local/share/flowd/nvim/<pid>.json
-- on every BufEnter and DirChanged event so the flowd daemon can read
-- the active file and filetype for accurate language attribution.
--
-- Install (lazy.nvim):
--   { dir = "/path/to/flowd/plugin", name = "flowd" }
--   -- or, once published:
--   { "mahi160/flowd", config = true }
--
-- The daemon degrades gracefully when this plugin is absent; all it loses
-- is per-file language precision before a git commit.

local M = {}

-- ── helpers ────────────────────────────────────────────────────────────────

-- flowd_dir returns ~/.local/share/flowd/nvim, creating it if needed.
local function flowd_dir()
  local base = vim.fn.stdpath("data") -- ~/.local/share/nvim
  -- Go one level up to get the flowd sibling directory.
  local parent = vim.fn.fnamemodify(base, ":h")
  local dir = parent .. "/flowd/nvim"
  vim.fn.mkdir(dir, "p")
  return dir
end

-- state_path returns the path for this neovim instance's state file,
-- keyed by the OS PID so multiple nvim sessions don't collide.
local function state_path()
  return flowd_dir() .. "/" .. vim.fn.getpid() .. ".json"
end

-- project_name derives a project name from the cwd: the git root's
-- basename when inside a repo, otherwise the cwd basename.
local function project_name(cwd)
  local git_root = vim.fn.systemlist("git -C " .. vim.fn.shellescape(cwd) .. " rev-parse --show-toplevel 2>/dev/null")[1]
  if git_root and git_root ~= "" then
    return vim.fn.fnamemodify(git_root, ":t")
  end
  return vim.fn.fnamemodify(cwd, ":t")
end

-- ── write ──────────────────────────────────────────────────────────────────

-- write_state encodes the current buffer context as JSON and atomically
-- writes it to the state file. Called on every BufEnter / DirChanged.
local function write_state()
  local bufnr = vim.api.nvim_get_current_buf()
  local ft    = vim.bo[bufnr].filetype or ""
  local file  = vim.api.nvim_buf_get_name(bufnr) or ""
  local cwd   = vim.fn.getcwd()

  -- Skip special buffers (terminals, oil.nvim, etc.).
  if vim.bo[bufnr].buftype ~= "" then
    return
  end

  local payload = vim.fn.json_encode({
    pid      = vim.fn.getpid(),
    cwd      = cwd,
    file     = file,
    filetype = ft,
    project  = project_name(cwd),
    ts       = os.time(),
  })

  local path = state_path()
  -- Write to a temp file then rename for atomicity.
  local tmp = path .. ".tmp"
  local f = io.open(tmp, "w")
  if f then
    f:write(payload)
    f:close()
    os.rename(tmp, path)
  end
end

-- ── cleanup ────────────────────────────────────────────────────────────────

local function cleanup()
  local path = state_path()
  os.remove(path)
  os.remove(path .. ".tmp")
end

-- ── setup ──────────────────────────────────────────────────────────────────

function M.setup(_opts)
  local group = vim.api.nvim_create_augroup("flowd", { clear = true })

  vim.api.nvim_create_autocmd({ "BufEnter", "BufWritePost" }, {
    group    = group,
    callback = write_state,
    desc     = "flowd: update active buffer state",
  })

  vim.api.nvim_create_autocmd("DirChanged", {
    group    = group,
    callback = write_state,
    desc     = "flowd: update cwd state",
  })

  vim.api.nvim_create_autocmd("VimLeave", {
    group    = group,
    callback = cleanup,
    desc     = "flowd: remove state file on exit",
  })

  -- Write initial state immediately.
  write_state()
end

-- Auto-setup when the file is sourced directly (e.g. placed in
-- ~/.config/nvim/plugin/). Plugin managers that call
-- require("flowd").setup() will trigger a second call, which is harmless
-- because the augroup is created with clear = true.
M.setup()

return M
