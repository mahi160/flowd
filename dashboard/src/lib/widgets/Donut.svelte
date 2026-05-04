<script lang="ts">
  import Card from "../components/Card.svelte";

  const total = "4h 02m";
  const slices = [
    { id: "flowd", label: "flowd", value: "1h 42m", pct: 42, color: "var(--primary)" },
    { id: "setthemacup", label: "setthemacup", value: "1h 08m", pct: 28, color: "var(--accent)" },
    { id: "pi-mono", label: "pi-mono", value: "38m", pct: 16, color: "var(--warning)" },
    { id: "wick-ui", label: "wick-ui", value: "22m", pct: 9, color: "var(--danger)" },
    { id: "dotfiles", label: "dotfiles", value: "12m", pct: 5, color: "color-mix(in oklch, var(--primary) 60%, var(--accent))" },
  ];

  const W = 220,
    H = 220,
    cx = W / 2,
    cy = H / 2,
    rO = 90,
    rI = 58;

  function arc(s: number, e: number) {
    const large = e - s > Math.PI ? 1 : 0;
    const sx = cx + rO * Math.cos(s);
    const sy = cy + rO * Math.sin(s);
    const ex = cx + rO * Math.cos(e);
    const ey = cy + rO * Math.sin(e);
    const sxi = cx + rI * Math.cos(e);
    const syi = cy + rI * Math.sin(e);
    const exi = cx + rI * Math.cos(s);
    const eyi = cy + rI * Math.sin(s);
    return `M ${sx} ${sy} A ${rO} ${rO} 0 ${large} 1 ${ex} ${ey} L ${sxi} ${syi} A ${rI} ${rI} 0 ${large} 0 ${exi} ${eyi} Z`;
  }

  let acc = -Math.PI / 2;
  const segs = slices.map((s) => {
    const start = acc;
    const end = acc + (s.pct / 100) * Math.PI * 2;
    acc = end;
    return { ...s, start, end };
  });
</script>

<Card heading="By project" description="today">
  <div class="relative aspect-square max-h-[200px] mx-auto w-full">
    <svg viewBox="0 0 {W} {H}" class="block w-full h-full">
      {#each segs as s (s.id)}
        <path d={arc(s.start, s.end)} fill={s.color} />
      {/each}
    </svg>
    <div
      class="pointer-events-none absolute inset-0 flex flex-col items-center justify-center"
    >
      <div class="font-display text-2xl tabular-nums text-foreground leading-none">
        {total}
      </div>
      <div
        class="font-mono text-[10px] uppercase tracking-wider text-foreground/50 mt-1"
      >
        total
      </div>
    </div>
  </div>

  <ul class="m-0 p-0 list-none flex flex-col gap-1">
    {#each segs as s (s.id)}
      <li
        class="grid grid-cols-[10px_1fr_auto_36px] gap-2 items-center px-1 py-0.5 text-xs"
      >
        <span class="block w-2.5 h-2.5 rounded-sm" style:background={s.color}
        ></span>
        <span class="font-mono text-foreground/80 truncate">{s.label}</span>
        <span class="font-mono text-foreground/60 tabular-nums">{s.value}</span>
        <span
          class="font-mono text-foreground/40 tabular-nums text-right"
        >
          {s.pct}%
        </span>
      </li>
    {/each}
  </ul>
</Card>
