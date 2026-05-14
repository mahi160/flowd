<script lang="ts">
  import Logo from "./Logo.svelte";
  import ThemeToggle from "./ThemeToggle.svelte";
  import { Download } from "lucide-svelte";

  const dateLabel = "Sun 3 May 2026";
  const time = "18:17";
  const live = "just now";
  const tabs = ["Today", "Week", "Month", "Project"];
  const activeTab = "Today";
</script>

<header class="flex items-center justify-between">
  <section class="flex items-center gap-3">
    <Logo />

    <div class="flex flex-col gap-1">
      <span
        class="font-display text-3xl leading-none tracking-tight text-foreground"
      >
        flowd
      </span>

      <div
        class="flex items-center gap-2 font-mono text-[11px] text-foreground/50"
      >
        <span>{dateLabel}</span>
        <span class="text-foreground/30">·</span>
        <span>{live}</span>
        <span class="text-foreground/30">·</span>
        <span>{time}</span>
        <span class="text-foreground/30">·</span>
        <span class="inline-flex items-center gap-1">
          <span class="block w-1.5 h-1.5 rounded-full bg-primary"></span>
          live
        </span>
      </div>
    </div>
  </section>

  <section class="flex items-center gap-3 text-xs font-medium">
    <div class="flex rounded-xl border border-border bg-surface p-1">
      {#each tabs as t}
        <button
          data-active={t === activeTab ? "" : undefined}
          class="rounded-lg px-3 py-1 data-active:bg-primary/15 data-active:text-primary text-foreground/60"
        >
          {t}
        </button>
      {/each}
    </div>
    <ThemeToggle />
    <button
      class="grid place-items-center w-8 h-8 rounded-lg border border-border bg-surface text-foreground/60"
      aria-label="Export"
    >
      <Download class="w-4 h-4" />
    </button>
  </section>
</header>
