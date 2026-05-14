<script lang="ts">
  import Card from "../components/Card.svelte";

  const weekLabel = "Apr 27 – May 03";
  const days = [
    {
      date: "2026-04-27",
      label: "Mon 27",
      focus: 240,
      focusLabel: "4h 00m",
      isToday: false,
      isBest: false,
    },
    {
      date: "2026-04-28",
      label: "Tue 28",
      focus: 312,
      focusLabel: "5h 12m",
      isToday: false,
      isBest: false,
    },
    {
      date: "2026-04-29",
      label: "Wed 29",
      focus: 198,
      focusLabel: "3h 18m",
      isToday: false,
      isBest: false,
    },
    {
      date: "2026-04-30",
      label: "Thu 30",
      focus: 384,
      focusLabel: "6h 24m",
      isToday: false,
      isBest: true,
    },
    {
      date: "2026-05-01",
      label: "Fri 01",
      focus: 264,
      focusLabel: "4h 24m",
      isToday: false,
      isBest: false,
    },
    {
      date: "2026-05-02",
      label: "Sat 02",
      focus: 60,
      focusLabel: "1h 00m",
      isToday: false,
      isBest: false,
    },
    {
      date: "2026-05-03",
      label: "Sun 03",
      focus: 342,
      focusLabel: "5h 42m",
      isToday: true,
      isBest: false,
    },
  ];

  const max = Math.max(...days.map((d) => d.focus), 1);

  function bg(d: (typeof days)[number]) {
    if (d.isToday) return "var(--accent)";
    if (d.isBest) return "var(--primary)";
    return "color-mix(in oklch, var(--primary) 60%, transparent)";
  }
</script>

<Card heading="Week at a Glance" description={weekLabel}>
  <div class="grid grid-cols-7 gap-2 h-[200px] mt-2">
    {#each days as d (d.date)}
      <div class="flex flex-col">
        <div class="flex-1 bg-background rounded-xl relative overflow-hidden">
          <span
            class="absolute inset-x-0 bottom-0 rounded-t-xl flex justify-center pt-2"
            style:height="{(d.focus / max) * 100}%"
            style:background={bg(d)}
          >
            <span
              class="font-mono text-[10px] font-semibold tabular-nums text-background"
            >
              {d.focusLabel}
            </span>
          </span>
        </div>
        <div class="text-center pt-2 mt-2 border-t border-border">
          <div
            class="font-mono text-[10px] uppercase tracking-wider text-foreground/50"
          >
            {d.label.split(" ")[0]}
          </div>
          <div
            class="font-display text-base mt-0.5 tabular-nums {d.isToday
              ? 'text-accent'
              : 'text-foreground'}"
          >
            {d.label.split(" ")[1]}
          </div>
        </div>
      </div>
    {/each}
  </div>
</Card>
