<script lang="ts">
  import Card from "../components/Card.svelte";
  import { Sparkles } from "lucide-svelte";

  const headline =
    "Your longest unbroken focus streak this week — and it landed on flowd.";
  const paragraphs = [
    "You spent 4h 02m in deep work today, with 86 minutes of unbroken editing on the dashboard split between 4–6 pm. That's your longest single block in 8 days.",
    "AI spend stayed disciplined at $3.42 — 12% below yesterday — driven mostly by Claude Code on flowd. Cache hit rate of 64% kept Sonnet costs in check.",
    "The morning saw 6 context switches, mostly between setthemacup and pi-mono. Batching shell setup work into a single afternoon block could reclaim ~15m tomorrow.",
  ];
  const highlights: { kind: "good" | "pattern" | "friction"; text: string }[] =
    [
      {
        kind: "good",
        text: "Day 14 of consistent commits — extend by tomorrow",
      },
      { kind: "pattern", text: "Peak flow at 4–6 pm, same as Tue / Fri" },
      { kind: "friction", text: "6 morning context switches, mostly tooling" },
    ];

  const toneClass: Record<string, string> = {
    good: "bg-primary/10 text-primary border-primary/30",
    pattern: "bg-accent/10 text-accent border-accent/30",
    friction: "bg-danger/10 text-danger border-danger/30",
  };
</script>

<Card heading="AI summary" description="18:17 · haiku-4.5">
  <div class="flex justify-end -mt-1">
    <button
      class="grid place-items-center w-7 h-7 rounded-full bg-accent/15 text-accent border border-accent/30"
      aria-label="Regenerate summary"
    >
      <Sparkles class="w-3.5 h-3.5" />
    </button>
  </div>

  <p
    class="font-display text-[19px] leading-snug text-foreground m-0 tracking-tight"
  >
    {headline}
  </p>

  <div
    class="flex flex-col gap-2 text-[13px] text-foreground/80 leading-relaxed"
  >
    {#each paragraphs as p, i (i)}
      <p class="m-0">{p}</p>
    {/each}
  </div>

  <div class="mt-2 flex flex-col gap-1.5 items-start">
    {#each highlights as h (h.text)}
      <span
        class="inline-flex items-center rounded-md border font-mono text-[11px] px-2 py-1 {toneClass[
          h.kind
        ]}"
      >
        {h.text}
      </span>
    {/each}
  </div>
</Card>
