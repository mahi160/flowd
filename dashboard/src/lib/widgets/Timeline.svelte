<script lang="ts">
  import Card from "../components/Card.svelte";

  type Block = {
    start: string;
    end: string;
    repo?: string;
    branch?: string;
    tool?: string;
    cost?: string;
    focus?: number;
    switches?: number;
    note?: string;
    idle?: boolean;
  };

  const entries: Block[] = [
    { start: "08:30", end: "09:00", repo: "setthemacup", branch: "main", tool: "nvim", focus: 22, switches: 0, note: "theme tweaks" },
    { start: "09:00", end: "09:30", repo: "pi-mono", branch: "main", tool: "nvim", focus: 18, switches: 2 },
    { start: "09:30", end: "10:30", repo: "flowd", branch: "main", tool: "nvim", focus: 50, switches: 1, note: "longest block · aggregate.ts" },
    { start: "10:30", end: "11:00", idle: true, note: "idle" },
    { start: "11:00", end: "11:30", repo: "wick-ui", branch: "sifat-main-i18n", tool: "nvim", focus: 22, switches: 1 },
    { start: "11:30", end: "12:30", repo: "flowd", branch: "main", tool: "Claude Code", cost: "$1.42", focus: 38, switches: 0, note: "session refactor" },
    { start: "12:30", end: "13:30", idle: true, note: "distraction · lunch away" },
    { start: "13:30", end: "15:00", repo: "flowd", branch: "feat/dashboard", tool: "Cursor", focus: 78, switches: 2, note: "dashboard split" },
    { start: "15:00", end: "16:00", idle: true, note: "meeting · standup + review" },
    { start: "16:00", end: "18:00", repo: "flowd", branch: "feat/dashboard", tool: "nvim", focus: 86, switches: 1, note: "peak · longest unbroken" },
  ];

  const maxFocus = Math.max(1, ...entries.map((e) => e.focus ?? 0));
</script>

<Card heading="Timeline" description="today · contiguous blocks">
  <ul class="m-0 p-0 list-none flex flex-col">
    {#each entries as e, i (i)}
      <li
        class="grid grid-cols-[80px_1fr] gap-3 items-start py-2 border-t border-dashed border-border first:border-t-0"
      >
        <span class="font-mono text-[11px] text-foreground/50 tabular-nums">
          {e.start} → {e.end}
        </span>
        {#if e.idle}
          <div class="text-xs text-foreground/40 italic">— {e.note}</div>
        {:else}
          <div class="flex flex-col gap-1 min-w-0">
            <div class="flex items-center gap-1.5 flex-wrap">
              <span class="font-mono text-sm text-foreground">{e.repo}</span>
              {#if e.branch}
                <span
                  class="inline-flex items-center rounded bg-primary/15 text-primary px-1.5 py-0.5 font-mono text-[10px]"
                >
                  ⎇ {e.branch}
                </span>
              {/if}
              {#if e.tool}
                <span
                  class="inline-flex items-center rounded bg-foreground/10 text-foreground/70 px-1.5 py-0.5 font-mono text-[10px]"
                >
                  {e.tool}
                </span>
              {/if}
              {#if e.cost}
                <span
                  class="inline-flex items-center rounded bg-accent/15 text-accent px-1.5 py-0.5 font-mono text-[10px]"
                >
                  {e.cost}
                </span>
              {/if}
            </div>
            <div class="font-mono text-[11px] text-foreground/60">
              {e.focus}m focus
              <span class="text-foreground/30 mx-0.5">·</span>
              {e.switches} switches
              {#if e.note}
                <span class="text-foreground/30 mx-0.5">·</span>
                <span class="italic">{e.note}</span>
              {/if}
            </div>
            <div class="h-1 bg-background rounded-full overflow-hidden mt-0.5">
              <div
                class="h-full bg-accent"
                style:width="{((e.focus ?? 0) / maxFocus) * 100}%"
              ></div>
            </div>
          </div>
        {/if}
      </li>
    {/each}
  </ul>
</Card>
