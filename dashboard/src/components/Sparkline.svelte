<script lang="ts">
  let { data }: { data: number[] } = $props();

  let max = $derived(Math.max(1, ...data));
  let pts = $derived(
    data.map((v, i) => `${(i / 23) * 180},${32 - (v / max) * 28}`).join(" ")
  );
  let fillPts = $derived(
    `0,34 ${data.map((v, i) => `${(i / 23) * 180},${32 - (v / max) * 28}`).join(" ")} 180,34`
  );
</script>

<svg class="block mt-auto w-full" viewBox="0 0 180 36" preserveAspectRatio="none">
  <defs>
    <linearGradient id="sf" x1="0" y1="0" x2="0" y2="1">
      <stop offset="0%" stop-color="rgb(99 102 241)" stop-opacity="0.2" />
      <stop offset="100%" stop-color="rgb(99 102 241)" stop-opacity="0" />
    </linearGradient>
  </defs>
  <polygon points={fillPts} fill="url(#sf)" />
  <polyline points={pts} fill="none" stroke="rgb(99 102 241)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
</svg>
