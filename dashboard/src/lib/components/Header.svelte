<script lang="ts">
  import { flowdData } from "../store.svelte";
  import Logo from "./Logo.svelte";
  import ThemeToggle from "./ThemeToggle.svelte";
  import { Download } from "lucide-svelte";

  const tabs = ["Today", "Week", "Month"];
</script>

<header class="flex items-center justify-between">
  <div class="flex items-center gap-3">
    <Logo />

    <div class="flex flex-col gap-1">
      <h1
        class="font-display text-3xl font-bold leading-none tracking-tight text-foreground"
      >
        flowd
      </h1>

      <div
        class="flex items-center gap-2 font-mono text-[11px] text-foreground/50"
      >
        <span>{flowdData.date}</span>
        <span class="text-foreground/30" aria-hidden="true">·</span>
        <span>{flowdData.date}</span>
        <span class="text-foreground/30" aria-hidden="true">·</span>
        <span class="inline-flex items-center gap-1">
          <span class="h-1.5 w-1.5 rounded-full bg-primary animate-pulse"
          ></span>
          live
        </span>
      </div>
    </div>
  </div>

  <div class="flex items-center gap-3 text-xs font-medium">
    <nav
      class="flex rounded-xl border border-border bg-surface p-1"
      aria-label="Time period"
    >
      {#each tabs as t}
        {@const isActive = t.toLowerCase() === flowdData.period}
        <button
          aria-pressed={isActive}
          class="rounded-lg px-3 py-1 transition-colors {isActive
            ? 'bg-primary/15 text-primary'
            : 'text-foreground/60 hover:text-foreground'}"
        >
          {t}
        </button>
      {/each}
    </nav>

    <ThemeToggle />

    <button
      class="grid h-8 w-8 place-items-center rounded-lg border border-border bg-surface text-foreground/60 hover:bg-muted transition-colors"
      aria-label="Download report"
    >
      <Download size={16} />
    </button>
  </div>
</header>
