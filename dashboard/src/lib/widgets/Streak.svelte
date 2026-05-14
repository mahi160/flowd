<script lang="ts">
  import Card from "../components/Card.svelte";

  const currentDays = 14;
  const longestDays = 21;
  const days = Array.from({ length: 30 }, (_, i) => ({
    date: `2026-04-${String(i + 4).padStart(2, "0")}`,
    intensity: Math.max(0, Math.sin(i / 2.5) * 0.55 + Math.random() * 0.45),
    isToday: i === 29,
    focusMin: Math.floor(Math.random() * 360),
  }));
</script>

<Card>
  <div class="flex items-baseline gap-2 mb-1">
    <span
      class="font-display text-[17px] font-medium tracking-tight text-foreground"
    >
      Streak
    </span>
    <span class="font-display tabular-nums text-3xl text-primary leading-none">
      {currentDays}
    </span>
    <span class="font-mono text-xs text-foreground/50">days</span>
  </div>
  <div class="font-mono text-[11px] text-foreground/50 mb-3">
    {days.length} days · longest {longestDays}
  </div>

  <div
    class="grid grid-rows-5 gap-1.5 flex-1"
    style:grid-template-columns="repeat({Math.ceil(days.length / 5)}, 1fr)"
  >
    {#each days as d (d.date)}
      <span
        class="aspect-square rounded-md {d.isToday ? 'ring-2 ring-accent' : ''}"
        style:background="color-mix(in oklch, var(--primary) {Math.round(
          Math.max(8, d.intensity * 90),
        )}%, transparent)"
        title="{d.date} · {d.focusMin}m"
      ></span>
    {/each}
  </div>

  <div
    class="flex items-center gap-3 mt-3 font-mono text-[10px] text-foreground/50"
  >
    <span>30D AGO</span>
    <span class="flex-1 h-px bg-border"></span>
    <span>TODAY</span>
  </div>
</Card>
