<script lang="ts">
  import Header from "./lib/components/Header.svelte";
  import FocusSummary from "./lib/widgets/FocusSummary.svelte";
  import AiSessionSummary from "./lib/widgets/AiSessionSummary.svelte";
  import CodeSummary from "./lib/widgets/CodeSummary.svelte";
  import MachineAndProjectSummary from "./lib/widgets/MachineAndProjectSummary.svelte";
  import Heatmap from "./lib/widgets/Heatmap.svelte";
  import Summary from "./lib/widgets/Summary.svelte";
  import Donut from "./lib/widgets/Donut.svelte";
  import ByTool from "./lib/widgets/ByTool.svelte";
  import Languages from "./lib/widgets/Languages.svelte";
  import AIUsage from "./lib/widgets/AIUsage.svelte";
  import Goals from "./lib/widgets/Goals.svelte";
  import Comparison from "./lib/widgets/Comparison.svelte";
  import BestHours from "./lib/widgets/BestHours.svelte";
  import Streak from "./lib/widgets/Streak.svelte";
  import HotFiles from "./lib/widgets/HotFiles.svelte";
  import Timeline from "./lib/widgets/Timeline.svelte";
  import TmuxSessions from "./lib/widgets/TmuxSessions.svelte";
  import Distractions from "./lib/widgets/Distractions.svelte";
  import WeekBars from "./lib/widgets/WeekBars.svelte";
  import MonthCalendar from "./lib/widgets/MonthCalendar.svelte";
</script>

<div class="container mx-auto p-4 min-h-screen bg-background text-foreground">
  <Header />

  <main class="space-y-4 mt-6">
    <!-- Hero strip -->
    <div class="grid grid-cols-1 gap-3 md:grid-cols-[1.5fr_1fr_1fr_1fr]">
      <FocusSummary />
      <AiSessionSummary />
      <CodeSummary />
      <MachineAndProjectSummary />
    </div>

    <!-- Heatmap + AI summary -->
    <div class="grid grid-cols-1 gap-3 lg:grid-cols-[1.5fr_1fr]">
      <Heatmap />
      <Summary />
    </div>

    <!-- Distributions -->
    <div class="grid grid-cols-1 gap-3 lg:grid-cols-3">
      <Donut />
      <ByTool />
      <Languages />
    </div>

    <!-- AI usage + Goals/Comparison stack -->
    <div class="grid grid-cols-1 gap-3 lg:grid-cols-[1.6fr_1fr]">
      <AIUsage />
      <div class="flex flex-col gap-3">
        <Goals />
        <Comparison />
      </div>
    </div>

    <!-- Best hours + Streak + Hot paths -->
    <div class="grid grid-cols-1 gap-3 lg:grid-cols-3">
      <BestHours />
      <Streak />
      <HotFiles />
    </div>

    <!-- Timeline + tmux/distractions stack -->
    <div class="grid grid-cols-1 gap-3 lg:grid-cols-[1.6fr_1fr]">
      <Timeline />
      <div class="flex flex-col gap-3">
        <TmuxSessions />
        <Distractions />
      </div>
    </div>

    <!-- Week + Month at end -->
    <WeekBars />
    <MonthCalendar />
  </main>

  <footer
    class="text-foreground/50 text-[10px] text-center mt-12 font-mono"
  >
    flowd — self-hosted activity tracker · ~/.flowd/today.json
  </footer>
</div>
