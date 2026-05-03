<script>
  import Card from "../components/Card.svelte";
  import { Sparkles } from "lucide-svelte";

  let { recap = "", aiPerBlock = 0 } = $props();
</script>

<Card heading="AI Insights">
  <div class="flex items-center gap-2 mb-4 text-accent">
    <Sparkles class="w-5 h-5" />
    {#if recap}
      <span class="text-xs bg-accent/10 text-accent px-2 py-1 rounded"
        >recap</span
      >
    {:else if aiPerBlock > 0}
      <span class="text-xs bg-accent/10 text-accent px-2 py-1 rounded"
        >{aiPerBlock} blocks</span
      >
    {/if}
  </div>

  {#if recap}
    <p class="text-sm text-gray-600">{recap}</p>
  {:else if aiPerBlock > 0}
    <p class="text-sm text-gray-600">
      Per-block AI summaries are inline in the timeline. Run <code
        class="text-xs bg-gray-100 px-1.5 py-0.5 rounded"
        >fw dashboard --ai-recap</code
      > for an aggregate.
    </p>
  {:else}
    <p class="text-sm text-gray-600">
      Set <code class="text-xs bg-gray-100 px-1.5 py-0.5 rounded"
        >ai_enabled: true</code
      > in your config to see AI insights here.
    </p>
  {/if}
</Card>
