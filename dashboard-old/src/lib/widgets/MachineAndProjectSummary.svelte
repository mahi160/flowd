<script lang="ts">
  import Card from "../components/Card.svelte";

  const machine = "workstation";
  const os = "macOS · darwin";
  const uptime = "up 4d 2h";
  const topRepo = "flowd";
  const topRepoTime = "1h 42m";
  const topRepoCommits = 3;
</script>

<Card heading="Machine" eyebrow>
  <div class="font-display text-2xl text-foreground leading-none">
    {machine}
  </div>

  <div class="font-mono text-xs text-foreground/60">
    {os} · {uptime}
  </div>

  <div class="mt-auto pt-2 font-mono text-[10px] text-foreground/50">
    top repo
    <span
      class="inline-flex items-center rounded bg-primary/15 text-primary px-1.5 py-0.5 mx-1"
    >
      {topRepo}
    </span>
    <span class="text-foreground/30 mx-1">·</span>
    {topRepoTime}
    <span class="text-foreground/30 mx-1">·</span>
    {topRepoCommits} commits
  </div>
</Card>
