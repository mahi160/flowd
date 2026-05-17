<script lang="ts">
  import type { Data } from "../lib/transform";
  import { fmtHM } from "../lib/format";
  import Tooltip from "./Tooltip.svelte";

  let { data }: { data: Data } = $props();

  const MONTH_NAMES = ["Jan","Feb","Mar","Apr","May","Jun","Jul","Aug","Sep","Oct","Nov","Dec"];

  let bars = $derived(data.monthBars);
  let maxMin = $derived(Math.max(1, ...bars.map((b) => b.min)));

  // Group bars by year so we can render year separators
  let byYear = $derived.by(() => {
    const map: Record<number, typeof bars> = {};
    for (const b of bars) {
      if (!map[b.year]) map[b.year] = [];
      map[b.year].push(b);
    }
    return Object.entries(map)
      .sort(([a], [b]) => Number(a) - Number(b))
      .map(([year, months]) => ({ year: Number(year), months }));
  });
</script>

<section class="rounded-xl border border-stone-200 dark:border-stone-800 bg-stone-50 dark:bg-stone-900 shadow-sm p-5">
  <div class="flex items-baseline gap-3 mb-5">
    <h3 class="font-display text-lg text-stone-900 dark:text-stone-100">Monthly overview</h3>
    <span class="text-[11px] font-mono text-stone-400">all time</span>
    <Tooltip text="Monthly total focused minutes since tracking started. Height = relative intensity. Hover for exact values." />
  </div>

  {#if !bars.length}
    <p class="text-stone-400 text-sm font-mono py-4">No data yet.</p>
  {:else}
    {#each byYear as { year, months }}
      <div class="mb-5">
        <!-- Year label -->
        <div class="text-[11px] font-mono text-stone-400 mb-2 uppercase tracking-wider">{year}</div>
        <div class="flex items-end gap-2 h-24">
          {#each months as bar}
            {@const pct = bar.min / maxMin}
            <div class="flex flex-col items-center gap-1 flex-1 min-w-[28px]">
              <div
                class="w-full rounded-t-md transition-all cursor-default relative group"
                style:height="{Math.max(4, pct * 80)}px"
                style:background={bar.min > 0 ? `rgb(99 102 241 / ${Math.round(Math.max(0.3, pct) * 100)}%)` : "rgb(226 232 240 / 60%)"}
                title="{MONTH_NAMES[bar.month - 1]} {bar.year}: {fmtHM(bar.min)} focused · {bar.blocks} block{bar.blocks !== 1 ? 's' : ''}"
              >
                <!-- Hover value label -->
                {#if bar.min > 0}
                  <div class="absolute -top-7 left-1/2 -translate-x-1/2 whitespace-nowrap
                              opacity-0 group-hover:opacity-100 transition-opacity
                              bg-stone-800 text-stone-100 text-[10px] font-mono
                              px-1.5 py-0.5 rounded pointer-events-none z-10">
                    {fmtHM(bar.min)}
                  </div>
                {/if}
              </div>
              <span class="text-[10px] font-mono text-stone-400">{MONTH_NAMES[bar.month - 1]}</span>
            </div>
          {/each}
          <!-- Pad remaining months in current year with empty slots -->
          {#if byYear[byYear.length - 1].year === year}
            {#each { length: 12 - months.length } as _}
              <div class="flex flex-col items-center gap-1 flex-1 min-w-[28px]">
                <div class="w-full rounded-t-md" style:height="4px" style:background="transparent"></div>
                <span class="text-[10px] font-mono text-transparent">—</span>
              </div>
            {/each}
          {/if}
        </div>
      </div>
    {/each}

    <!-- Summary row -->
    <div class="flex flex-wrap gap-x-6 gap-y-1.5 pt-3 border-t border-stone-100 dark:border-stone-800 text-[11px]">
      {#if data.trackingSince}
        <span class="font-mono text-stone-400">
          since: <span class="text-stone-700 dark:text-stone-200 font-medium">{data.trackingSince}</span>
        </span>
      {/if}
      <span class="font-mono text-stone-400">
        <span class="text-stone-700 dark:text-stone-200 font-medium">{data.activeDays}</span> active days
      </span>
      {#if data.bestDayDate}
        <span class="font-mono text-stone-400">
          best day: <span class="text-stone-700 dark:text-stone-200 font-medium">
            {new Date(data.bestDayDate + "T00:00:00").toLocaleDateString("default", { day: "numeric", month: "short", year: "numeric" })}
            ({fmtHM(data.bestDayMin)})
          </span>
        </span>
      {/if}
    </div>
  {/if}
</section>
