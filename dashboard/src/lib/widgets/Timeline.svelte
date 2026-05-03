<script>
  import Card from "../components/Card.svelte";
  import { GitBranch, TrendingUp } from "lucide-svelte";

  let { entries = [] } = $props();

  const maxFocus = $derived(Math.max(1, ...entries.map((e) => e.focus ?? 0)));
</script>

<Card heading="Timeline">
  <div class="flex items-center gap-2 mb-4 text-gray-500">
    <TrendingUp class="w-4 h-4" />
    <span class="text-xs text-gray-400">focus blocks · newest first</span>
  </div>
  <div class="space-y-3">
    {#each entries.slice(0, 8) as entry}
      <div class="flex gap-4 border-l-2 border-border pl-4 py-1">
        <div class="text-xs font-mono text-gray-400 min-w-16">
          {entry.start} → {entry.end}
        </div>
        <div class="flex-1">
          <div class="font-medium text-sm">{entry.repo || "—"}</div>
          {#if entry.branch}
            <div class="text-xs text-gray-500 flex items-center gap-1">
              <GitBranch class="w-3 h-3" />
              {entry.branch}
            </div>
          {/if}
          <div class="text-xs text-gray-500 mt-1">
            {entry.focus}m focus {entry.switches > 0
              ? `· ${entry.switches} switches`
              : ""}
          </div>
          <div class="mt-2 h-1 bg-background rounded-full overflow-hidden">
            <div
              class="h-full bg-primary"
              style="width: {(entry.focus / maxFocus) * 100}%"
            ></div>
          </div>
        </div>
      </div>
    {/each}
  </div>
</Card>
