<script lang="ts">
  import type { Data } from "../lib/transform";

  let { data }: { data: Data } = $props();
  let entries = $derived([...data.timeline].reverse());
  let max = $derived(Math.max(1, ...entries.map((e) => e.focus)));
</script>

<section class="rounded-xl border border-stone-200 dark:border-stone-800 bg-stone-50 dark:bg-stone-900 shadow-sm dark:shadow-stone-950/30 p-5">
  <div class="flex items-baseline gap-3 mb-4">
    <h3 class="font-display text-lg text-stone-900 dark:text-stone-100">Timeline</h3>
    <span class="text-[11px] font-mono text-stone-400">newest first</span>
  </div>
  {#if !entries.length}
    <p class="text-stone-400 text-sm font-mono py-4">No blocks yet.</p>
  {:else}
    <div class="max-h-[440px] overflow-y-auto pr-1">
      <ol class="list-none m-0 p-0">
        {#each entries as e, i}
          <li class="grid grid-cols-[85px_16px_1fr] gap-3 py-2.5" class:opacity-35={!e.project}>
            <div class="text-[11px] text-stone-400 pt-0.5 font-mono tabular-nums">
              {e.from} → {e.to}
            </div>
            <div class="relative">
              <span
                class="absolute top-[5px] left-1/2 -translate-x-1/2 w-2.5 h-2.5 rounded-full border-2 z-[1]
                       {e.project ? 'bg-indigo-500 border-white dark:border-stone-900' : 'bg-stone-300 dark:bg-stone-600 border-white dark:border-stone-900'}"
              ></span>
              {#if i < entries.length - 1}
                <span class="absolute left-1/2 top-4 bottom-[-10px] w-px bg-stone-200 dark:bg-stone-700 -translate-x-1/2"></span>
              {/if}
            </div>
            <div class="pb-1 min-h-[26px]">
              {#if !e.project}
                <div class="text-stone-400 font-mono text-xs">— idle</div>
              {:else}
                <div class="flex items-center gap-2 flex-wrap">
                  <span class="text-sm text-stone-900 dark:text-stone-100 font-medium">{e.project}</span>
                  {#if e.branch}
                    <span class="inline-flex items-center px-1.5 py-0.5 rounded bg-indigo-50 dark:bg-indigo-950 text-indigo-700 dark:text-indigo-300 text-[10px] font-mono ring-1 ring-indigo-100 dark:ring-indigo-900">{e.branch}</span>
                  {/if}
                </div>
                <div class="text-xs text-stone-500 mt-0.5 font-mono">
                  <span class="tabular-nums font-medium text-stone-700 dark:text-stone-300">{e.focus}m</span> focus
                  {#if e.switches > 0}
                    <span class="text-stone-300 dark:text-stone-600 mx-1">·</span>
                    <span class="tabular-nums">{e.switches}</span> switches
                  {/if}
                </div>
                <div class="h-1 bg-stone-100 dark:bg-stone-800 rounded-full mt-1.5 w-3/5 overflow-hidden">
                  <span class="block h-full rounded-full bg-indigo-500" style:width="{(e.focus / max) * 100}%"></span>
                </div>
                {#if e.ai}
                  <div class="text-xs text-stone-600 dark:text-stone-300 bg-indigo-50 dark:bg-indigo-950/30 border-l-2 border-indigo-500 py-1.5 px-2.5 rounded-r-md leading-relaxed mt-2 whitespace-pre-wrap">{e.ai}</div>
                {/if}
              {/if}
            </div>
          </li>
        {/each}
      </ol>
    </div>
  {/if}
</section>
