<script lang="ts">
  import type { ParsedData } from '../types';

  interface Props { data: ParsedData }
  let { data }: Props = $props();

  function streakSub(s: number): string {
    if (s === 0)   return 'start coding today';
    if (s >= 14)   return 'on fire 🔥';
    if (s >= 7)    return 'keep it up 🌿';
    return 'building momentum';
  }

  let s   = $derived(data.streakDays);
  let sub = $derived(streakSub(s));
</script>

<section class="bg-white dark:bg-slate-800 rounded-2xl border border-slate-200 dark:border-slate-700 p-5">
  <div class="flex items-baseline gap-3 mb-3.5">
    <h2 class="font-display text-[17px] tracking-tight text-slate-900 dark:text-slate-100">Streak</h2>
    <span class="text-[11.5px] font-mono text-slate-400">last 30 days</span>
  </div>

  <div class="flex items-end gap-2.5 mb-4">
    <span class="font-display text-[68px] leading-none tabular-nums text-primary tracking-[-2px]">{s}</span>
    <div class="flex flex-col gap-1 pb-2.5">
      <span class="text-[14.5px] text-slate-700 dark:text-slate-200 font-medium">day streak</span>
      <span class="font-mono text-[10.5px] tracking-widest uppercase text-slate-400">{sub}</span>
    </div>
  </div>

  <div class="grid grid-cols-10 gap-1 mb-2">
    {#each data.streakCells as cell}
      {@const active = cell.v > 0}
      <span
        class="rounded h-[14px] cursor-default transition-transform hover:scale-110 {cell.d === 29 ? 'ring-2 ring-offset-1 ring-primary/50' : ''} {active ? '' : 'bg-slate-100 dark:bg-slate-700'}"
        style={active ? `background: rgba(16,185,129,${0.15 + (cell.v / 4) * 0.85})` : undefined}
        title={cell.d === 29 ? 'today' : `${29 - cell.d}d ago`}
      ></span>
    {/each}
  </div>

  <div class="flex justify-between font-mono text-[10px] text-slate-400">
    <span>30d ago</span><span>today</span>
  </div>
</section>
