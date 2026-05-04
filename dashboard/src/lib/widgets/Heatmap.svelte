<script lang="ts">
  import Card from "../components/Card.svelte";

  const cells = Array.from({ length: 96 }, (_, i) => {
    const peak1 = Math.exp(-Math.pow((i - 24) / 8, 2));
    const peak2 = Math.exp(-Math.pow((i - 64) / 10, 2));
    return {
      i,
      intensity: Math.min(1, peak1 * 0.85 + peak2 + Math.random() * 0.1),
    };
  });
  const axis = ["8a", "10a", "12p", "2p", "4p", "6p", "8p"];
  const peakTime = "4:00 pm";
  const deepest = "14m deepest flow";
</script>

<Card heading="Activity heatmap" description="today · 15-min buckets · 8a → 8p">
  <div class="flex justify-end -mt-1">
    <span
      class="inline-flex items-center rounded-full bg-accent/15 text-accent border border-accent/30 font-mono text-[11px] px-2 py-0.5"
    >
      peak {peakTime}
    </span>
  </div>

  <div
    class="flex justify-between font-mono text-[10px] text-foreground/40 px-0.5"
  >
    {#each axis as a (a)}<span>{a}</span>{/each}
  </div>

  <div
    class="grid gap-[3px] h-12"
    style:grid-template-columns="repeat({cells.length}, 1fr)"
  >
    {#each cells as c (c.i)}
      <span
        class="rounded-[2px]"
        style:background="color-mix(in oklch, var(--accent) {Math.round(
          Math.max(4, c.intensity * 95),
        )}%, transparent)"
      ></span>
    {/each}
  </div>

  <div
    class="flex justify-between items-center pt-1 font-mono text-[10px] text-foreground/50"
  >
    <span class="inline-flex items-center gap-1">
      less
      {#each [0.1, 0.3, 0.55, 0.8, 1] as i (i)}
        <span
          class="block w-3 h-3 rounded-[3px]"
          style:background="color-mix(in oklch, var(--accent) {Math.round(
            i * 95,
          )}%, transparent)"
        ></span>
      {/each}
      more
    </span>
    <span>
      peak <span class="text-accent">{peakTime}</span>
      <span class="text-foreground/30 mx-1">·</span>
      {deepest}
    </span>
  </div>
</Card>
