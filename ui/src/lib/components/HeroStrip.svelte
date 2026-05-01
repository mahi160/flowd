<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import Chart from 'chart.js/auto';
  import type { ParsedData } from '../types';
  import BranchIcon from '../icons/BranchIcon.svelte';
  import { CHART_PRIMARY } from '../palette';

  export let data: ParsedData;
  export let label: string;

  let sparkCanvas: HTMLCanvasElement;
  let sparkChart: Chart | undefined;

  const CARD = 'bg-white dark:bg-slate-800 rounded-2xl border border-slate-200 dark:border-slate-700 p-5 flex flex-col min-h-[130px]';

  onMount(() => {
    sparkChart = new Chart(sparkCanvas, {
      type: 'line',
      data: {
        labels: data.hourly.map(() => ''),
        datasets: [{
          data: data.hourly,
          borderColor: CHART_PRIMARY,
          borderWidth: 1.5,
          pointRadius: 0,
          tension: 0.4,
          fill: { target: 'origin', above: `${CHART_PRIMARY}1a` },
        }],
      },
      options: {
        animation: false,
        responsive: true,
        maintainAspectRatio: false,
        plugins: { legend: { display: false }, tooltip: { enabled: false } },
        scales: { x: { display: false }, y: { display: false, min: 0 } },
      },
    });
  });

  onDestroy(() => sparkChart?.destroy());
</script>

<div class="grid grid-cols-2 md:grid-cols-4 gap-3.5">

  <!-- Focus -->
  <div class="bg-gradient-to-br from-white to-slate-50 dark:from-slate-800 dark:to-slate-900 rounded-2xl border border-slate-200 dark:border-slate-700 p-5 flex flex-col min-h-[130px]">
    <span class="font-mono text-[10.5px] tracking-widest uppercase text-slate-400">Focus {label}</span>
    <p class="font-display text-[52px] leading-none mt-1 tabular-nums text-primary">
      {Math.floor(data.focus.totalMin / 60)}<span class="text-[20px] opacity-60">h</span>{data.focus.totalMin % 60}<span class="text-[20px] opacity-60">m</span>
    </p>
    <p class="text-[12.5px] text-slate-400 tabular-nums mt-1">
      {data.focus.blocks} blocks · {data.focus.switches} switches
    </p>
    <div class="mt-auto h-8">
      <canvas bind:this={sparkCanvas} class="w-full h-full"></canvas>
    </div>
  </div>

  <!-- Machine -->
  <div class={CARD}>
    <span class="font-mono text-[10.5px] tracking-widest uppercase text-slate-400">Machine</span>
    <p class="font-display text-[24px] leading-snug text-slate-900 dark:text-slate-100 mt-1">{data.machine}</p>
    <p class="text-[12.5px] text-slate-400">{data.os}</p>
    <div class="mt-auto h-1 bg-slate-100 dark:bg-slate-700 rounded-full overflow-hidden">
      <div class="h-full w-1/2 rounded-full bg-gradient-to-r from-primary to-secondary"></div>
    </div>
  </div>

  <!-- Top repo -->
  <div class={CARD}>
    <span class="font-mono text-[10.5px] tracking-widest uppercase text-slate-400">Top repo</span>
    <p class="font-display text-[24px] leading-snug text-slate-900 dark:text-slate-100 mt-1">{data.topRepo.name}</p>
    {#if data.topRepo.branch}
      <div class="mt-1">
        <span class="inline-flex items-center gap-1 px-2 py-0.5 rounded text-[11px] font-mono bg-primary/10 text-primary">
          <BranchIcon size={10} /> {data.topRepo.branch}
        </span>
      </div>
    {/if}
    {#if data.byProject[0]}
      <p class="font-mono text-[11px] text-slate-400 mt-auto tabular-nums">{data.byProject[0].name} · {data.byProject[0].min}m</p>
    {/if}
  </div>

  <!-- Code -->
  <div class={CARD}>
    <span class="font-mono text-[10.5px] tracking-widest uppercase text-slate-400">Code</span>
    <p class="font-display text-[24px] leading-snug text-slate-900 dark:text-slate-100 mt-1 tabular-nums">{data.code.files} files</p>
    <p class="text-[12.5px] tabular-nums mt-1">
      <span class="text-accent">+{data.code.added}</span>
      <span class="text-slate-300 dark:text-slate-600 mx-1">/</span>
      <span class="text-danger">−{data.code.removed}</span>
    </p>
  </div>

</div>
