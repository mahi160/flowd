<script lang="ts">
  import Card from "../components/Card.svelte";

  const spend = "$3.42";
  const tokensIn = "412K";
  const tokensOut = "88K";
  const cacheHit = "64%";
  const convos = 12;

  const byModel = [
    {
      id: "sonnet",
      label: "Claude Sonnet 4.5",
      value: 192,
      valueLabel: "$1.92",
      color: "var(--primary)",
    },
    {
      id: "haiku",
      label: "Claude Haiku 4.5",
      value: 74,
      valueLabel: "$0.74",
      color: "var(--accent)",
    },
    {
      id: "gpt5",
      label: "GPT-5",
      value: 46,
      valueLabel: "$0.46",
      color: "var(--warning)",
    },
    {
      id: "pi",
      label: "Pi",
      value: 30,
      valueLabel: "$0.30",
      color: "var(--danger)",
    },
  ];

  const byTool = [
    {
      id: "claude-code",
      label: "Claude Code",
      value: 218,
      valueLabel: "$2.18",
      color: "var(--primary)",
    },
    {
      id: "cursor",
      label: "Cursor",
      value: 78,
      valueLabel: "$0.78",
      color: "var(--accent)",
    },
    {
      id: "aider",
      label: "aider",
      value: 32,
      valueLabel: "$0.32",
      color: "var(--warning)",
    },
    {
      id: "copilot",
      label: "Copilot",
      value: 14,
      valueLabel: "$0.14",
      color: "var(--danger)",
    },
  ];

  const expensive = {
    label: "Refactor aggregate.ts",
    tool: "Claude Code",
    model: "Sonnet 4.5",
    tokens: "84K → 12K",
    cost: "$0.42",
    when: "3:42 pm",
  };
</script>

<Card heading="AI usage" description="cost · tokens · cache">
  <div class="flex justify-end -mt-1">
    <span
      class="inline-flex items-center rounded-full bg-accent/15 text-accent border border-accent/30 font-mono text-[11px] px-2 py-0.5"
    >
      {spend}
    </span>
  </div>

  <div class="flex items-end justify-between gap-4 flex-wrap">
    <div>
      <div
        class="font-mono text-[10px] uppercase tracking-wider text-foreground/50"
      >
        spend
      </div>
      <div
        class="font-display text-4xl text-accent tabular-nums leading-none mt-1"
      >
        {spend}
      </div>
    </div>
    <div class="grid grid-cols-2 gap-x-5 gap-y-1 text-xs">
      <div class="flex items-baseline gap-2">
        <span
          class="font-mono text-[10px] uppercase tracking-wider text-foreground/50"
        >
          tokens in
        </span>
        <span class="tabular-nums font-mono text-foreground">{tokensIn}</span>
      </div>
      <div class="flex items-baseline gap-2">
        <span
          class="font-mono text-[10px] uppercase tracking-wider text-foreground/50"
        >
          tokens out
        </span>
        <span class="tabular-nums font-mono text-foreground">{tokensOut}</span>
      </div>
      <div class="flex items-baseline gap-2">
        <span
          class="font-mono text-[10px] uppercase tracking-wider text-foreground/50"
        >
          cache hit
        </span>
        <span class="tabular-nums font-mono text-primary">{cacheHit}</span>
      </div>
      <div class="flex items-baseline gap-2">
        <span
          class="font-mono text-[10px] uppercase tracking-wider text-foreground/50"
        >
          convos
        </span>
        <span class="tabular-nums font-mono text-foreground">{convos}</span>
      </div>
    </div>
  </div>

  <div>
    <div
      class="font-mono text-[10px] uppercase tracking-wider text-foreground/50 mb-1.5"
    >
      by model
    </div>
    <div class="flex h-2 rounded-full overflow-hidden border border-border">
      {#each byModel as s (s.id)}
        <span
          class="block h-full"
          style:flex={s.value}
          style:background={s.color}
        ></span>
      {/each}
    </div>
    <ul class="m-0 p-0 mt-2 grid grid-cols-2 gap-x-4 gap-y-1 list-none">
      {#each byModel as s (s.id)}
        <li class="grid grid-cols-[10px_1fr_auto] gap-2 items-center text-xs">
          <span class="block w-2.5 h-2.5 rounded-sm" style:background={s.color}
          ></span>
          <span class="font-mono text-foreground/80 truncate">{s.label}</span>
          <span class="font-mono tabular-nums text-foreground/60">
            {s.valueLabel}
          </span>
        </li>
      {/each}
    </ul>
  </div>

  <div>
    <div
      class="font-mono text-[10px] uppercase tracking-wider text-foreground/50 mb-1.5"
    >
      by tool
    </div>
    <div class="flex h-2 rounded-full overflow-hidden border border-border">
      {#each byTool as s (s.id)}
        <span
          class="block h-full"
          style:flex={s.value}
          style:background={s.color}
        ></span>
      {/each}
    </div>
    <ul class="m-0 p-0 mt-2 grid grid-cols-2 gap-x-4 gap-y-1 list-none">
      {#each byTool as s (s.id)}
        <li class="grid grid-cols-[10px_1fr_auto] gap-2 items-center text-xs">
          <span class="block w-2.5 h-2.5 rounded-sm" style:background={s.color}
          ></span>
          <span class="font-mono text-foreground/80 truncate">{s.label}</span>
          <span class="font-mono tabular-nums text-foreground/60">
            {s.valueLabel}
          </span>
        </li>
      {/each}
    </ul>
  </div>

  <div class="border-t border-border pt-3">
    <div
      class="font-mono text-[10px] uppercase tracking-wider text-foreground/50 mb-1"
    >
      most expensive call
    </div>
    <div class="text-sm text-foreground mb-0.5">{expensive.label}</div>
    <div class="flex items-center gap-2 flex-wrap text-[11px] font-mono">
      <span class="text-foreground/60">{expensive.tool}</span>
      <span class="text-foreground/30">·</span>
      <span class="text-foreground/60">{expensive.model}</span>
      <span class="text-foreground/30">·</span>
      <span class="text-foreground/60 tabular-nums">{expensive.tokens}</span>
      <span class="text-foreground/30">·</span>
      <span class="text-accent tabular-nums">{expensive.cost}</span>
      <span class="text-foreground/30">·</span>
      <span class="text-foreground/50 tabular-nums">{expensive.when}</span>
    </div>
  </div>
</Card>
