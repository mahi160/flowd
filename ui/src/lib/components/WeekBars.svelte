<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import Chart from 'chart.js/auto';
  import type { ParsedData } from '../types';
  import { fmtHM } from '../format';
  import { CHART_PRIMARY } from '../palette';

  export let data: ParsedData;

  let canvas: HTMLCanvasElement;
  let chart: Chart | undefined;

  onMount(() => {
    chart = new Chart(canvas, {
      type: 'bar',
      data: {
        labels: data.weekDays.map(d => [d.day, d.date]),
        datasets: [{
          data: data.weekDays.map(d => d.min),
          backgroundColor: `${CHART_PRIMARY}cc`,
          borderRadius: 8,
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
            callbacks: { label: ctx => ` ${fmtHM(ctx.parsed.y as number)}` },
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
    <h2 class="font-display text-[17px] tracking-tight text-slate-900 dark:text-slate-100">Week at a glance</h2>
    <span class="text-[11.5px] font-mono text-slate-400">focus per day</span>
  </div>

  <div class="h-[220px]">
    <canvas bind:this={canvas} class="w-full h-full"></canvas>
  </div>
</section>
