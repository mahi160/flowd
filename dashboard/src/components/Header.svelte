<script lang="ts">
  import type { Data } from "../lib/transform";
  import type { Theme } from "../lib/theme";

  let {
    data, view, theme, onSetView, onCycleTheme,
  }: {
    data: Data; view: string; theme: Theme;
    onSetView: (v: string) => void; onCycleTheme: () => void;
  } = $props();

  const themeIcon: Record<string, string> = { dark: "🌙", light: "☀️", system: "💻" };
</script>

<header class="flex items-center justify-between pb-6 mb-2 border-b border-stone-200 dark:border-stone-800">
  <div class="flex items-center gap-3">
    <div class="w-9 h-9 rounded-lg bg-indigo-600 flex items-center justify-center">
      <span class="text-stone-100 font-medium text-sm">fw</span>
    </div>
    <div>
      <div class="font-display text-2xl leading-none tracking-tight text-stone-900 dark:text-stone-100">flowd</div>
      <div class="flex items-center gap-2 text-xs text-stone-500 mt-1">
        <span class="uppercase tracking-widest text-[10px] font-mono text-stone-400">Generated</span>
        <span class="font-mono text-[11px] tabular-nums">{data.generated}</span>
        {#if data.hasData}
          <span class="text-stone-300 dark:text-stone-600">·</span>
          <span class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-[11px] font-mono bg-indigo-50 text-indigo-700 dark:bg-indigo-950 dark:text-indigo-300 ring-1 ring-indigo-200 dark:ring-indigo-800">
            <span class="w-1.5 h-1.5 rounded-full bg-green-500"></span>
            {data.period}
          </span>
        {/if}
      </div>
    </div>
  </div>
  <div class="flex items-center gap-2">
    <div class="inline-flex rounded-lg bg-stone-100 dark:bg-stone-800 p-0.5">
      {#each ["today", "week"] as v}
        <button
          class="px-3.5 py-1.5 text-xs font-medium rounded-md transition-colors {v === view ? 'bg-stone-50 dark:bg-stone-700 text-stone-900 dark:text-stone-100 shadow-sm' : 'text-stone-500 hover:text-stone-700 dark:hover:text-stone-300'}"
          onclick={() => onSetView(v)}
        >
          {v === "today" ? "Today" : "Week"}
        </button>
      {/each}
    </div>
    <button
      class="w-8 h-8 rounded-lg border border-stone-200 dark:border-stone-700 flex items-center justify-center text-sm hover:bg-stone-50 dark:hover:bg-stone-800 transition-colors"
      onclick={onCycleTheme}
    >{themeIcon[theme]}</button>
  </div>
</header>
