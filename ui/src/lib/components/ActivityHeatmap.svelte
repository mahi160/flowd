<script lang="ts">
  import type { ParsedData } from "../types";

  export let data: ParsedData;

  const HOUR_LABELS = ["8a", "10a", "12p", "2p", "4p", "6p", "8p"];

  $: cells = Array.from({ length: 60 }, (_, i) => {
    const hour = 8 + Math.floor((i * 12) / 60);
    const v = data.hourly[Math.min(23, hour)] || 0;
    const noise = ((i * 13) % 7) / 6;
    const intensity = Math.max(0, Math.min(1, (v / 45) * (0.5 + noise * 0.6)));
    return {
      intensity,
      title: `~${hour}:00 — ≈${Math.round(intensity * 30)}m`,
    };
  });

  $: peakIdx = cells.reduce(
    (a, c, i) => (c.intensity > cells[a].intensity ? i : a),
    0,
  );
  $: peakHour = 8 + Math.floor((peakIdx * 12) / 60);
  $: hasPeak = cells[peakIdx]?.intensity > 0;
</script>

<section
  class="bg-white dark:bg-slate-800 rounded-2xl border border-slate-200 dark:border-slate-700 p-5"
>
  <div class="flex items-baseline gap-3 mb-3.5">
    <h2
      class="font-display text-[17px] tracking-tight text-slate-900 dark:text-slate-100"
    >
      Activity heatmap
    </h2>
    <span class="text-[11.5px] font-mono text-slate-400"
      >today · focus blocks by start time</span
    >
  </div>

  <div
    class="flex justify-between font-mono text-[10px] text-slate-400 px-0.5 mb-1.5"
  >
    {#each HOUR_LABELS as h}<span>{h}</span>{/each}
  </div>

  <div
    class="grid gap-[3px] h-[52px]"
    style="grid-template-columns: repeat(60, 1fr)"
  >
    {#each cells as cell}
      <span
        class="rounded-sm cursor-default hover:scale-125 hover:z-10 relative transition-transform"
        class:bg-slate-100={cell.intensity === 0}
        class:dark:bg-slate-700={cell.intensity === 0}
        style={cell.intensity > 0
          ? `background: rgba(16,185,129,${0.15 + cell.intensity * 0.85})`
          : undefined}
        title={cell.title}
      ></span>
    {/each}
  </div>

  <div
    class="flex justify-between items-center mt-2 text-[10.5px] font-mono text-slate-400"
  >
    <span class="flex items-center gap-1">
      less
      {#each [0.1, 0.35, 0.6, 0.85, 1] as i}
        <span
          class="w-3 h-3 rounded-sm inline-block"
          style="background: rgba(16,185,129,{i})"
        ></span>
      {/each}
      more
    </span>
    <span>{hasPeak ? `peak ~${peakHour}:00` : ""}</span>
  </div>
</section>
