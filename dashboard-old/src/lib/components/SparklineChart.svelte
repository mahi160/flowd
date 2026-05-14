<script lang="ts">
  type Props = {
    data: number[];
  };

  let { data }: Props = $props();

  const max = $derived.by(() => {
    if (!data.length) return 1;
    return Math.max(1, ...data);
  });

  const pts = $derived.by(() => {
    const len = data.length || 1;

    return data
      .map((v, i) => {
        const x = (i / Math.max(1, len - 1)) * 180;
        const y = 32 - (v / max) * 28;
        return `${x},${y}`;
      })
      .join(" ");
  });
</script>

<svg class="sparkline" viewBox="0 0 180 36">
  <polyline
    points={pts}
    fill="none"
    stroke="var(--color-primary)"
    stroke-width="1"
    stroke-linecap="round"
    stroke-linejoin="round"
    opacity=".8"
  />
</svg>
