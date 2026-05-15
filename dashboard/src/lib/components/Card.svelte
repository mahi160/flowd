<script lang="ts">
  import type { Snippet } from "svelte";
  import type { HTMLAttributes } from "svelte/elements";

  interface CardProps extends Omit<HTMLAttributes<HTMLDivElement>, "children"> {
    children?: Snippet;
    heading?: string;
    description?: string;
    eyebrow?: boolean;
  }

  let {
    children,
    heading,
    description,
    eyebrow = false,
    class: className = "",
    ...props
  }: CardProps = $props();
</script>

<div
  class="flex flex-col gap-4 rounded-xl border border-border bg-surface p-4 {className}"
  {...props}
>
  {#if heading}
    <header class="flex flex-col">
      {#if eyebrow}
        <span
          class="font-mono text-[10px] uppercase tracking-[0.2em] leading-none text-foreground/50"
        >
          {heading}
        </span>
        {#if description}
          <span class="mt-1 font-mono text-[10px] text-foreground/40">
            {description}
          </span>
        {/if}
      {:else}
        <h2
          class="font-display text-[17px] font-medium leading-tight tracking-tight text-foreground"
        >
          {heading}
        </h2>
        {#if description}
          <p class="mt-0.5 font-mono text-[11px] text-foreground/50">
            {description}
          </p>
        {/if}
      {/if}
    </header>
  {/if}

  {#if children}
    <div class="flex flex-1 flex-col gap-3">
      {@render children()}
    </div>
  {/if}
</div>
