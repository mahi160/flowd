<span class="inline-flex items-center gap-1 text-[10px] text-stone-400 font-mono">
  less
  {#each [0.05, 0.3, 0.55, 0.8, 1] as i}
    <span class="w-3 h-3 rounded" style:background="rgb(99 102 241 / {Math.round(i * 100)}%)"></span>
  {/each}
  more
</span>
