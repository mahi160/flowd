<script lang="ts">
  import type { RawPayload } from "./lib/types";
  import { transform } from "./lib/transform";
  import { loadTheme, cycleTheme, resolveTheme, saveTheme } from "./lib/theme";
  import Header      from "./components/Header.svelte";
  import HeroStrip   from "./components/HeroStrip.svelte";
  import Heatmap     from "./components/Heatmap.svelte";
  import WeekHeatmap from "./components/WeekHeatmap.svelte";
  import CalHeatmap  from "./components/CalHeatmap.svelte";
  import AllHeatmap  from "./components/AllHeatmap.svelte";
  import Donut       from "./components/Donut.svelte";
  import Languages   from "./components/Languages.svelte";
  import Streak      from "./components/Streak.svelte";
  import AISummary   from "./components/AISummary.svelte";
  import Timeline    from "./components/Timeline.svelte";
  import WeekBars    from "./components/WeekBars.svelte";
  import Standup     from "./components/Standup.svelte";

  let { raw }: { raw: RawPayload } = $props();

  let period = $state(raw.initial_period || "today");
  let theme  = $state(loadTheme());

  let data = $derived(transform(raw, period));
  // Human label for the selected period (currently the period id itself).
  let periodLabel = $derived(period);

  $effect(() => {
    const resolved = resolveTheme(theme);
    document.documentElement.dataset.theme = resolved;
    document.documentElement.classList.toggle("dark",  resolved === "dark");
    document.documentElement.classList.toggle("light", resolved === "light");
    saveTheme(theme);
  });

  // Which heatmap component to render for each period.
  // "today"     → Heatmap (48 half-hour buckets)
  // "week"      → WeekBars + WeekHeatmap (7-day hourly grid)
  // "month"/"year" → CalHeatmap (calendar grid)
  // "all"       → AllHeatmap (month bars)
  const HEATMAP_TYPE: Record<string, "today" | "week" | "cal" | "all"> = {
    today:     "today",
    yesterday: "today",
    week:      "week",
    month:     "cal",
    year:      "cal",
    all:       "all",
  };
  let heatmapType = $derived(HEATMAP_TYPE[period] ?? "today");
  let showStandup = $derived(period === "today" && !!(data.standup));
</script>

<div class="min-h-screen max-w-[1480px] mx-auto px-6 py-8 md:px-10">
  <Header
    {data}
    view={period}
    {theme}
    onSetView={(v) => (period = v)}
    onCycleTheme={() => (theme = cycleTheme(theme))}
  />

  {#if !data.anyData}
    <!-- ── Global empty state ─────────────────────────────────────────── -->
    <main>
      <div class="rounded-xl border border-stone-200 dark:border-stone-800
                  bg-stone-50 dark:bg-stone-900 shadow-sm flex flex-col
                  items-center justify-center min-h-[280px] text-center gap-3 p-8">
        <h2 class="font-display text-xl text-stone-500">No activity yet</h2>
        <p class="text-stone-400 text-xs">Start the daemon to begin tracking.</p>
        <code class="bg-stone-100 dark:bg-stone-800 px-2 py-0.5 rounded font-mono text-xs">fw start</code>
      </div>
    </main>

  {:else if !data.hasData}
    <!-- ── Period-specific empty state ───────────────────────────────── -->
    <main>
      <div class="rounded-xl border border-stone-200 dark:border-stone-800
                  bg-stone-50 dark:bg-stone-900 shadow-sm flex flex-col
                  items-center justify-center min-h-[280px] text-center gap-2.5 p-8">
        <h2 class="font-display text-xl text-stone-500">No {period} activity</h2>
        <p class="text-stone-400 text-xs">Nothing tracked for this period yet.</p>
      </div>
    </main>

  {:else}
    <!-- ── Single unified layout across all periods ───────────────────── -->
    <main class="flex flex-col gap-5">

      <!-- Standup (today only, when AI or raw data is available) -->
      {#if showStandup}
        <Standup {data} />
      {/if}

      <!-- ── Heatmap centrepiece ───────────────────────────────────────── -->
      {#if heatmapType === "today"}
        <Heatmap {data} />
      {:else if heatmapType === "week"}
        {#if data.weekDays.length > 0}<WeekBars {data} />{/if}
        <WeekHeatmap {data} />
      {:else if heatmapType === "cal"}
        <CalHeatmap {data} />
      {:else}
        <AllHeatmap {data} />
      {/if}

      <!-- ── Hero numbers ──────────────────────────────────────────────── -->
      <HeroStrip {data} label={periodLabel} />

      <!-- ── Breakdowns ────────────────────────────────────────────────── -->
      <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-5">
        <Donut
          title="By project"
          subtitle="{periodLabel} · top projects"
          items={data.byProject}
        />
        <Donut
          title="By command"
          subtitle="{periodLabel} · top tools"
          items={data.byCommand.slice(0, 8)}
        />
        <Languages {data} />
      </div>

      <!-- ── AI sessions ───────────────────────────────────────────────── -->
      {#if data.aiTools.length > 0}
        <AISummary {data} />
      {/if}

      <!-- ── Timeline + Streak ─────────────────────────────────────────── -->
      <div class="grid grid-cols-1 lg:grid-cols-[1.5fr_1fr] gap-5">
        <Timeline {data} />
        <Streak {data} />
      </div>

    </main>
  {/if}

  <footer class="flex justify-center items-center gap-1.5 mt-8 text-[11px] text-stone-400 font-mono">
    flowd
    <span class="text-stone-300 dark:text-stone-600">— local activity tracker</span>
  </footer>
</div>
