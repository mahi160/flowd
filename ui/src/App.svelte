<script lang="ts">
  import Header from "./lib/components/Header.svelte";
  import type { IRawPayload } from "./lib/types";

  const raw: IRawPayload | null =
    (window.__FLOWD_DATA__ as IRawPayload) ?? null;
</script>

<div class="p-8">
  <Header
    generated={raw?.generated}
    period={raw?.period}
    hasData={raw != null}
  />
</div>
