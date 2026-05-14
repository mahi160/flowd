<script lang="ts">
  import Card from "../components/Card.svelte";

  const spend = "$3.42";
  const tokensIn = "412K";
  const tokensOut = "88K";
  const trend = "-12%";
  const cache = "64%";
</script>

<Card heading="AI today" eyebrow>
  <div class="font-display text-4xl text-accent tabular-nums leading-none">
    {spend}
  </div>

  <div class="font-mono text-xs text-foreground/60 tabular-nums">
    {tokensIn} in · {tokensOut} out
  </div>

  <div class="mt-auto pt-2 font-mono text-[10px] text-foreground/50">
    <span class="text-primary">↓ {trend}</span>
    <span class="text-foreground/30 mx-1">·</span>
    {cache} cache
  </div>
</Card>
