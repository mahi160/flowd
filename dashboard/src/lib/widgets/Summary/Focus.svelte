<script lang="ts">
  import Card from "../../components/Card.svelte";
  import SparklineChart from "../../components/SparklineChart.svelte";
  import { formatDuration } from "../../helper";
  import { flowd } from "../../store.svelte";

  const mins = $derived(flowd.total_focus_min);
  const time = $derived(formatDuration(mins));
  const isToday = $derived(flowd.period !== "week");
  const hourly = $derived.by(() => {
    const arr = Array(24).fill(0) as number[];

    for (const c of flowd.heatmap ?? []) {
      const h = isToday ? Math.floor(c.hour / 2) : c.hour;

      if (h >= 0 && h < 24) {
        arr[h] += c.minute;
      }
    }
    return arr;
  });
</script>

<Card heading="Focus">
  <div class="flex gap-2">
    {#each Object.entries(time) as [key, value]}
      {#if value}
        <div class="flex items-baseline gap-1">
          <h1
            class="text-5xl font-light text-primary font-display tabular-nums"
          >
            {value}
          </h1>
          <span class="font-mono text-foreground/50">{key}</span>
        </div>
      {/if}
    {/each}
  </div>

  <span class="font-mono text-foreground/50 text-[10px]">
    {flowd.total_blocks} focus blocks · {flowd.total_switches} context switches
  </span>

  <div class="mt-4">
    <SparklineChart data={hourly} />
    <hr class="border-foreground/15" />
  </div>
</Card>
