<script lang="ts">
  import type { ParsedData } from '../types';

  export let data: ParsedData;

  $: days = Object.entries(data.weekHourlyByDay);
  $: max = Math.max(1, ...days.flatMap(([, h]) => h));
</script>

<section class="bg-white dark:bg-slate-800 rounded-2xl border border-slate-200 dark:border-slate-700 p-5">
  <div class="flex items-baseline gap-3 mb-3.5">
    <h2 class="font-display text-[17px] tracking-tight text-slate-900 dark:text-slate-100">Activity heatmap</h2>
    <span class="text-[11.5px] font-mono text-slate-400">last 7 days × hour of day</span>
  </div>

  <div class="flex flex-col gap-[5px] my-2">
    {#each days as [day, hrs]}
      <div class="flex items-center gap-2">
        <span class="font-mono text-[10px] text-slate-400 w-11 text-right shrink-0">{day}</span>
        <div class="grid gap-[3px] flex-1 h-[16px]" style="grid-template-columns: repeat(24, 1fr)">
          {#each hrs as v, h}
            <span
              class="rounded-sm cursor-default"
              class:bg-slate-100={v === 0}
              class:dark:bg-slate-700={v === 0}
              style={v > 0 ? `background: rgba(16,185,129,${0.15 + (v / max) * 0.85})` : undefined}
              title="{day} {String(h).padStart(2, '0')}:00 · {v}m"
            ></span>
          {/each}
        </div>
      </div>
    {/each}
  </div>

  <div class="flex justify-between items-center mt-2 text-[10.5px] font-mono text-slate-400">
    <span class="flex items-center gap-1">
      less
      {#each [0.1, 0.35, 0.6, 0.85, 1] as i}
        <span class="w-3 h-3 rounded-sm inline-block" style="background: rgba(16,185,129,{i})"></span>
      {/each}
      more
    </span>
    <span>24h × 7 days</span>
  </div>
</section>
