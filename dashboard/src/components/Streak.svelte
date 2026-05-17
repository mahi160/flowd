<script lang="ts">
  import type { Data } from "../lib/transform";

  let { data }: { data: Data } = $props();
  let s = $derived(data.streakDays);
  let sub = $derived(
    s === 0 ? "start coding today" : s >= 14 ? "on fire 🔥" : s >= 7 ? "keep it up 🌿" : "building momentum"
  );
</script>

<section class="rounded-xl border border-stone-200 dark:border-stone-800 bg-stone-50 dark:bg-stone-900 shadow-sm dark:shadow-stone-950/30 p-5">
  <div class="flex items-baseline gap-3 mb-4">
    <h3 class="font-display text-lg text-stone-900 dark:text-stone-100">Streak</h3>
    <span class="text-[11px] font-mono text-stone-400">30 days</span>
  </div>
  <div class="flex items-end gap-3 mb-5">
    <span class="font-display text-6xl leading-none text-indigo-600 dark:text-indigo-400 tabular-nums tracking-tighter">{s}</span>
    <div class="flex flex-col gap-0.5 pb-2">
      <span class="text-sm text-stone-700 dark:text-stone-200 font-medium">day streak</span>
      <span class="text-[11px] text-stone-400 font-mono">{sub}</span>
    </div>
  </div>
  <div class="grid grid-cols-10 gap-1 mb-2">
    {#each data.streakCells as cell}
      {@const intensity = cell.v / 4}
      <span
        class="h-3.5 rounded transition-transform hover:scale-125 cursor-default"
        title={cell.d === 29 ? "today" : `${29 - cell.d}d ago`}
        style:background={intensity > 0.01 ? `rgb(99 102 241 / ${Math.round(intensity * 90)}%)` : ''}
        class:bg-stone-100={intensity <= 0.01}
        class:dark:bg-stone-800={intensity <= 0.01}
        style:outline={cell.d === 29 ? "2px solid rgb(99 102 241 / 50%)" : "none"}
        style:outline-offset="1px"
      ></span>
    {/each}
  </div>
  <div class="flex justify-between font-mono text-[10px] text-stone-400">
    <span>29d ago</span>
    <span>today</span>
  </div>
</section>
