<script>
  import { flowd } from "./lib/store.svelte";

  import ActivityHeatmap from "./lib/widgets/ActivityHeatmap.svelte";
  import Insights from "./lib/widgets/Insights.svelte";
  import Breakdown from "./lib/widgets/Breakdown.svelte";
  import Timeline from "./lib/widgets/Timeline.svelte";
  import Header from "./lib/components/Header.svelte";
  import FocusSummary from "./lib/widgets/FocusSummary.svelte";
  import MachineAndProjectSummary from "./lib/widgets/MachineAndProjectSummary.svelte";
  import AiSessionSummary from "./lib/widgets/AiSessionSummary.svelte";
  import CodeSummary from "./lib/widgets/CodeSummary.svelte";
</script>

<div class="container mx-auto p-4 min-h-screen bg-background text-foreground">
  <Header />
  <main class="space-y-6 mt-6">
    <div class="grid grid-cols-1 gap-4 md:grid-cols-[30%_1fr_1fr_1fr]">
      <FocusSummary />
      <MachineAndProjectSummary />
      <AiSessionSummary />
      <CodeSummary />
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <ActivityHeatmap data={flowd.heatmap} />
      <Insights recap={flowd.ai_recap} aiPerBlock={flowd.ai_per_block} />
    </div>

    <Breakdown
      languages={flowd.languages
        ? Object.entries(flowd.languages).map(([name, min]) => ({ name, min }))
        : []}
      projects={flowd.by_project
        ? Object.entries(flowd.by_project).map(([name, min]) => ({ name, min }))
        : []}
      commands={flowd.by_tool
        ? Object.entries(flowd.by_tool).map(([name, min]) => ({ name, min }))
        : []}
    />

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-4">
      <div class="lg:col-span-2">
        <Timeline entries={flowd.timeline} />
      </div>
    </div>
  </main>

  <footer class="text-foreground/50 text-[10px] text-center mt-12">
    flowd — local activity tracker · self-hosted
  </footer>
</div>
