<script lang="ts">
  import Card from "../components/Card.svelte";

  const total = "7,800";
  const slices = [
    {
      id: "ts",
      label: "TypeScript",
      lines: 4220,
      pct: 54,
      color: "var(--primary)",
    },
    {
      id: "svelte",
      label: "Svelte",
      lines: 1840,
      pct: 24,
      color: "var(--accent)",
    },
    { id: "css", label: "CSS", lines: 980, pct: 13, color: "var(--danger)" },
    {
      id: "md",
      label: "Markdown",
      lines: 540,
      pct: 7,
      color: "var(--warning)",
    },
    {
      id: "shell",
      label: "Shell",
      lines: 220,
      pct: 2,
      color: "var(--foreground)",
    },
  ];
</script>

<Card heading="Languages" description="by lines touched">
  <div class="flex h-2 rounded-full overflow-hidden border border-border">
    {#each slices as l (l.id)}
      <span class="block h-full" style:flex={l.pct} style:background={l.color}
      ></span>
    {/each}
  </div>

  <ul class="m-0 p-0 list-none flex flex-col gap-2 mt-2">
    {#each slices as l (l.id)}
      <li
        class="grid grid-cols-[10px_100px_1fr_auto] gap-3 items-center text-xs"
      >
        <span class="block w-2.5 h-2.5 rounded-sm" style:background={l.color}
        ></span>
        <span class="font-mono text-foreground/80">{l.label}</span>
        <span class="h-1 bg-background rounded-full overflow-hidden">
          <span
            class="block h-full rounded-full"
            style:width="{l.pct}%"
            style:background={l.color}
          ></span>
        </span>
        <span class="font-mono tabular-nums text-foreground/60">
          {l.lines.toLocaleString()}
        </span>
      </li>
    {/each}
  </ul>

  <div
    class="mt-auto pt-3 border-t border-border flex justify-between items-baseline"
  >
    <span
      class="font-mono text-[10px] uppercase tracking-wider text-foreground/50"
    >
      total
    </span>
    <span class="font-display tabular-nums text-xl text-foreground">
      {total}
      <span class="text-foreground/40 text-sm">lines</span>
    </span>
  </div>
</Card>
