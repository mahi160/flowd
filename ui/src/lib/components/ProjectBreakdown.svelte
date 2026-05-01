<script lang="ts">
  import type { ParsedData } from '../types';

  interface Props { data: ParsedData }
  let { data }: Props = $props();

  let items  = $derived(data.byProject.slice(0, 8));
  let topMin = $derived(items[0]?.min ?? 1);
</script>

<section class="bg-white dark:bg-slate-800 rounded-2xl border border-slate-200 dark:border-slate-700 p-5">
  <div class="flex items-baseline gap-3 mb-3.5">
    <h2 class="font-display text-[17px] tracking-tight text-slate-900 dark:text-slate-100">By project</h2>
    <span class="text-[11.5px] font-mono text-slate-400">time breakdown</span>
  </div>

  {#if !items.length}
    <p class="text-[12.5px] text-slate-400">No project data.</p>
  {:else}
    <ul class="flex flex-col gap-2">
      {#each items as p}
        <li class="grid items-center gap-2.5" style="grid-template-columns: 10px 1fr 1fr auto">
          <span class="w-2.5 h-2.5 rounded-sm" style="background:{p.color}"></span>
          <span class="font-mono text-[12px] text-slate-600 dark:text-slate-300 truncate">{p.name}</span>
          <div class="h-1 bg-slate-100 dark:bg-slate-700 rounded-full overflow-hidden">
            <span class="block h-full rounded-full" style="width:{(p.min / topMin) * 100}%; background:{p.color}"></span>
          </div>
          <span class="font-mono text-[11.5px] text-slate-500 tabular-nums">{p.min}m</span>
        </li>
      {/each}
    </ul>
  {/if}
</section>
