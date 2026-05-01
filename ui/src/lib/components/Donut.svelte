<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import Chart from 'chart.js/auto';
  import type { Item } from '../types';
  import { fmtHM } from '../format';

  export let title: string;
  export let subtitle: string = '';
  export let items: Item[];

  let canvas: HTMLCanvasElement;
  let chart: Chart | undefined;
  let hoveredIdx: number | null = null;

  $: total = items.reduce((n, x) => n + x.min, 0);
  $: hovered = hoveredIdx !== null ? items[hoveredIdx] : null;

  onMount(() => {
    chart = new Chart(canvas, {
      type: 'doughnut',
      data: {
        labels: items.map(i => i.name),
        datasets: [{
          data: items.map(i => i.min),
          backgroundColor: items.map(i => i.color),
          borderWidth: 2,
          borderColor: '#fff',
          hoverBorderColor: '#fff',
          hoverOffset: 4,
        }],
      },
      options: {
        animation: false,
        cutout: '62%',
        plugins: {
          legend: { display: false },
          tooltip: { enabled: false },
        },
        onHover: (_, elements) => {
          hoveredIdx = elements[0]?.index ?? null;
        },
      },
    });
  });

  onDestroy(() => chart?.destroy());
</script>

<section class="bg-white dark:bg-slate-800 rounded-2xl border border-slate-200 dark:border-slate-700 p-5">
  <div class="flex items-baseline gap-3 mb-3.5">
    <h2 class="font-display text-[17px] tracking-tight text-slate-900 dark:text-slate-100">{title}</h2>
    {#if subtitle}<span class="text-[11.5px] font-mono text-slate-400">{subtitle}</span>{/if}
  </div>

  {#if !items.length}
    <p class="text-[12.5px] text-slate-400">No data</p>
  {:else}
    <!-- Donut + center label overlay -->
    <div class="relative mx-auto max-h-[220px] aspect-square">
      <canvas bind:this={canvas} class="w-full h-full"></canvas>
      <div class="absolute inset-0 flex flex-col items-center justify-center pointer-events-none text-center">
        {#if hovered}
          <span class="font-display text-[28px] leading-none tabular-nums text-primary">
            {Math.round((hovered.min / total) * 100)}%
          </span>
          <span class="text-[12px] text-slate-500 mt-1 max-w-[80px] truncate">{hovered.name}</span>
          <span class="font-mono text-[11px] text-slate-400 tabular-nums">{hovered.min}m</span>
        {:else}
          <span class="font-display text-[26px] leading-none tabular-nums text-slate-900 dark:text-slate-100">{fmtHM(total)}</span>
          <span class="font-mono text-[10px] text-slate-400 uppercase tracking-widest mt-0.5">total</span>
        {/if}
      </div>
    </div>

    <!-- Legend -->
    <ul class="mt-1 flex flex-col gap-0.5">
      {#each items as item, i}
        <li
          class="grid gap-2.5 px-1.5 py-1 rounded-md text-[12.5px] cursor-default transition-colors"
          class:bg-slate-50={hoveredIdx === i}
          class:dark:bg-slate-700={hoveredIdx === i}
          style="grid-template-columns: 10px 1fr auto auto"
          on:mouseenter={() => { hoveredIdx = i; chart?.setActiveElements([{datasetIndex:0,index:i}]); chart?.update(); }}
          on:mouseleave={() => { hoveredIdx = null; chart?.setActiveElements([]); chart?.update(); }}
        >
          <span class="w-2.5 h-2.5 rounded-sm self-center" style="background:{item.color}"></span>
          <span class="font-mono text-[12px] text-slate-600 dark:text-slate-300 truncate">{item.name}</span>
          <span class="font-mono text-[11.5px] text-slate-500 tabular-nums">{item.min}m</span>
          <span class="font-mono text-[11px] text-slate-400 tabular-nums text-right">{Math.round((item.min / total) * 100)}%</span>
        </li>
      {/each}
    </ul>
  {/if}
</section>
