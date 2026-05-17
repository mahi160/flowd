<script lang="ts">
  import type { Data } from "../lib/transform";
  import { fmtHM } from "../lib/format";
  import Tooltip from "./Tooltip.svelte";

  const COLORS = ["#6366f1","#f59e0b","#ef4444","#10b981","#8b5cf6","#ec4899","#14b8a6","#f97316"];

  let { data }: { data: Data } = $props();
  let langs = $derived(data.byLanguage.slice(0, 8));
  let top = $derived(langs[0]?.min || 1);
  let total = $derived(langs.reduce((n, l) => n + l.min, 0));
</script>

<section class="rounded-xl border border-stone-200 dark:border-stone-800 bg-stone-50 dark:bg-stone-900 shadow-sm dark:shadow-stone-950/30 p-5">
  <div class="flex items-baseline gap-3 mb-4">
    <Tooltip text="Languages inferred from git diff file extensions during editor time in each focus block. Time is distributed across languages proportionally by lines touched.">
      <h3 class="font-display text-lg text-stone-900 dark:text-stone-100">Languages</h3>
    </Tooltip>
    <span class="text-[11px] font-mono text-stone-400">by session time</span>
  </div>
  {#if !langs.length}
    <p class="text-stone-400 text-sm font-mono py-4">No language data.</p>
  {:else}
    <div class="flex h-2 rounded-full overflow-hidden bg-stone-100 dark:bg-stone-800 mb-4">
      {#each langs as l, i}
        <span class="block h-full" style:flex={l.min} style:background={COLORS[i % COLORS.length]}></span>
      {/each}
    </div>
    <ul class="flex flex-col gap-2">
      {#each langs as l, i}
        <li class="grid grid-cols-[10px_80px_1fr_auto] gap-2.5 items-center">
          <span class="w-2.5 h-2.5 rounded-sm" style:background={COLORS[i % COLORS.length]}></span>
          <span class="text-stone-600 dark:text-stone-300 font-mono text-xs">{l.name}</span>
          <div class="h-1.5 bg-stone-100 dark:bg-stone-800 rounded-full overflow-hidden">
            <span class="block h-full rounded-full" style:width="{(l.min / top) * 100}%" style:background={COLORS[i % COLORS.length]}></span>
          </div>
          <span class="font-mono text-[11px] text-stone-400 tabular-nums">{l.min}m</span>
        </li>
      {/each}
    </ul>
    <div class="mt-4 pt-3 border-t border-stone-100 dark:border-stone-800 flex justify-between items-baseline">
      <span class="uppercase tracking-widest text-[10px] font-mono text-stone-400">total</span>
      <span class="font-display text-xl text-stone-900 dark:text-stone-100 tabular-nums">{fmtHM(total)}</span>
    </div>
  {/if}
</section>
