<script lang="ts">
  import type { Data } from "../lib/transform";

  let { data }: { data: Data } = $props();
  let days = $derived(Object.entries(data.weekHourlyByDay));
  let max = $derived(Math.max(1, ...days.flatMap(([, h]) => h)));
</script>

<section class="rounded-xl border border-stone-200 dark:border-stone-800 bg-stone-50 dark:bg-stone-900 shadow-sm dark:shadow-stone-950/30 p-5">
  <div class="flex items-baseline gap-3 mb-4">
    <h3 class="font-display text-lg text-stone-900 dark:text-stone-100">Activity heatmap</h3>
    <span class="text-[11px] font-mono text-stone-400">7 days × 24h</span>
  </div>
  <div class="flex flex-col gap-1.5 my-2">
    {#each days as [day, hrs]}
      <div class="flex items-center gap-2">
        <span class="font-mono text-[10px] text-stone-400 w-12 text-right">{day}</span>
        <div class="grid grid-cols-[repeat(24,1fr)] gap-[3px] flex-1 h-4">
          {#each hrs as v, h}
            {@const intensity = v / max}
            <span
              class="rounded bg-stone-100 dark:bg-stone-800 hover:scale-150 hover:z-10 relative transition-transform cursor-default"
              style:background={intensity > 0.01 ? `rgb(99 102 241 / ${Math.round(intensity * 100)}%)` : undefined}
              title="{day} {String(h).padStart(2, '0')}:00 · {v}m"
            ></span>
          {/each}
        </div>
      </div>
    {/each}
  </div>
</section>
