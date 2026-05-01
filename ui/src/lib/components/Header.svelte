<script lang="ts">
  import type { ParsedData } from '../types';
  import type { Theme } from '../theme';
  import FlowMark from '../icons/FlowMark.svelte';
  import SvgIcon from '../icons/SvgIcon.svelte';
  import { ICONS } from '../icons/icons';

  interface Props {
    data: ParsedData;
    view: string;
    setView: (v: 'today' | 'week') => void;
    theme: Theme;
    cycleTheme: () => void;
  }
  let { data, view, setView, theme, cycleTheme }: Props = $props();

  const views = ['today', 'week'] as const;
</script>

<header class="flex items-center justify-between pb-[22px] mb-1.5">

  <div class="flex items-center gap-3.5">
    <FlowMark />
    <div>
      <div class="font-display text-[27px] leading-none mb-1 tracking-[-0.02em] text-slate-900 dark:text-slate-100">flowd</div>
      <div class="flex items-center gap-2 text-[12.5px] text-slate-500">
        <span class="font-mono text-[10.5px] tracking-[0.14em] uppercase text-slate-400">Generated</span>
        <span class="font-mono text-[12px] tabular-nums">{data.generated}</span>
        {#if data.hasData}
          <span class="text-slate-300 dark:text-slate-600">·</span>
          <span class="inline-flex items-center gap-[5px] px-2 py-0.5 rounded-full text-[11.5px] font-mono bg-slate-100 dark:bg-slate-700 border border-slate-200 dark:border-slate-600 text-slate-600 dark:text-slate-300">
            <span class="w-1.5 h-1.5 rounded-full bg-primary shadow-[0_0_0_3px_rgba(16,185,129,0.22)]"></span>
            {data.period}
          </span>
        {/if}
      </div>
    </div>
  </div>

  <div class="flex items-center gap-2.5">
    <div class="inline-flex bg-slate-100 dark:bg-slate-700 border border-slate-200 dark:border-slate-600 rounded-lg p-[3px]">
      {#each views as v}
        <button
          class="border-0 px-3.5 py-1.5 text-[12.5px] rounded-md cursor-pointer font-sans transition-colors duration-100 {view === v ? 'bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 shadow-sm' : 'bg-transparent text-slate-500 hover:text-slate-700 dark:hover:text-slate-300'}"
          title={v !== data.period ? `Run fw dashboard ${v} to generate ${v} data` : ''}
          onclick={() => setView(v)}
        >{v === 'today' ? 'Today' : 'Week'}</button>
      {/each}
    </div>
    <button
      class="w-9 h-9 rounded-lg bg-slate-100 dark:bg-slate-700 border border-slate-200 dark:border-slate-600 text-slate-500 dark:text-slate-400 inline-grid place-items-center cursor-pointer hover:text-slate-700 dark:hover:text-slate-200 hover:border-slate-300 dark:hover:border-slate-500 transition-colors duration-100"
      title="Theme: {theme}"
      onclick={cycleTheme}
    >
      <SvgIcon html={ICONS[theme]} size={18} />
    </button>
  </div>

</header>
