<script lang="ts">
  import Logo from "../components/Logo.svelte";
  import ThemeToggle from "../components/ThemeToggle.svelte";
  import { flowd } from "../store.svelte";

  const periods = ["day", "week"];
</script>

<header class="flex items-center justify-between">
  <section class="flex items-center gap-4">
    <Logo />

    <div class="flex flex-col gap-2">
      <span
        class="font-display text-3xl leading-none tracking-widest text-primary"
      >
        flowd
      </span>

      <div class="flex items-center gap-3 font-mono text-xs text-foreground/50">
        <span class="uppercase tracking-[0.25em] text-[10px]"> Generated </span>
        <span>|</span>
        <span class="tracking-tight">{flowd?.generated}</span>
      </div>
    </div>
  </section>

  <section class="flex items-center gap-3 text-xs font-medium">
    <div class="flex rounded-xl bg-surface p-1 shadow-sm">
      {#each periods as p}
        <button
          data-active={p === flowd?.period ? "" : undefined}
          class="capitalize rounded-lg data-active:bg-primary/8 px-4 py-1.5"
        >
          {p}
        </button>
      {/each}
    </div>
    <ThemeToggle />
  </section>
</header>
