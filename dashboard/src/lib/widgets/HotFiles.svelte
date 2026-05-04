<script lang="ts">
  import Card from "../components/Card.svelte";

  const files = [
    { path: "src/dashboard/routes/+page.svelte", repo: "flowd", changes: 42, diff: "+312 -180" },
    { path: "src/lib/aggregate.ts", repo: "flowd", changes: 28, diff: "+204 -41" },
    { path: "themes/macos/setup.sh", repo: "setthemacup", changes: 24, diff: "+88 -2,210" },
    { path: "src/lib/components/Donut.svelte", repo: "flowd", changes: 18, diff: "+142 -22" },
    { path: "init.lua", repo: "pi-mono", changes: 9, diff: "+44 -12" },
  ];
</script>

<Card heading="Hot paths" description="most-touched files today">
  <ul class="m-0 p-0 list-none">
    {#each files as f, i (f.path)}
      <li
        class="grid grid-cols-[22px_1fr_auto_auto] gap-2 items-center py-2 border-t border-dashed border-border first:border-t-0"
      >
        <span class="font-mono text-[10px] text-foreground/40 tabular-nums">
          {String(i + 1).padStart(2, "0")}
        </span>
        <div class="flex flex-col gap-0.5 min-w-0">
          <code class="font-mono text-xs text-foreground truncate">
            {f.path}
          </code>
          <span class="font-mono text-[10px] text-foreground/50">
            {f.repo}
          </span>
        </div>
        <span class="font-mono text-xs text-foreground/80 tabular-nums">
          {f.changes}<span class="text-foreground/40"> edits</span>
        </span>
        <span class="font-mono text-[10px] text-foreground/50 tabular-nums">
          {f.diff}
        </span>
      </li>
    {/each}
  </ul>
</Card>
