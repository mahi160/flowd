<script lang="ts">
  import Card from "../components/Card.svelte";

  const commits = 7;
  const added = 1220;
  const removed = 6545;
  const files = 28;
  const repos = 3;
  const trend = "-4%";
</script>

<Card heading="Code" eyebrow>
  <div class="font-display text-4xl text-foreground tabular-nums leading-none">
    {commits} commits
  </div>

  <div class="font-mono text-xs tabular-nums">
    <span class="text-primary">+{added.toLocaleString()}</span>
    <span class="text-foreground/30"> / </span>
    <span class="text-danger">-{removed.toLocaleString()}</span>
  </div>

  <div class="mt-auto pt-2 font-mono text-[10px] text-foreground/50">
    {files} files <span class="text-foreground/30 mx-1">·</span>
    {repos} repos <span class="text-foreground/30 mx-1">·</span>
    <span class="text-danger">↓ {trend}</span>
  </div>
</Card>
