<script lang="ts">
  import Card from "../components/Card.svelte";

  const monthLabel = "May 2026";
  const dayLabels = ["M", "T", "W", "T", "F", "S", "S"];

  const days = Array.from({ length: 42 }, (_, i) => {
    const dayNum = i - 4;
    const isOther = dayNum < 1 || dayNum > 31;
    return {
      iso: isOther ? "" : `2026-05-${String(dayNum).padStart(2, "0")}`,
      dayLabel: isOther ? "" : String(dayNum).padStart(2, "0"),
      intensity: isOther ? 0 : Math.max(0, Math.sin(i / 4) * 0.6 + Math.random() * 0.4),
      isToday: dayNum === 3,
      isOther,
      focusMin: isOther ? 0 : Math.floor(Math.random() * 360),
    };
  });
</script>

<Card heading="Month" description={monthLabel}>
  <div class="grid grid-cols-7 gap-1.5">
    {#each dayLabels as l, i (i)}
      <div
        class="font-mono text-[10px] uppercase text-foreground/50 text-center pb-1"
      >
        {l}
      </div>
    {/each}
    {#each days as d, i (i)}
      <div
        class="aspect-square rounded-md border border-transparent flex items-end justify-end p-1 {d.isToday
          ? 'ring-2 ring-accent'
          : ''} {d.isOther ? 'opacity-30' : ''}"
        style:background="color-mix(in oklch, var(--primary) {Math.round(
          Math.max(5, d.intensity * 85),
        )}%, transparent)"
        title="{d.iso} · {d.focusMin}m"
      >
        <span class="font-mono text-[10px] text-foreground/70 tabular-nums">
          {d.dayLabel}
        </span>
      </div>
    {/each}
  </div>
</Card>
