<script lang="ts">
  import Card from "../components/Card.svelte";

  const total = "38m";
  const pctOfDay = 8;
  const apps = [
    { id: "twitter", label: "twitter", value: 18, valueLabel: "18m", color: "var(--danger)" },
    { id: "youtube", label: "youtube", value: 12, valueLabel: "12m", color: "var(--warning)" },
    { id: "hn", label: "hn", value: 8, valueLabel: "8m", color: "var(--foreground)" },
  ];
</script>

<Card heading="Distractions" description="{total} · {pctOfDay}% of day">
  <div class="flex h-1.5 rounded-full overflow-hidden border border-border">
    {#each apps as a (a.id)}
      <span class="block h-full" style:flex={a.value} style:background={a.color}
      ></span>
    {/each}
  </div>

  <ul class="m-0 p-0 list-none flex flex-col gap-1.5 mt-1">
    {#each apps as a (a.id)}
      <li class="grid grid-cols-[10px_1fr_auto] gap-2 items-center text-xs">
        <span class="block w-2.5 h-2.5 rounded-sm" style:background={a.color}
        ></span>
        <span class="font-mono text-foreground/80">{a.label}</span>
        <span class="font-mono tabular-nums text-foreground/60">
          {a.valueLabel}
        </span>
      </li>
    {/each}
  </ul>
</Card>
