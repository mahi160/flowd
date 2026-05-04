<script lang="ts">
  import Card from "../components/Card.svelte";

  const metrics = [
    { key: "focus", label: "focus", today: "4h 02m", other: "3h 24m", direction: "up", isGood: true, delta: "+38m" },
    { key: "blocks", label: "blocks", today: "15", other: "12", direction: "up", isGood: true, delta: "+3" },
    { key: "ai_spend", label: "ai spend", today: "$3.42", other: "$3.88", direction: "down", isGood: true, delta: "-$0.46" },
    { key: "switches", label: "switches", today: "6", other: "11", direction: "down", isGood: true, delta: "-5" },
  ];

  function arrow(d: string) {
    return d === "up" ? "↑" : d === "down" ? "↓" : "→";
  }
  function tone(m: (typeof metrics)[number]) {
    if (m.direction === "flat") return "text-foreground/60";
    return m.isGood ? "text-primary" : "text-danger";
  }
</script>

<Card heading="Comparison" description="vs yesterday">
  <ul class="m-0 p-0 list-none flex flex-col">
    {#each metrics as m (m.key)}
      <li
        class="grid grid-cols-[1fr_auto_auto_auto] gap-3 items-baseline py-2 border-t border-dashed border-border first:border-t-0"
      >
        <span
          class="font-mono text-[10px] uppercase tracking-wider text-foreground/50"
        >
          {m.label}
        </span>
        <span class="font-mono text-[11px] tabular-nums text-foreground/50">
          {m.other}
        </span>
        <span class="font-mono text-sm tabular-nums text-foreground">
          {m.today}
        </span>
        <span
          class="font-mono text-[11px] tabular-nums px-1.5 py-0.5 rounded {tone(
            m,
          )} bg-current/10"
        >
          <span class={tone(m)}>{arrow(m.direction)} {m.delta}</span>
        </span>
      </li>
    {/each}
  </ul>
</Card>
