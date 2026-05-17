<script lang="ts">
  import type { Data } from "../lib/transform";

  let { data }: { data: Data } = $props();

  // Show all 24 hours (no artificial 8am-8pm clipping, no fake noise).
  let cells = $derived.by(() => {
    const max = Math.max(1, ...data.hourly);
    const peakHour = data.hourly.indexOf(Math.max(...data.hourly));
    const result = data.hourly.map((v, hour) => ({
      intensity: v / max,
      hour,
      minutes: v,
    }));
    return { cells: result, peakHour, hasData: max > 0 };
  });

  const fmtHour = (h: number) =>
    h === 0 ? "12a" : h < 12 ? `${h}a` : h === 12 ? "12p" : `${h - 12}p`;
</script>

<section class="rounded-xl border border-stone-200 dark:border-stone-800 bg-stone-50 dark:bg-stone-900 shadow-sm dark:shadow-stone-950/30 p-5">
  <div class="flex items-baseline gap-3 mb-4">
    <h3 class="font-display text-lg text-stone-800 dark:text-stone-100">Activity heatmap</h3>
    <span class="text-[11px] font-mono text-stone-400">today · 24h</span>
  </div>
  <!-- Hour labels at 0, 6, 12, 18, 23 -->
  <div class="flex justify-between font-mono text-[10px] text-stone-400 px-0.5 mb-2">
    {#each [0, 4, 8, 12, 16, 20, 23] as h}
      <span>{fmtHour(h)}</span>
    {/each}
  </div>
  <!-- 24 cells, one per hour -->
  <div class="grid grid-cols-[repeat(24,1fr)] gap-[3px] h-14">
    {#each cells.cells as cell}
      <span
        class="rounded bg-stone-100 dark:bg-stone-800 hover:scale-150 hover:z-10 relative transition-transform cursor-default"
        style:background={cell.intensity > 0.01
          ? `rgb(99 102 241 / ${Math.round(Math.max(0.15, cell.intensity) * 100)}%)`
          : undefined}
        title="{fmtHour(cell.hour)} — {cell.minutes}m"
      ></span>
    {/each}
  </div>
  <div class="flex justify-between items-center text-[10px] text-stone-400 font-mono pt-2">
    <span class="flex items-center gap-1">
      less
      {#each [0.15, 0.4, 0.65, 0.85, 1] as i}
        <span class="w-3 h-3 rounded" style:background="rgb(99 102 241 / {Math.round(i * 100)}%)"></span>
      {/each}
      more
    </span>
    {#if cells.hasData}
      <span>peak {fmtHour(cells.peakHour)}</span>
    {/if}
  </div>
</section>
