<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import Chart from 'chart.js/auto';
  import type { ParsedData } from '../types';
  import { fmtHour } from '../format';
  import { CHART_PRIMARY } from '../palette';

  export let data: ParsedData;

  let canvas: HTMLCanvasElement;
  let chart: Chart | undefined;

  $: peak = data.hourly.indexOf(Math.max(...data.hourly));
  $: first = data.hourly.findIndex(v => v > 0);
  $: last = data.hourly.reduce((a, v, i) => (v > 0 ? i : a), -1);

  onMount(() => {
    chart = new Chart(canvas, {
      type: 'bar',
      data: {
        labels: data.hourly.map((_, i) => i % 4 === 0 ? fmtHour(i) : ''),
        datasets: [{
          data: data.hourly,
          backgroundColor: `${CHART_PRIMARY}cc`,
          borderRadius: 3,
          borderSkipped: false,
        }],
      },
      options: {
        animation: false,
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: { display: false },
          tooltip: {
            callbacks: { label: ctx => ` ${fmtHour(ctx.dataIndex)}: ${ctx.parsed.y}m` },
          },
        },
        scales: {
          x: {
            grid: { display: false },
            ticks: { color: '#94a3b8', font: { size: 10, family: 'inherit' }, maxRotation: 0 },
            border: { display: false },
          },
          y: { display: false },
        },
      },
    });
  });

  onDestroy(() => chart?.destroy());
</script>

<section class="bg-white dark:bg-slate-800 rounded-2xl border border-slate-200 dark:border-slate-700 p-5">
  <div class="flex items-baseline gap-3 mb-3.5">
    <h2 class="font-display text-[17px] tracking-tight text-slate-900 dark:text-slate-100">Best hours</h2>
    <span class="text-[11.5px] font-mono text-slate-400">focus by hour of day</span>
  </div>

  <div class="h-[88px]">
    <canvas bind:this={canvas} class="w-full h-full"></canvas>
  </div>

  <div class="flex gap-[18px] pt-2.5 mt-2 border-t border-slate-100 dark:border-slate-700">
    <div class="flex flex-col gap-0.5">
      <span class="font-mono text-[10.5px] tracking-widest uppercase text-slate-400">peak</span>
      <span class="tabular-nums text-[13px] text-slate-700 dark:text-slate-200">{fmtHour(peak)}</span>
    </div>
    {#if first >= 0}
      <div class="flex flex-col gap-0.5">
        <span class="font-mono text-[10.5px] tracking-widest uppercase text-slate-400">start</span>
        <span class="tabular-nums text-[13px] text-slate-700 dark:text-slate-200">{fmtHour(first)}</span>
      </div>
    {/if}
    {#if last >= 0}
      <div class="flex flex-col gap-0.5">
        <span class="font-mono text-[10.5px] tracking-widest uppercase text-slate-400">end</span>
        <span class="tabular-nums text-[13px] text-slate-700 dark:text-slate-200">{fmtHour(last)}</span>
      </div>
    {/if}
  </div>
</section>
