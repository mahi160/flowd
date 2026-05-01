<script lang="ts">
  import type { ParsedData } from '../types';

  interface Props { data: ParsedData }
  let { data }: Props = $props();

  let tag      = $derived(data.aiRecap ? 'summary' : data.aiPerBlock > 0 ? 'inline' : 'setup');
  let tagClass = $derived(data.aiRecap
    ? 'bg-accent/10 border-accent/30 text-accent'
    : 'bg-slate-100 dark:bg-slate-700 border-slate-200 dark:border-slate-600 text-slate-500');

  const CODE = 'bg-slate-100 dark:bg-slate-700 px-1 rounded font-mono text-[11.5px]';
</script>

<section class="bg-white dark:bg-slate-800 rounded-2xl border border-slate-200 dark:border-slate-700 p-5">
  <div class="flex items-center gap-2.5 mb-4">
    <div class="w-[30px] h-[30px] rounded-[9px] shrink-0 bg-primary/10 text-primary border border-primary/20 grid place-items-center">
      <svg width="15" height="15" viewBox="0 0 24 24" fill="none">
        <path d="M12 3v3M12 18v3M3 12h3M18 12h3M5.5 5.5l2 2M16.5 16.5l2 2M5.5 18.5l2-2M16.5 7.5l2-2" stroke="currentColor" stroke-width="1.6" stroke-linecap="round"/>
        <circle cx="12" cy="12" r="3" stroke="currentColor" stroke-width="1.6"/>
      </svg>
    </div>
    <h2 class="font-display text-[17px] tracking-tight text-slate-900 dark:text-slate-100">
      {data.aiRecap ? 'AI recap' : 'AI insights'}
    </h2>
    {#if data.aiPerBlock > 0 && !data.aiRecap}
      <span class="ml-auto inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[11.5px] font-mono bg-slate-100 dark:bg-slate-700 border border-slate-200 dark:border-slate-600 text-slate-600 dark:text-slate-300">
        {data.aiPerBlock} block{data.aiPerBlock === 1 ? '' : 's'}
      </span>
    {/if}
  </div>

  <div class="flex gap-2.5 items-start">
    <span class="shrink-0 font-mono text-[9.5px] tracking-[0.1em] uppercase px-1.5 py-1 rounded border mt-0.5 {tagClass}">
      {tag}
    </span>
    <p class="text-[12.5px] text-slate-600 dark:text-slate-300 leading-[1.55]">
      {#if data.aiRecap}
        {data.aiRecap}
      {:else if data.aiPerBlock > 0}
        Per-block AI summaries are inline in the timeline. Run
        <code class={CODE}>fw dashboard --ai-recap</code> for an aggregate.
      {:else}
        Set <code class={CODE}>ai_enabled: true</code> and
        <code class={CODE}>ai_command</code> in your config to see AI insights here.
      {/if}
    </p>
  </div>
</section>
