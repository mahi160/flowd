<script lang="ts">
  import { transform } from './lib/data';
  import { THEMES, applyTheme, savedTheme, type Theme } from './lib/theme';

  import Header from './lib/components/Header.svelte';
  import HeroStrip from './lib/components/HeroStrip.svelte';
  import ActivityHeatmap from './lib/components/ActivityHeatmap.svelte';
  import WeekHeatmap from './lib/components/WeekHeatmap.svelte';
  import Insights from './lib/components/Insights.svelte';
  import Donut from './lib/components/Donut.svelte';
  import Languages from './lib/components/Languages.svelte';
  import HourPattern from './lib/components/HourPattern.svelte';
  import StreakCard from './lib/components/StreakCard.svelte';
  import Timeline from './lib/components/Timeline.svelte';
  import Summary from './lib/components/Summary.svelte';
  import WeekBars from './lib/components/WeekBars.svelte';
  import ProjectBreakdown from './lib/components/ProjectBreakdown.svelte';
  import FlowMark from './lib/icons/FlowMark.svelte';

  const data = transform(window.__FLOWD_DATA__ ?? {});

  let view  = $state<'today' | 'week'>(data.period);
  let theme = $state<Theme>(savedTheme());

  $effect(() => {
    applyTheme(theme);
    const mq = matchMedia('(prefers-color-scheme: dark)');
    const onChange = () => { if (theme === 'system') applyTheme(theme); };
    mq.addEventListener('change', onChange);
    return () => mq.removeEventListener('change', onChange);
  });

  function cycleTheme() {
    theme = THEMES[(THEMES.indexOf(theme) + 1) % THEMES.length];
  }
</script>

<div class="max-w-[1480px] mx-auto px-9 py-8 pb-20">
  <Header {data} {view} setView={(v) => (view = v)} {theme} {cycleTheme} />

  {#if !data.hasData}
    <div class="bg-white dark:bg-slate-800 rounded-2xl border border-slate-200 dark:border-slate-700 p-8 flex flex-col items-center justify-center min-h-[280px] gap-3 text-center">
      <FlowMark size={48} />
      <h2 class="font-display text-[22px] text-slate-600 dark:text-slate-300 m-0">No activity yet</h2>
      <p class="text-slate-400 text-[12.5px] m-0">Start the daemon to begin tracking.</p>
      <p class="text-slate-400 text-[12.5px] m-0">
        <code class="bg-slate-100 dark:bg-slate-700 px-2 py-0.5 rounded font-mono text-[12px]">fw start</code>
      </p>
    </div>

  {:else if view !== data.period}
    <div class="bg-white dark:bg-slate-800 rounded-2xl border border-slate-200 dark:border-slate-700 p-8 flex flex-col items-center justify-center min-h-[280px] gap-3 text-center">
      <h2 class="font-display text-[22px] text-slate-600 dark:text-slate-300 m-0">
        {view[0].toUpperCase() + view.slice(1)} view
      </h2>
      <p class="text-slate-400 text-[12.5px] m-0">This dashboard was generated for the {data.period} period.</p>
      <p class="text-slate-400 text-[12.5px] m-0">
        <code class="bg-slate-100 dark:bg-slate-700 px-2 py-0.5 rounded font-mono text-[12px]">fw dashboard {view}</code>
      </p>
    </div>

  {:else if view === 'week'}
    <main class="flex flex-col gap-[18px]">
      <HeroStrip {data} label="this week" />
      {#if data.weekDays.length > 0}<WeekBars {data} />{/if}
      <WeekHeatmap {data} />
      <div class="grid grid-cols-3 gap-3.5">
        <Donut title="By project" subtitle="this week" items={data.byProject} />
        <Donut title="By command" subtitle="this week" items={data.byCommand.slice(0, 8)} />
        <Languages {data} />
      </div>
      <div class="grid gap-3.5 grid-cols-[1.5fr_1fr]">
        <Timeline {data} />
        <Summary {data} />
      </div>
    </main>

  {:else}
    <main class="flex flex-col gap-[18px]">
      <HeroStrip {data} label="today" />
      <div class="grid gap-3.5 grid-cols-[1.5fr_1fr]">
        <ActivityHeatmap {data} />
        <Insights {data} />
      </div>
      <div class="grid grid-cols-3 gap-3.5">
        <Donut title="By project" subtitle="today" items={data.byProject} />
        <Donut title="By command" subtitle="today · top tools" items={data.byCommand.slice(0, 8)} />
        <Languages {data} />
      </div>
      <div class="grid grid-cols-3 gap-3.5">
        <HourPattern {data} />
        <StreakCard {data} />
        <Timeline {data} />
      </div>
      <div class="grid gap-3.5 grid-cols-[1.5fr_1fr]">
        <Summary {data} />
        <ProjectBreakdown {data} />
      </div>
    </main>
  {/if}

  <footer class="flex justify-center items-center gap-1.5 mt-7 font-mono text-[11px] text-slate-400">
    flowd<span class="text-slate-300 dark:text-slate-600"> — local activity tracker · self-hosted</span>
  </footer>
</div>
