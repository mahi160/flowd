<script lang="ts">
  import type { Data } from "../lib/transform";
  import Sparkline from "./Sparkline.svelte";
  import Tooltip from "./Tooltip.svelte";

  let { data, label }: { data: Data; label: string } = $props();
  let f = $derived(data.focus);
  let c = $derived(data.code);
  let top = $derived(data.byProject[0]);

  const PERIOD_TIPS: Record<string, string> = {
    today: "today so far",
    week:  "rolling 7-day window",
    month: "current calendar month",
    year:  "current calendar year",
    all:   "all recorded history",
  };
  let periodTip = $derived(PERIOD_TIPS[data.period] ?? data.period);
</script>

<section class="grid grid-cols-[1.5fr_1fr_1fr_1fr] gap-4 max-lg:grid-cols-[1.4fr_1fr_1fr] max-md:grid-cols-2">
  <!-- Focus -->
  <div class="rounded-xl border border-stone-200 dark:border-stone-800 bg-stone-50 dark:bg-stone-900 shadow-sm p-5 flex flex-col gap-1.5 min-h-[140px]">
    <Tooltip text="Total time you were actively in a tracked tmux pane ({periodTip}). Measured in real focused minutes — idle time and detached sessions are excluded.">
      <div class="uppercase tracking-widest text-[10px] font-mono text-stone-400">Focus {label}</div>
    </Tooltip>
    <div class="font-display text-5xl tabular-nums text-indigo-600 dark:text-indigo-400 leading-none mt-1">
      {Math.floor(f.totalMin / 60)}<span class="text-xl text-indigo-400 dark:text-indigo-500 mx-0.5">h</span>{f.totalMin % 60}<span class="text-xl text-indigo-400 dark:text-indigo-500 mx-0.5">m</span>
    </div>
    <div class="text-xs text-stone-500 mt-1">
      <Tooltip text="A focus block is a window of time with at least {data.period === 'today' ? '30' : '30+'} recorded focused minutes. Each block triggers a git commit to your journal.">
        <span class="font-mono tabular-nums font-medium text-stone-700 dark:text-stone-300">{f.blocks}</span>
        <span> blocks</span>
      </Tooltip>
      <span class="text-stone-300 dark:text-stone-600 mx-1">·</span>
      <Tooltip text="A context switch happens when the active tmux session changes, indicating a project or task switch.">
        <span class="font-mono tabular-nums font-medium text-stone-700 dark:text-stone-300">{f.switches}</span>
        <span> switches</span>
      </Tooltip>
    </div>
    <Sparkline data={data.hourly} />
  </div>

  <!-- Machine -->
  <div class="rounded-xl border border-stone-200 dark:border-stone-800 bg-stone-50 dark:bg-stone-900 shadow-sm p-5 flex flex-col gap-1.5 min-h-[140px]">
    <Tooltip text="The machine name from your flowd config (machine_name). Used to separate data when tracking on multiple devices.">
      <div class="uppercase tracking-widest text-[10px] font-mono text-stone-400">Machine</div>
    </Tooltip>
    <div class="font-display text-2xl text-stone-900 dark:text-stone-100 mt-1">{data.machine}</div>
    <div class="text-xs text-stone-500 font-mono">{data.os}</div>
    <div class="mt-auto h-1 bg-stone-100 dark:bg-stone-800 rounded-full overflow-hidden">
      <span class="block h-full w-1/2 rounded-full bg-gradient-to-r from-indigo-500 to-purple-500"></span>
    </div>
  </div>

  <!-- Top repo -->
  <div class="rounded-xl border border-stone-200 dark:border-stone-800 bg-stone-50 dark:bg-stone-900 shadow-sm p-5 flex flex-col gap-1.5 min-h-[140px]">
    <Tooltip text="The git repository you spent the most focused time in during this period. Detected automatically from your working directory.">
      <div class="uppercase tracking-widest text-[10px] font-mono text-stone-400">Top repo</div>
    </Tooltip>
    <div class="font-display text-2xl text-stone-900 dark:text-stone-100 mt-1">{data.topRepo.name}</div>
    {#if data.topRepo.branch}
      <span class="inline-flex items-center gap-1 w-fit px-2 py-0.5 rounded bg-indigo-50 dark:bg-indigo-950 text-indigo-700 dark:text-indigo-300 text-[11px] font-mono ring-1 ring-indigo-100 dark:ring-indigo-900">⎇ {data.topRepo.branch}</span>
    {/if}
    {#if top}
      <div class="mt-auto font-mono text-[11px] text-stone-400 tabular-nums">{top.name} · {top.min}m</div>
    {/if}
  </div>

  <!-- Code -->
  <div class="rounded-xl border border-stone-200 dark:border-stone-800 bg-stone-50 dark:bg-stone-900 shadow-sm p-5 flex flex-col gap-1.5 min-h-[140px] max-lg:hidden max-md:flex">
    <Tooltip text="Lines of code added and removed across all tracked git repos, based on committed changes within this period. Uncommitted changes are not counted to avoid double-counting.">
      <div class="uppercase tracking-widest text-[10px] font-mono text-stone-400">Code</div>
    </Tooltip>
    <div class="font-display text-2xl text-stone-900 dark:text-stone-100 tabular-nums mt-1">{c.files} <span class="text-stone-400 text-lg">files</span></div>
    <div class="flex gap-2 items-center text-base font-mono tabular-nums mt-1">
      <span class="text-green-600 dark:text-green-400 font-medium">+{c.added}</span>
      <span class="text-stone-300 dark:text-stone-600">/</span>
      <span class="text-red-500 dark:text-red-400 font-medium">−{c.removed}</span>
    </div>
  </div>
</section>
