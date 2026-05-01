<script lang="ts">
  import type { ParsedData } from '../types';
  import BranchIcon from '../icons/BranchIcon.svelte';

  export let data: ParsedData;

  $: entries = [...data.timeline].reverse();
  $: maxFocus = Math.max(1, ...entries.map(e => e.focus));
</script>

<section class="bg-white dark:bg-slate-800 rounded-2xl border border-slate-200 dark:border-slate-700 p-5">
  <div class="flex items-baseline gap-3 mb-3.5">
    <h2 class="font-display text-[17px] tracking-tight text-slate-900 dark:text-slate-100">Timeline</h2>
    <span class="text-[11.5px] font-mono text-slate-400">focus blocks · newest first</span>
  </div>

  {#if !entries.length}
    <p class="text-[12.5px] text-slate-400">No blocks recorded yet.</p>
  {:else}
    <ol class="flex flex-col max-h-[420px] overflow-y-auto pr-1 -mr-1 scrollbar-thin">
      {#each entries as entry, i}
        <li class="grid gap-3 py-2.5" style="grid-template-columns: 90px 16px 1fr" class:opacity-40={!entry.project}>
          <!-- Time -->
          <span class="font-mono text-[11px] text-slate-400 pt-0.5 tabular-nums">{entry.from} → {entry.to}</span>

          <!-- Rail -->
          <div class="relative">
            <span
              class="absolute top-[5px] left-1/2 -translate-x-1/2 w-2 h-2 rounded-full z-10 ring-2"
              class:bg-primary={!!entry.project}
              class:ring-white={!!entry.project}
              class:dark:ring-slate-800={!!entry.project}
              class:bg-slate-300={!entry.project}
              class:ring-slate-200={!entry.project}
              class:dark:bg-slate-600={!entry.project}
            ></span>
            {#if i < entries.length - 1}
              <span class="absolute left-1/2 top-[14px] bottom-[-10px] w-px bg-slate-200 dark:bg-slate-700 -translate-x-1/2"></span>
            {/if}
          </div>

          <!-- Body -->
          <div class="pb-1 min-h-[26px]">
            {#if !entry.project}
              <span class="text-[13px] text-slate-400">— idle</span>
            {:else}
              <div class="flex items-center gap-2 flex-wrap">
                <span class="text-[14.5px] font-medium text-slate-900 dark:text-slate-100">{entry.project}</span>
                {#if entry.branch}
                  <span class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[11px] font-mono bg-primary/10 text-primary">
                    <BranchIcon size={10} /> {entry.branch}
                  </span>
                {/if}
              </div>
              <p class="text-[12.5px] text-slate-400 mt-0.5 tabular-nums">
                {entry.focus}m focus{entry.switches > 0 ? ` · ${entry.switches} switches` : ''}
              </p>
              <div class="h-[3px] bg-slate-100 dark:bg-slate-700 rounded-full mt-1.5 w-[60%] overflow-hidden">
                <span class="block h-full rounded-full bg-primary" style="width:{(entry.focus / maxFocus) * 100}%"></span>
              </div>
              {#if entry.ai}
                <p class="text-[12.5px] text-slate-600 dark:text-slate-300 bg-primary/5 border-l-2 border-primary px-2.5 py-1.5 rounded-r mt-2 leading-[1.5] whitespace-pre-wrap">{entry.ai}</p>
              {/if}
            {/if}
          </div>
        </li>
      {/each}
    </ol>
  {/if}
</section>
