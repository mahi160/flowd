<script lang="ts">
  import type { Data } from "../lib/transform";
  import { fmtHM } from "../lib/format";

  let { data }: { data: Data } = $props();
  let max = $derived(Math.max(1, ...data.weekDays.map((d) => d.min)));
</script>

<section class="rounded-xl border border-stone-200 dark:border-stone-800 bg-stone-50 dark:bg-stone-900 shadow-sm dark:shadow-stone-950/30 p-5">
  <div class="flex items-baseline gap-3 mb-4">
    <h3 class="font-display text-lg text-stone-900 dark:text-stone-100">Week at a glance</h3>
    <span class="text-[11px] font-mono text-stone-400">focus per day</span>
  </div>
  <div class="grid grid-cols-7 gap-2 h-[200px] mt-3">
    {#each data.weekDays as d}
      <div class="flex flex-col">
        <div class="flex-1 bg-stone-50 dark:bg-stone-800 rounded-lg relative overflow-hidden">
          <span
            class="absolute left-0 right-0 bottom-0 bg-indigo-500 rounded-lg rounded-b-none flex justify-center pt-2"
            style:height="{(d.min / max) * 100}%"
          >
            {#if d.min > 0}
              <span class="font-mono text-[11px] font-medium text-stone-100">{fmtHM(d.min)}</span>
            {/if}
          </span>
        </div>
        <div class="text-center pt-2 mt-1.5 border-t border-stone-100 dark:border-stone-800">
          <div class="font-mono text-[10px] text-stone-400 uppercase tracking-wider">{d.day}</div>
          <div class="font-display text-base text-stone-900 dark:text-stone-100 tabular-nums">{d.date}</div>
        </div>
      </div>
    {/each}
  </div>
</section>
