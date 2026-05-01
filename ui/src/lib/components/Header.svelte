<script lang="ts">
  import FlowMark from "../icons/FlowMark.svelte";

  interface IHeaderProps {
    generated: string;
    hasData: boolean;
    period: string;
  }

  let data: IHeaderProps = $props();
</script>

<header class="flex items-center justify-between pb-6 mb-1.5">
  <div class="flex items-center gap-3.5">
    <FlowMark />
    <div>
      <div class="font-display text-3xl leading-none mb-1 tracking-tight">
        flowd
      </div>
      <div class="flex items-center gap-2 text-xs text-foreground/60">
        <span
          class="font-mono text-[10px] tracking-wider uppercase text-foreground/40"
        >
          Generated
        </span>
        <span class="font-mono text-sm tabular-nums">
          {data.generated}
        </span>
        {#if data.hasData}
          <span class="text-foreground/30">·</span>
          <span
            class="inline-flex items-center gap-2 px-2 py-0.5 rounded-full text-sm font-mono bg-foreground/5 border border-foreground/15 text-foreground/70"
          >
            <span
              class="w-1.5 h-1.5 rounded-full bg-primary shadow-[0_0_0_3px_oklch(from_var(--color-primary)_l_c_h/0.22)]"
            ></span>
            {data.period}
          </span>
        {/if}
      </div>
    </div>
  </div>

  <!-- <!-- Right controls -->
  <!-- <div class="flex items-center gap-2.5"> -->
  <!--   <!-- Segmented control -->
  <!--   <div -->
  <!--     class="inline-flex bg-foreground/5 border border-foreground/15 rounded-lg p-[3px] gap-0" -->
  <!--   > -->
  <!--     {#each views as v} -->
  <!--       <button -->
  <!--         class="border-0 px-3.5 py-1.5 text-[12.5px] rounded-md cursor-pointer font-sans transition-colors duration-100 -->
  <!--           {view === v -->
  <!--           ? 'bg-background text-foreground shadow-[0_1px_0_oklch(0_0_0_/_0.05)]' -->
  <!--           : 'bg-transparent text-foreground/50 hover:text-foreground/80'}" -->
  <!--         title={v !== data.period -->
  <!--           ? `Run fw dashboard ${v} to generate ${v} data` -->
  <!--           : ""} -->
  <!--         on:click={() => setView(v)} -->
  <!--       > -->
  <!--         {v === "today" ? "Today" : "Week"} -->
  <!--       </button> -->
  <!--     {/each} -->
  <!--   </div> -->
  <!---->
  <!--   <!-- Theme toggle -->
  <!--   <button -->
  <!--     class="w-9 h-9 rounded-lg bg-foreground/5 border border-foreground/15 text-foreground/60 -->
  <!--            inline-grid place-items-center cursor-pointer -->
  <!--            hover:text-foreground hover:border-foreground/40 -->
  <!--            transition-colors duration-100" -->
  <!--     title="Theme: {theme}" -->
  <!--     on:click={cycleTheme} -->
  <!--   > -->
  <!--     <SvgIcon html={ICONS[theme]} size={18} /> -->
  <!--   </button> -->
  <!-- </div> -->
</header>
