<script lang="ts">
  import type { Data } from "../lib/transform";
  import { fmtHM } from "../lib/format";
  import Tooltip from "./Tooltip.svelte";

  let { data }: { data: Data } = $props();

  const DOW_LABELS = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"];
  // Mon-first row order: Mon=0 … Sun=6
  const ROWS = [1, 2, 3, 4, 5, 6, 0]; // Go dow values in display order

  // Organise cal_days into week columns (each col = 7 cells, Sun…Sat).
  let weeks = $derived.by(() => {
    const days = data.calDays;
    if (!days.length) return [];
    const map: Record<string, typeof days[0]> = {};
    for (const d of days) map[d.date] = d;

    // Find first Sunday on or before the first day.
    const firstDate = new Date(days[0].date + "T00:00:00");
    const start = new Date(firstDate);
    start.setDate(start.getDate() - firstDate.getDay()); // back to Sunday

    const lastDate = new Date(days[days.length - 1].date + "T00:00:00");
    const end = new Date(lastDate);
    end.setDate(end.getDate() + (6 - lastDate.getDay())); // forward to Saturday

    const cols: ({ date: string; min: number; blocks: number; inPeriod: boolean } | null)[][] = [];
    let col: ({ date: string; min: number; blocks: number; inPeriod: boolean } | null)[] = [];

    for (let d = new Date(start); d <= end; d.setDate(d.getDate() + 1)) {
      const iso = d.toISOString().slice(0, 10);
      const entry = map[iso];
      col.push(entry ? { date: iso, min: entry.min, blocks: entry.blocks, inPeriod: true }
                     : { date: iso, min: 0, blocks: 0, inPeriod: false });
      if (d.getDay() === 6) { // Saturday = end of week
        cols.push(col);
        col = [];
      }
    }
    if (col.length) cols.push(col);
    return cols;
  });

  let maxMin = $derived(Math.max(1, ...data.calDays.map((d) => d.min)));

  // For year view: extract month labels (show label at first week of each month)
  let monthLabels = $derived.by(() => {
    if (data.period !== "year") return [];
    const labels: { col: number; label: string }[] = [];
    let lastMonth = -1;
    weeks.forEach((week, i) => {
      const day = week.find((d) => d?.inPeriod);
      if (!day) return;
      const m = new Date(day.date + "T00:00:00").getMonth();
      if (m !== lastMonth) {
        labels.push({
          col: i,
          label: new Date(day.date + "T00:00:00").toLocaleString("default", { month: "short" }),
        });
        lastMonth = m;
      }
    });
    return labels;
  });

  function fmtDate(iso: string) {
    const d = new Date(iso + "T00:00:00");
    return d.toLocaleDateString("default", { weekday: "short", day: "numeric", month: "short", year: "numeric" });
  }

  function cellColor(min: number, inPeriod: boolean) {
    if (!inPeriod) return "transparent";
    if (min <= 0) return "";
    const intensity = Math.min(1, min / maxMin);
    const pct = Math.round(Math.max(0.12, intensity) * 100);
    return `rgb(99 102 241 / ${pct}%)`;
  }
</script>

<section class="rounded-xl border border-stone-200 dark:border-stone-800 bg-stone-50 dark:bg-stone-900 shadow-sm p-5">
  <div class="flex items-baseline gap-3 mb-4">
    <h3 class="font-display text-lg text-stone-900 dark:text-stone-100">Activity calendar</h3>
    <span class="text-[11px] font-mono text-stone-400">
      {data.period === "year" ? "this year · daily" : "this month · daily"}
    </span>
    <Tooltip text="Each cell is one calendar day. Colour intensity = focused minutes. Only days with recorded activity are filled in." />
  </div>

  {#if !data.calDays.length}
    <p class="text-stone-400 text-sm font-mono py-4">No data for this period.</p>
  {:else}
    <div class="overflow-x-auto pb-1">
      <div class="inline-block min-w-full">
        <!-- Month labels (year only) -->
        {#if data.period === "year" && monthLabels.length}
          <div class="flex mb-1 pl-8">
            {#each weeks as _, i}
              {@const lbl = monthLabels.find((l) => l.col === i)}
              <div class="w-[14px] shrink-0 text-[10px] text-stone-400 font-mono">
                {lbl ? lbl.label : ""}
              </div>
            {/each}
          </div>
        {/if}

        <div class="flex gap-0.5">
          <!-- Day-of-week labels (Mon-first) -->
          <div class="flex flex-col gap-0.5 mr-1.5 shrink-0 justify-around">
            {#each ROWS as dow}
              <div class="h-[13px] w-6 text-[9px] text-stone-400 font-mono flex items-center justify-end pr-1">
                {dow % 2 === 1 ? DOW_LABELS[dow] : ""}
              </div>
            {/each}
          </div>

          <!-- Week columns -->
          {#each weeks as week}
            <div class="flex flex-col gap-0.5">
              {#each ROWS as dow}
                {@const cell = week[dow]}
                <div
                  class="w-[13px] h-[13px] rounded-[2px] transition-transform hover:scale-150 hover:z-10 relative cursor-default
                         {cell?.inPeriod ? 'bg-stone-100 dark:bg-stone-800' : 'opacity-0'}"
                  style:background={cell ? cellColor(cell.min, cell.inPeriod) : "transparent"}
                  title={cell?.inPeriod
                    ? `${fmtDate(cell.date)}${cell.min > 0 ? " · " + fmtHM(cell.min) + " focused" + (cell.blocks > 0 ? " · " + cell.blocks + " block" + (cell.blocks > 1 ? "s" : "") : "") : " · no activity"}`
                    : ""}
                ></div>
              {/each}
            </div>
          {/each}
        </div>

        <!-- Legend -->
        <div class="flex items-center gap-1.5 mt-3 font-mono text-[10px] text-stone-400 pl-8">
          less
          {#each [0.12, 0.35, 0.6, 0.8, 1] as i}
            <span class="w-3 h-3 rounded-[2px] bg-stone-100 dark:bg-stone-800"
              style:background="rgb(99 102 241 / {Math.round(i * 100)}%)"></span>
          {/each}
          more
        </div>
      </div>
    </div>

    <!-- Summary row -->
    <div class="flex flex-wrap gap-x-6 gap-y-1.5 mt-4 pt-3 border-t border-stone-100 dark:border-stone-800 text-[11px]">
      <span class="font-mono text-stone-400">
        <span class="text-stone-700 dark:text-stone-200 font-medium">{data.activeDays}</span> active days
      </span>
      {#if data.bestDayDate}
        <span class="font-mono text-stone-400">
          best day: <span class="text-stone-700 dark:text-stone-200 font-medium">
            {new Date(data.bestDayDate + "T00:00:00").toLocaleDateString("default", { day: "numeric", month: "short" })}
            ({fmtHM(data.bestDayMin)})
          </span>
        </span>
      {/if}
      <span class="font-mono text-stone-400">
        total: <span class="text-stone-700 dark:text-stone-200 font-medium">{fmtHM(data.focus.totalMin)}</span>
      </span>
    </div>
  {/if}
</section>
