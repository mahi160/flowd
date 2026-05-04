<script lang="ts">
  import type { HTMLAttributes } from "svelte/elements";

  type CardProps = HTMLAttributes<HTMLDivElement> & {
    heading?: string;
    description?: string;
    eyebrow?: boolean;
  };

  let {
    children,
    heading,
    description,
    eyebrow = false,
    class: klass = "",
    ...props
  }: CardProps = $props();
</script>

<div
  class="rounded-xl border border-border bg-surface p-4 flex flex-col gap-3 {klass}"
  {...props}
>
  {#if heading}
    {#if eyebrow}
      <div class="flex flex-col">
        <span
          class="font-mono text-[10px] uppercase tracking-[0.18em] text-foreground/50"
        >
          {heading}
        </span>
        {#if description}
          <span class="font-mono text-[10px] text-foreground/40 mt-0.5">
            {description}
          </span>
        {/if}
      </div>
    {:else}
      <div class="flex items-baseline gap-3">
        <div>
          <h2
            class="font-display text-[17px] font-medium tracking-tight text-foreground leading-tight"
          >
            {heading}
          </h2>
          {#if description}
            <p class="font-mono text-[11px] text-foreground/50 mt-0.5">
              {description}
            </p>
          {/if}
        </div>
      </div>
    {/if}
  {/if}

  <div class="flex flex-col gap-3 flex-1">
    {@render children?.()}
  </div>
</div>
