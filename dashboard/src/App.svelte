<script lang="ts">
  import type { Data } from "./lib/transform";
  import { loadTheme, cycleTheme, resolveTheme, saveTheme } from "./lib/theme";
  import Header from "./components/Header.svelte";
  import HeroStrip from "./components/HeroStrip.svelte";
  import Heatmap from "./components/Heatmap.svelte";
  import WeekHeatmap from "./components/WeekHeatmap.svelte";
  import CalHeatmap from "./components/CalHeatmap.svelte";
  import AllHeatmap from "./components/AllHeatmap.svelte";
  import Donut from "./components/Donut.svelte";
  import Languages from "./components/Languages.svelte";
  import Streak from "./components/Streak.svelte";
  import AISummary from "./components/AISummary.svelte";
  import Timeline from "./components/Timeline.svelte";
  import WeekBars from "./components/WeekBars.svelte";

  let { data }: { data: Data } = $props();

  const initialPeriod = data.period;
  let view = $state(initialPeriod);
  let theme = $state(loadTheme());

  $effect(() => {
    const resolved = resolveTheme(theme);
    document.documentElement.dataset.theme = resolved;
    document.documentElement.classList.toggle("dark", resolved === "dark");
    document.documentElement.classList.toggle("light", resolved === "light");
    saveTheme(theme);
  });

  function onCycleTheme() {
    theme = cycleTheme(theme);
  }

  const PERIOD_LABELS: Record<string, string> = {
    today: "today",
    week: "this week",
    month: "this month",
    year: "this year",
    all: "all time",
  };
</script>

<div class="min-h-screen max-w-[1480px] mx-auto px-6 py-8 md:px-10">
  <Header {data} {view} {theme} onSetView={(v) => (view = v)} {onCycleTheme} />

  {#if !data.hasData}
    <!-- ── Empty state ─────────────────────────────────────────── -->
    <main class="flex flex-col gap-5">
      <div class="rounded-xl border border-stone-200 dark:border-stone-800 bg-stone-50 dark:bg-stone-900
                  shadow-sm flex flex-col items-center justify-center min-h-[280px] text-center gap-2.5 p-8">
        <h2 class="font-display text-xl text-stone-500">No activity yet</h2>
        <p class="text-stone-400 text-xs">Start the daemon to begin tracking.</p>
        <code class="bg-stone-100 dark:bg-stone-800 px-2 py-0.5 rounded font-mono text-xs">fw start</code>
      </div>
    </main>

  {:else if view !== data.period}
    <!-- ── Wrong period: prompt user to regenerate ─────────────── -->
    <main class="flex flex-col gap-5">
      <div class="rounded-xl border border-stone-200 dark:border-stone-800 bg-stone-50 dark:bg-stone-900
                  shadow-sm flex flex-col items-center justify-center min-h-[280px] text-center gap-3 p-8">
        <h2 class="font-display text-xl text-stone-500">
          {PERIOD_LABELS[view] ?? view} view
        </h2>
        <p class="text-stone-400 text-xs max-w-xs">
          This dashboard was generated for <strong class="text-stone-600 dark:text-stone-300">{data.period}</strong>.
          To see {PERIOD_LABELS[view] ?? view} data, regenerate:
        </p>
        <code class="bg-stone-100 dark:bg-stone-800 px-3 py-1 rounded font-mono text-xs">
          fw dashboard {view}
        </code>
      </div>
    </main>

  {:else if view === "today"}
    <!-- ── Today ───────────────────────────────────────────────── -->
    <main class="flex flex-col gap-5">
      <HeroStrip {data} label="today" />
      <div class="grid grid-cols-1 lg:grid-cols-[1.5fr_1fr] gap-5">
        <Heatmap {data} />
        <Streak {data} />
      </div>
      <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-5">
        <Donut title="By project" subtitle="today" items={data.byProject} />
        <Donut title="By command" subtitle="today · top tools" items={data.byCommand.slice(0, 8)} />
        <Languages {data} />
      </div>
      {#if data.aiTools.length > 0}<AISummary {data} />{/if}
      <Timeline {data} />
    </main>

  {:else if view === "week"}
    <!-- ── Week ────────────────────────────────────────────────── -->
    <main class="flex flex-col gap-5">
      <HeroStrip {data} label="this week" />
      {#if data.weekDays.length > 0}<WeekBars {data} />{/if}
      <WeekHeatmap {data} />
      <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-5">
        <Donut title="By project" subtitle="this week" items={data.byProject} />
        <Donut title="By command" subtitle="this week" items={data.byCommand.slice(0, 8)} />
        <Languages {data} />
      </div>
      {#if data.aiTools.length > 0}<AISummary {data} />{/if}
      <div class="grid grid-cols-1 lg:grid-cols-[1.5fr_1fr] gap-5">
        <Timeline {data} />
        <Streak {data} />
      </div>
    </main>

  {:else if view === "month"}
    <!-- ── Month ───────────────────────────────────────────────── -->
    <main class="flex flex-col gap-5">
      <HeroStrip {data} label="this month" />
      <CalHeatmap {data} />
      <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-5">
        <Donut title="By project" subtitle="this month" items={data.byProject} />
        <Donut title="By command" subtitle="this month" items={data.byCommand.slice(0, 8)} />
        <Languages {data} />
      </div>
      {#if data.aiTools.length > 0}<AISummary {data} />{/if}
      <div class="grid grid-cols-1 lg:grid-cols-[1.5fr_1fr] gap-5">
        <Timeline {data} />
        <Streak {data} />
      </div>
    </main>

  {:else if view === "year"}
    <!-- ── Year ────────────────────────────────────────────────── -->
    <main class="flex flex-col gap-5">
      <HeroStrip {data} label="this year" />
      <CalHeatmap {data} />
      <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-5">
        <Donut title="By project" subtitle="this year" items={data.byProject} />
        <Donut title="By command" subtitle="this year" items={data.byCommand.slice(0, 8)} />
        <Languages {data} />
      </div>
      {#if data.aiTools.length > 0}<AISummary {data} />{/if}
      <div class="grid grid-cols-1 lg:grid-cols-[1.5fr_1fr] gap-5">
        <Timeline {data} />
        <Streak {data} />
      </div>
    </main>

  {:else if view === "all"}
    <!-- ── All time ─────────────────────────────────────────────── -->
    <main class="flex flex-col gap-5">
      <HeroStrip {data} label="all time" />
      <AllHeatmap {data} />
      <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-5">
        <Donut title="By project" subtitle="all time" items={data.byProject} />
        <Donut title="By command" subtitle="all time" items={data.byCommand.slice(0, 8)} />
        <Languages {data} />
      </div>
      {#if data.aiTools.length > 0}<AISummary {data} />{/if}
      <div class="grid grid-cols-1 lg:grid-cols-[1.5fr_1fr] gap-5">
        <Timeline {data} />
        <Streak {data} />
      </div>
    </main>
  {/if}

  <footer class="flex justify-center items-center gap-1.5 mt-8 text-[11px] text-stone-400 font-mono">
    flowd <span class="text-stone-300 dark:text-stone-600">— local activity tracker</span>
  </footer>
</div>
