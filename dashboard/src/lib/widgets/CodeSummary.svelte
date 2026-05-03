<script>
  import Card from "../components/Card.svelte";
  import { flowd } from "../store.svelte";
</script>

<Card heading="Code" description="">
  <p class="font-display text-2xl font-light tabular-nums">
    {flowd?.files_changed} files
  </p>
  <div
    class="flex gap-2 items-center text-xs font-mono text-foreground/30 tabular-nums"
  >
    <span class="text-accent">+{flowd?.lines_added}</span>/
    <span class="text-danger">-{flowd?.lines_removed}</span>
  </div>
  <div>
    <span>js</span>
  </div>
</Card>
