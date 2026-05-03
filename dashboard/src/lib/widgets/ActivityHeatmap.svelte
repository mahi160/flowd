<script>
  import { Activity } from "lucide-svelte";
  import Card from "../components/Card.svelte";

  let { data = [] } = $props();

  const cells = $derived.by(() => {
    const hourly = Array(24).fill(0);
    for (const item of data) {
      hourly[item.hour ?? 0] += item.minute ?? 0;
    }
    return hourly;
  });

  const max = $derived(Math.max(1, ...cells));
</script>

<Card heading="Activity Heatmap">
  <div class="flex items-center gap-2 mb-4 text-gray-500">
    <Activity class="w-4 h-4" />
    <span class="text-xs text-gray-400">by hour of day</span>
  </div>
  <div class="heat-axis flex gap-1 justify-between text-xs text-gray-400 mb-2">
    {#each ["8a", "12p", "4p", "8p"] as label}
      <span>{label}</span>
    {/each}
  </div>
  <div class="flex items-end gap-1 h-16">
    {#each cells as value}
      <div
        class="flex-1 bg-primary/60 rounded-t transition"
        style="height: {(value / max) * 100}%"
        title="{Math.round((value / max) * 100)}%"
      ></div>
    {/each}
  </div>
</Card>
