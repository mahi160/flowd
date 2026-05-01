<script lang="ts">
  import Chart from 'chart.js/auto';
  import type { Item } from '../types';
  import { fmtHM } from '../format';

  interface Props { title: string; subtitle?: string; items: Item[] }
  let { title, subtitle = '', items }: Props = $props();

  type Segment = Item & { pct: number };

  let canvas = $state<HTMLCanvasElement | undefined>(undefined);
  let chart: Chart | undefined;
  let hoveredIdx = $state<number | null>(null);

  let total    = $derived(items.reduce((n, x) => n + x.min, 0));
  let segments = $derived<Segment[]>(items.map(item => ({
    ...item,
    pct: total > 0 ? Math.round((item.min / total) * 100) : 0,
  })));
  let hovered = $derived(hoveredIdx !== null ? segments[hoveredIdx] : null);

  $effect(() => {
    if (!canvas) return;
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
        plugins: { legend: { display: false }, tooltip: { enabled: false } },
        onHover: (_, elements) => { hoveredIdx = elements[0]?.index ?? null; },
      },
    });
    return () => chart?.destroy();
  });

  function activate(i: number | null) {
    hoveredIdx = i;
    if (!chart) return;
    chart.setActiveElements(i !== null ? [{ datasetIndex: 0, index: i }] : []);
    chart.update();
  }
</script>

<section class="bg-white dark:bg-slate-800 rounded-2xl border border-slate-200 dark:border-slate-700 p-5">
  <div class="flex items-baseline gap-3 mb-3.5">
    <h2 class="font-display text-[17px] tracking-tight text-slate-900 dark:text-slate-100">{title}</h2>
    {#if subtitle}<span class="text-[11.5px] font-mono text-slate-400">{subtitle}</span>{/if}
  </div>

  {#if !items.length}
    <p class="text-[12.5px] text-slate-400">No data</p>
  {:else}
    <div class="relative mx-auto max-h-[220px] aspect-square">
      <canvas bind:this={canvas} class="w-full h-full"></canvas>
      <div class="absolute inset-0 flex flex-col items-center justify-center pointer-events-none text-center">
        {#if hovered}
          <span class="font-display text-[28px] leading-none tabular-nums text-primary">{hovered.pct}%</span>
          <span class="text-[12px] text-slate-500 mt-1 max-w-[80px] truncate">{hovered.name}</span>
          <span class="font-mono text-[11px] text-slate-400 tabular-nums">{hovered.min}m</span>
        {:else}
          <span class="font-display text-[26px] leading-none tabular-nums text-slate-900 dark:text-slate-100">{fmtHM(total)}</span>
          <span class="font-mono text-[10px] text-slate-400 uppercase tracking-widest mt-0.5">total</span>
        {/if}
      </div>
    </div>

    <ul class="mt-1 flex flex-col gap-0.5">
      {#each segments as seg, i}
        <li
          class="grid gap-2.5 px-1.5 py-1 rounded-md cursor-default transition-colors {hoveredIdx === i ? 'bg-slate-50 dark:bg-slate-700' : ''}"
          style="grid-template-columns: 10px 1fr auto auto"
          onmouseenter={() => activate(i)}
          onmouseleave={() => activate(null)}
        >
          <span class="w-2.5 h-2.5 rounded-sm self-center" style="background:{seg.color}"></span>
          <span class="font-mono text-[12px] text-slate-600 dark:text-slate-300 truncate">{seg.name}</span>
          <span class="font-mono text-[11.5px] text-slate-500 tabular-nums">{seg.min}m</span>
          <span class="font-mono text-[11px] text-slate-400 tabular-nums text-right">{seg.pct}%</span>
        </li>
      {/each}
    </ul>
  {/if}
</section>
