<script lang="ts">
  import type { HTMLAttributes } from "svelte/elements";

  type CardProps = HTMLAttributes<HTMLDivElement> & {
    heading: string;
    description?: string;
  };

  let { children, heading, description, ...props }: CardProps = $props();
</script>

<div class="rounded-xl shadow bg-surface p-4 flex flex-col gap-3" {...props}>
  <div class="flex gap-2 items-baseline justify-start">
    <h2
      class="font-display text-xs tracking-wider font-light leading-none text-foreground/60 uppercase"
    >
      {heading}
    </h2>

    {#if description}
      <h4 class="font-mono text-xs text-foreground/40">
        {description}
      </h4>
    {/if}
  </div>

  <div>
    {@render children?.()}
  </div>
</div>
