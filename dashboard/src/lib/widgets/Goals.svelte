<script lang="ts">
  import Card from "../components/Card.svelte";

  const goals = [
    { id: "deep", label: "Deep work", progress: "4h 02m / 4h", pct: 100, color: "var(--primary)" },
    { id: "commits", label: "Commits", progress: "7 / 5", pct: 100, color: "var(--primary)" },
    { id: "ai", label: "AI budget", progress: "$3.42 / $5.00", pct: 68, color: "var(--accent)" },
    { id: "switches", label: "Switches under", progress: "6 / 10", pct: 60, color: "var(--accent)" },
  ];
</script>

<Card heading="Goals" description="today">
  <ul class="m-0 p-0 list-none flex flex-col gap-3">
    {#each goals as g (g.id)}
      <li>
        <div class="flex justify-between items-baseline mb-1">
          <span class="text-sm text-foreground">{g.label}</span>
          <span class="font-mono text-xs tabular-nums text-foreground/60">
            {g.progress}
          </span>
        </div>
        <div class="h-[5px] bg-background rounded-full overflow-hidden">
          <span
            class="block h-full rounded-full"
            style:width="{Math.min(100, g.pct)}%"
            style:background={g.color}
          ></span>
        </div>
      </li>
    {/each}
  </ul>
</Card>
