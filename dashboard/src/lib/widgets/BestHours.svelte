<script lang="ts">
  import Card from "../components/Card.svelte";

  const hourly = Array.from({ length: 24 }, (_, h) => {
    const morning = Math.exp(-Math.pow((h - 9) / 2, 2)) * 0.5;
    const afternoon = Math.exp(-Math.pow((h - 16) / 2.2, 2));
    return {
      hour: h,
      intensity: Math.min(1, morning + afternoon + Math.random() * 0.05),
    };
  });

  const peak = "4–5 pm";
  const first = "8:42a";
  const last = "5:58p";

  function tick(h: number) {
    if (h === 0) return "12a";
    if (h === 12) return "12p";
    return h > 12 ? `${h - 12}p` : `${h}a`;
  }
</script>

<Card heading="Best hours" description="focus by hour-of-day">
  <div class="grid grid-cols-24 gap-0.5 h-20 items-end mt-1 mb-5 relative">
    {#each hourly as h (h.hour)}
      <div class="relative h-full flex items-end">
        <span
          class="block w-full min-h-0.5 rounded-t-0.5 bg-[linear-gradient(180deg,var(--accent),color-mix(in_oklch,var(--accent)_40%,transparent))]"
          style:height="{h.intensity * 100}%"
        ></span>
        {#if h.hour % 4 === 0}
          <span
            class="absolute -bottom-5 left-0 font-mono text-[10px] text-foreground/40 whitespace-nowrap"
          >
            {tick(h.hour)}
          </span>
        {/if}
      </div>
    {/each}
  </div>

  <div class="flex gap-5 pt-3 border-t border-border">
    <div class="flex flex-col gap-0.5">
      <span
        class="font-mono text-[10px] uppercase tracking-wider text-foreground/50"
      >
        peak
      </span>
      <span class="font-mono text-xs tabular-nums text-foreground">{peak}</span>
    </div>
    <div class="flex flex-col gap-0.5">
      <span
        class="font-mono text-[10px] uppercase tracking-wider text-foreground/50"
      >
        first
      </span>
      <span class="font-mono text-xs tabular-nums text-foreground">{first}</span
      >
    </div>
    <div class="flex flex-col gap-0.5">
      <span
        class="font-mono text-[10px] uppercase tracking-wider text-foreground/50"
      >
        last
      </span>
      <span class="font-mono text-xs tabular-nums text-foreground">{last}</span>
    </div>
  </div>
</Card>

<style>
  .grid-cols-24 {
    grid-template-columns: repeat(24, 1fr);
  }
</style>
