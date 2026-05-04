<script lang="ts">
  import Card from "../components/Card.svelte";

  const sessions = [
    { id: 1, name: "flowd", status: "active", windows: 5, pane: "nvim src/dashboard · flowd", uptime: "4h 08m", lastActive: "now" },
    { id: 2, name: "setthemacup", status: "active", windows: 3, pane: "lazygit · setthemacup", uptime: "2h 12m", lastActive: "4m ago" },
    { id: 3, name: "pi-mono", status: "idle", windows: 2, pane: "zsh · pi-mono", uptime: "1h 26m", lastActive: "42m ago" },
    { id: 4, name: "scratch", status: "idle", windows: 1, pane: "", uptime: "22m", lastActive: "1h ago" },
    { id: 5, name: "logs", status: "detached", windows: 1, pane: "tail -f", uptime: "5h 20m", lastActive: "2h ago" },
  ];

  const dotMap: Record<string, string> = {
    active: "var(--primary)",
    idle: "var(--warning)",
    detached: "var(--foreground)",
  };

  const sub = `${sessions.length} sessions`;
</script>

<Card heading="tmux sessions" description={sub}>
  <ul class="m-0 p-0 list-none flex flex-col">
    {#each sessions as s (s.id)}
      <li
        class="grid grid-cols-[10px_1fr_auto] gap-3 items-center py-2 border-t border-dashed border-border first:border-t-0"
      >
        <span
          class="block w-2 h-2 rounded-full"
          style:background={dotMap[s.status]}
        ></span>
        <div class="flex flex-col gap-0.5 min-w-0">
          <div class="flex items-center gap-2 flex-wrap">
            <span class="font-mono text-sm text-foreground">{s.name}</span>
            <span
              class="font-mono text-[10px] text-foreground/50 capitalize px-1.5 py-0.5 rounded bg-foreground/5"
            >
              {s.status}
            </span>
          </div>
          <div class="font-mono text-[10px] text-foreground/50 truncate">
            {s.windows} windows{s.pane ? ` · ${s.pane}` : ""}
          </div>
        </div>
        <div class="text-right">
          <div class="font-mono text-xs tabular-nums text-foreground/80">
            {s.uptime}
          </div>
          <div class="font-mono text-[10px] text-foreground/50">
            {s.lastActive}
          </div>
        </div>
      </li>
    {/each}
  </ul>
</Card>
