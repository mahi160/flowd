<script lang="ts">
  import Card from "../components/Card.svelte";

  const hours = 4;
  const mins = 2;
  const blocks = 15;
  const longest = "38m";
  const switches = 6;
  const trend = "+18%";
  const friction = "low";
</script>

<Card heading="Focus today" eyebrow>
  <div class="flex items-baseline gap-1 text-accent font-display tabular-nums">
    <span class="text-5xl leading-none">{hours}</span>
    <span class="text-2xl text-accent/70 mr-1">h</span>
    <span class="text-5xl leading-none">{String(mins).padStart(2, "0")}</span>
    <span class="text-2xl text-accent/70">m</span>
  </div>

  <div
    class="flex items-center gap-2 text-xs text-foreground/60 flex-wrap font-mono"
  >
    <span
      ><span class="tabular-nums text-foreground/80">{blocks}</span> blocks</span
    >
    <span class="text-foreground/30">·</span>
    <span
      >longest <span class="tabular-nums text-foreground/80">{longest}</span
      ></span
    >
    <span class="text-foreground/30">·</span>
    <span
      ><span class="tabular-nums text-foreground/80">{switches}</span> switches</span
    >
  </div>

  <div class="mt-auto pt-2 font-mono text-[10px] text-foreground/50">
    <span class="text-primary">↑ {trend}</span> vs yesterday
    <span class="text-foreground/30 mx-1">·</span>
    friction {friction}
  </div>
</Card>
