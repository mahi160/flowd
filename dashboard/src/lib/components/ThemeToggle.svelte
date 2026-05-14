<script>
  import { Moon, Sun } from "lucide-svelte";
  import { onMount } from "svelte";

  let dark = $state(false);

  function toggleTheme() {
    dark = !dark;
  }

  onMount(() => {
    dark = window.matchMedia("(prefers-color-scheme: dark)").matches;
  });

  $effect(() => {
    document.documentElement.dataset.theme = dark ? "dark" : "light";
  });
</script>

<button
  onclick={toggleTheme}
  aria-label="Toggle theme"
  class="size-8 flex items-center justify-center bg-surface rounded-lg shadow-sm cursor-pointer"
>
  {#if dark}
    <Moon size={16} />
  {:else}
    <Sun size={16} />
  {/if}
</button>
