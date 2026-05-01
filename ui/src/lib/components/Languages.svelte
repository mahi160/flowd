<script lang="ts">
  import type { ParsedData } from '../types';
  import { fmtHM } from '../format';

  interface Props { data: ParsedData }
  let { data }: Props = $props();

  let langs = $derived(data.byLanguage.slice(0, 8));
  let top   = $derived(langs[0]?.min ?? 1);
  let total = $derived(langs.reduce((n, l) => n + l.min, 0));
</script>

<section class="bg-white dark:bg-slate-800 rounded-2xl border border-slate-200 dark:border-slate-700 p-5">
  <div class="flex items-baseline gap-3 mb-3.5">
    <h2 class="font-display text-[17px] tracking-tight text-slate-900 dark:text-slate-100">Languages</h2>
    <span class="text-[11.5px] font-mono text-slate-400">by time in session</span>
  </div>

  {#if !langs.length}
    <p class="text-[12.5px] text-slate-400">No language data — inferred from git diffs.</p>
  {:else}
    <div class="flex h-[7px] rounded-full overflow-hidden border border-slate-200 dark:border-slate-600 mb-3">
      {#each langs as l}
        <span class="block h-full" style="flex:{l.min}; background:{l.color}" title="{l.name} · {l.min}m"></span>
      {/each}
    </div>

    <ul class="flex flex-col gap-2">
      {#each langs as l}
        <li class="grid items-center gap-2.5" style="grid-template-columns: 10px 90px 1fr auto">
          <span class="w-2.5 h-2.5 rounded-sm" style="background:{l.color}"></span>
          <span class="font-mono text-[12px] text-slate-600 dark:text-slate-300">{l.name}</span>
          <div class="h-1 bg-slate-100 dark:bg-slate-700 rounded-full overflow-hidden">
            <span class="block h-full rounded-full" style="width:{(l.min / top) * 100}%; background:{l.color}"></span>
          </div>
          <span class="font-mono text-[11.5px] text-slate-500 tabular-nums">{l.min}m</span>
        </li>
      {/each}
    </ul>

    <div class="flex justify-between items-baseline mt-3 pt-2.5 border-t border-slate-100 dark:border-slate-700">
      <span class="font-mono text-[10.5px] tracking-widest uppercase text-slate-400">total</span>
      <span class="font-display text-[21px] tabular-nums text-slate-900 dark:text-slate-100">
        {fmtHM(total)} <span class="text-slate-400 text-[15px]">tracked</span>
      </span>
    </div>
  {/if}
</section>
