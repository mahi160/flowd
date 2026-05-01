<script lang="ts">
  export let html: string;
  export let size: number = 16;
</script>

<span style="width:{size}px; height:{size}px; display:inline-flex;">
  {@html html}
</span>
