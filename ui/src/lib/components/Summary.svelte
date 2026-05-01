<script lang="ts">
  import type { ParsedData } from '../types';
  import { fmtHM } from '../format';

  interface Props { data: ParsedData }
  let { data }: Props = $props();

  const top3 = (arr: { name: string; min: number }[]) =>
    arr.slice(0, 3).map(x => `${x.name} ${x.min}m`).join(' · ') || '—';

  // grid-template-columns folded into the ROW constant via TW4 arbitrary property
  const ROW  = 'grid [grid-template-columns:82px_1fr] gap-2.5 py-2.5 border-t border-dashed border-slate-200 dark:border-slate-700 items-baseline';
  const TERM = 'font-mono text-[10.5px] tracking-[0.12em] uppercase text-slate-400';
  const DEF  = 'text-[12.5px] text-slate-600 dark:text-slate-300 leading-[1.55]';
</script>

<section class="bg-white dark:bg-slate-800 rounded-2xl border border-slate-200 dark:border-slate-700 p-5">
  <div class="flex items-baseline gap-3 mb-3.5">
    <h2 class="font-display text-[17px] tracking-tight text-slate-900 dark:text-slate-100">Summary</h2>
    <span class="text-[11.5px] font-mono text-slate-400">{data.period} · narrative</span>
  </div>

  <dl class="max-h-[420px] overflow-y-auto">
    <div class="{ROW} border-t-0 pt-1">
      <dt class={TERM}>Focus</dt>
      <dd class={DEF}>
        <b class="tabular-nums text-slate-900 dark:text-slate-100">{fmtHM(data.focus.totalMin)}</b>
        across <b class="tabular-nums text-slate-900 dark:text-slate-100">{data.focus.blocks}</b> blocks
      </dd>
    </div>
    {#if data.byProject.length}
      <div class={ROW}>
        <dt class={TERM}>Projects</dt>
        <dd class={DEF}>{top3(data.byProject)}</dd>
      </div>
    {/if}
    {#if data.byCommand.length}
      <div class={ROW}>
        <dt class={TERM}>Tools</dt>
        <dd class={DEF}>{top3(data.byCommand)}</dd>
      </div>
    {/if}
    <div class={ROW}>
      <dt class={TERM}>Code</dt>
      <dd class={DEF}>
        <span class="tabular-nums">{data.code.files} files</span>
        (<span class="text-accent tabular-nums">+{data.code.added}</span>
        <span class="text-danger tabular-nums"> −{data.code.removed}</span>)
      </dd>
    </div>
    {#if data.topRepo.name !== '—'}
      <div class={ROW}>
        <dt class={TERM}>Top repo</dt>
        <dd class={DEF}>
          <code class="bg-slate-100 dark:bg-slate-700 px-1 rounded font-mono text-[11px]">{data.topRepo.name}</code>
          {#if data.topRepo.branch}
            on <code class="bg-slate-100 dark:bg-slate-700 px-1 rounded font-mono text-[11px]">{data.topRepo.branch}</code>
          {/if}
        </dd>
      </div>
    {/if}
  </dl>
</section>
