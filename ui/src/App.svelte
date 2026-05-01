<script lang="ts">
  import { onMount } from "svelte";
  import { transform } from "./lib/data";
  import type { ParsedData } from "./lib/types";

  import Header from "./lib/components/Header.svelte";
  import HeroStrip from "./lib/components/HeroStrip.svelte";
  import ActivityHeatmap from "./lib/components/ActivityHeatmap.svelte";
  import WeekHeatmap from "./lib/components/WeekHeatmap.svelte";
  import Insights from "./lib/components/Insights.svelte";
  import Donut from "./lib/components/Donut.svelte";
  import Languages from "./lib/components/Languages.svelte";
  import HourPattern from "./lib/components/HourPattern.svelte";
  import StreakCard from "./lib/components/StreakCard.svelte";
  import Timeline from "./lib/components/Timeline.svelte";
  import Summary from "./lib/components/Summary.svelte";
  import WeekBars from "./lib/components/WeekBars.svelte";
  import FlowMark from "./lib/icons/FlowMark.svelte";

  type Theme = "dark" | "light" | "system";
  const THEMES: Theme[] = ["dark", "light", "system"];

  const data: ParsedData = transform((window.__FLOWD_DATA__ as any) ?? {});

  let view: string = data.period;
  let theme: Theme = (localStorage.getItem("fw-theme") as Theme) ?? "system";

  function resolveTheme(t: Theme): "dark" | "light" {
    return t === "system"
      ? matchMedia("(prefers-color-scheme: dark)").matches
        ? "dark"
        : "light"
      : t;
  }

  function applyTheme() {
    document.documentElement.dataset.theme = resolveTheme(theme);
    localStorage.setItem("fw-theme", theme);
  }

  function cycleTheme() {
    theme = THEMES[(THEMES.indexOf(theme) + 1) % THEMES.length];
    applyTheme();
  }

  function setView(v: string) {
    view = v;
  }

  onMount(() => {
    applyTheme();
    const mq = matchMedia("(prefers-color-scheme: dark)");
    const onChange = () => {
      if (theme === "system") applyTheme();
    };
    mq.addEventListener("change", onChange);
    return () => mq.removeEventListener("change", onChange);
  });

  $: headerData = {
    generated: data.generated,
    hasData: data.hasData,
    period: data.period,
  };
</script>

<div class="max-w-[1480px] mx-auto px-9 py-8 pb-20">
  <Header data={headerData} {view} {setView} {theme} {cycleTheme} />

  {#if !data.hasData}
    <!-- Empty state -->
    <div
      class="bg-white dark:bg-slate-800 rounded-2xl border border-slate-200 dark:border-slate-700 p-8 flex flex-col items-center justify-center min-h-[280px] gap-3 text-center"
    >
      <FlowMark size={48} />
      <h2
        class="font-display text-[22px] text-slate-600 dark:text-slate-300 m-0"
      >
        No activity yet
      </h2>
      <p class="text-slate-400 text-[12.5px] m-0">
        Start the daemon to begin tracking.
      </p>
      <p class="text-slate-400 text-[12.5px] m-0">
        <code
          class="bg-slate-100 dark:bg-slate-700 px-2 py-0.5 rounded font-mono text-[12px]"
          >fw start</code
        >
      </p>
    </div>
  {:else if view !== data.period}
    <!-- Wrong period -->
    <div
      class="bg-white dark:bg-slate-800 rounded-2xl border border-slate-200 dark:border-slate-700 p-8 flex flex-col items-center justify-center min-h-[280px] gap-3 text-center"
    >
      <h2
        class="font-display text-[22px] text-slate-600 dark:text-slate-300 m-0"
      >
        {view[0].toUpperCase() + view.slice(1)} view
      </h2>
      <p class="text-slate-400 text-[12.5px] m-0">
        This dashboard was generated for the {data.period} period.
      </p>
      <p class="text-slate-400 text-[12.5px] m-0">
        <code
          class="bg-slate-100 dark:bg-slate-700 px-2 py-0.5 rounded font-mono text-[12px]"
          >fw dashboard {view}</code
        >
      </p>
    </div>
  {:else if view === "week"}
    <!-- Week view -->
    <main class="flex flex-col gap-[18px]">
      <HeroStrip {data} label="this week" />
      {#if data.weekDays.length > 0}
        <WeekBars {data} />
      {/if}
      <WeekHeatmap {data} />
      <div class="grid grid-cols-3 gap-3.5">
        <Donut title="By project" subtitle="this week" items={data.byProject} />
        <Donut
          title="By command"
          subtitle="this week"
          items={data.byCommand.slice(0, 8)}
        />
        <Languages {data} />
      </div>
      <div class="grid gap-3.5 grid-cols-[1.5fr_1fr]">
        <Timeline {data} />
        <Summary {data} />
      </div>
    </main>
  {:else}
    <!-- Today view -->
    <main class="flex flex-col gap-[18px]">
      <HeroStrip {data} label="today" />
      <div class="grid gap-3.5 grid-cols-[1.5fr_1fr]">
        <ActivityHeatmap {data} />
        <Insights {data} />
      </div>
      <div class="grid grid-cols-3 gap-3.5">
        <Donut title="By project" subtitle="today" items={data.byProject} />
        <Donut
          title="By command"
          subtitle="today · top tools"
          items={data.byCommand.slice(0, 8)}
        />
        <Languages {data} />
      </div>
      <div class="grid grid-cols-3 gap-3.5">
        <HourPattern {data} />
        <StreakCard {data} />
        <Timeline {data} />
      </div>
      <div class="grid gap-3.5 grid-cols-[1.5fr_1fr]">
        <Summary {data} />
        {#if data.byProject.length > 0}
          {@const topMin = data.byProject[0].min}
          <div
            class="bg-white dark:bg-slate-800 rounded-2xl border border-slate-200 dark:border-slate-700 p-5"
          >
            <!-- Project breakdown inline -->
            <div class="flex items-baseline gap-3 mb-3.5">
              <h2
                class="font-display text-[17px] tracking-tight text-slate-900 dark:text-slate-100"
              >
                By project
              </h2>
              <span class="text-[11.5px] font-mono text-slate-400"
                >time breakdown</span
              >
            </div>
            <ul class="flex flex-col gap-2">
              {#each data.byProject.slice(0, 8) as p}
                <li
                  class="grid items-center gap-2.5"
                  style="grid-template-columns: 10px 1fr 1fr auto"
                >
                  <span
                    class="w-2.5 h-2.5 rounded-sm"
                    style="background:{p.color}"
                  ></span>
                  <span
                    class="font-mono text-[12px] text-slate-600 dark:text-slate-300 truncate"
                    >{p.name}</span
                  >
                  <div
                    class="h-1 bg-slate-100 dark:bg-slate-700 rounded-full overflow-hidden"
                  >
                    <span
                      class="block h-full rounded-full"
                      style="width:{(p.min / topMin) *
                        100}%; background:{p.color}"
                    ></span>
                  </div>
                  <span
                    class="font-mono text-[11.5px] text-slate-500 tabular-nums"
                    >{p.min}m</span
                  >
                </li>
              {/each}
            </ul>
          </div>
        {/if}
      </div>
    </main>
  {/if}

  <footer
    class="flex justify-center items-center gap-1.5 mt-7 font-mono text-[11px] text-slate-400"
  >
    flowd<span class="text-slate-300 dark:text-slate-600">
      — local activity tracker · self-hosted</span
    >
  </footer>
</div>
