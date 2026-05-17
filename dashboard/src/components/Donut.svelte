<script lang="ts">
  import type { Item } from "../lib/types";
  import { arcPath, fmtHM } from "../lib/format";

  let { title, subtitle, items }: { title: string; subtitle: string; items: Item[] } = $props();
  let hover = $state<number | null>(null);

  const COLORS = ["#6366f1","#f59e0b","#ef4444","#10b981","#8b5cf6","#ec4899","#14b8a6","#f97316"];

  let computed = $derived.by(() => {
    let acc = -Math.PI / 2;
    const total = items.reduce((n, x) => n + x.min, 0);
    return {
      total,
      segs: items.map((item, i) => {
        const frac = total ? item.min / total : 0;
        const s = acc;
        acc += frac * Math.PI * 2;
        return { ...item, frac, s, e: acc, color: COLORS[i % COLORS.length] };
      }),
    };
  });

  let active = $derived(hover != null ? computed.segs[hover] : null);
</script>

<section class="rounded-xl border border-stone-200 dark:border-stone-800 bg-stone-50 dark:bg-stone-900 shadow-sm dark:shadow-stone-950/30 p-5">
  <div class="flex items-baseline gap-3 mb-4">
    <h3 class="font-display text-lg text-stone-900 dark:text-stone-100">{title}</h3>
    {#if subtitle}<span class="text-[11px] font-mono text-stone-400">{subtitle}</span>{/if}
  </div>
  {#if !items.length}
    <div class="text-stone-400 text-sm font-mono py-8 text-center">No data</div>
  {:else}
    <div class="relative aspect-square max-h-[200px] mx-auto w-full mb-4">
      <svg class="block w-full h-full" viewBox="0 0 220 220">
        {#each computed.segs as seg, i}
          {#if computed.segs.length === 1}
            <!-- Single-item: SVG arc with identical endpoints renders nothing; use two circles instead -->
            <circle
              cx="110" cy="110" r="86"
              role="img"
              aria-label="{seg.name}: {seg.min}m"
              fill={seg.color}
              opacity={hover == null || hover === 0 ? 1 : 0.25}
              onmouseenter={() => (hover = 0)}
              onmouseleave={() => (hover = null)}
            />
            <!-- Use style not fill attr so CSS dark mode class can override -->
            <circle cx="110" cy="110" r="56"
              style="fill: white"
              class="dark:!fill-stone-900"
            />
          {:else}
            <path
              d={arcPath(110, 110, 86, 56, seg.s, seg.e)}
              fill={seg.color}
              role="img"
              aria-label="{seg.name}: {seg.min}m"
              opacity={hover == null || hover === i ? 1 : 0.25}
              style:transition="opacity 0.15s"
              onmouseenter={() => (hover = i)}
              onmouseleave={() => (hover = null)}
            />
          {/if}
        {/each}
      </svg>
      <div class="absolute inset-0 flex flex-col items-center justify-center pointer-events-none text-center">
        {#if active}
          <div class="font-display text-3xl text-indigo-600 dark:text-indigo-400 tabular-nums">{Math.round(active.frac * 100)}%</div>
          <div class="text-xs mt-1 text-stone-600 dark:text-stone-300 font-medium">{active.name}</div>
          <div class="font-mono text-[11px] text-stone-400 mt-0.5 tabular-nums">{active.min}m</div>
        {:else}
          <div class="font-display text-2xl text-stone-900 dark:text-stone-100 tabular-nums">{fmtHM(computed.total)}</div>
          <div class="uppercase tracking-widest text-[10px] font-mono text-stone-400 mt-1">total</div>
        {/if}
      </div>
    </div>
    <ul class="flex flex-col gap-0.5">
      {#each computed.segs as seg, i}
        <li
          class="grid grid-cols-[10px_1fr_auto_auto] gap-2.5 items-center py-1.5 px-2 rounded-md text-xs cursor-default transition-colors {hover === i ? 'bg-stone-50 dark:bg-stone-800' : ''}"
          onmouseenter={() => (hover = i)}
          onmouseleave={() => (hover = null)}
        >
          <span class="w-2.5 h-2.5 rounded-sm" style:background={seg.color}></span>
          <span class="text-stone-600 dark:text-stone-300 font-mono text-[12px]">{seg.name}</span>
          <span class="text-stone-400 font-mono text-[11px] tabular-nums">{seg.min}m</span>
          <span class="text-stone-400 font-mono text-[10px] tabular-nums min-w-[28px] text-right">{Math.round(seg.frac * 100)}%</span>
        </li>
      {/each}
    </ul>
  {/if}
</section>
