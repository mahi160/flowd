<script lang="ts">
  import type { Snippet } from "svelte";
  let {
    text,
    children,
    position = "top",
  }: { text: string; children?: Snippet; position?: "top" | "bottom" } = $props();
  let visible = $state(false);
</script>

<span class="relative inline-flex items-center gap-1">
  {@render children?.()}
  <!-- Info icon trigger -->
  <button
    class="text-stone-300 dark:text-stone-600 hover:text-indigo-400 dark:hover:text-indigo-400
           transition-colors cursor-help flex-shrink-0 focus:outline-none"
    onmouseenter={() => (visible = true)}
    onmouseleave={() => (visible = false)}
    onfocus={() => (visible = true)}
    onblur={() => (visible = false)}
    tabindex="-1"
    aria-label={text}
  >
    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round">
      <circle cx="12" cy="12" r="10" />
      <line x1="12" y1="16" x2="12" y2="12" />
      <circle cx="12" cy="8" r="0.5" fill="currentColor" />
    </svg>
  </button>

  {#if visible}
    <div
      role="tooltip"
      class="absolute left-1/2 -translate-x-1/2 z-50 w-56 rounded-lg
             bg-stone-800 dark:bg-stone-700 text-stone-100
             text-[11px] leading-relaxed px-3 py-2 shadow-xl pointer-events-none text-left font-normal
             {position === 'top' ? 'bottom-full mb-2' : 'top-full mt-2'}"
    >
      {text}
      <!-- Arrow -->
      <div class="absolute left-1/2 -translate-x-1/2 w-0 h-0 border-[5px] border-transparent
                  {position === 'top'
                    ? 'top-full border-t-stone-800 dark:border-t-stone-700'
                    : 'bottom-full border-b-stone-800 dark:border-b-stone-700'}">
      </div>
    </div>
  {/if}
</span>
