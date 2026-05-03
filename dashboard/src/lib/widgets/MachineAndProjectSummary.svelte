<script lang="ts">
  import { GitBranch, Laptop } from "lucide-svelte";
  import Card from "../components/Card.svelte";
  import { flowd } from "../store.svelte";

  const OS_MAP: Record<string, string> = {
    darwin: "macos",
    linux: "linux",
  };
</script>

<div class="flex flex-col gap-4">
  <Card heading="Machine">
    <div class="flex flex-col">
      <p class="text-2xl font-light">{flowd?.machine ?? "Unknown"}</p>
      <span
        class="inline-flex items-center gap-1.5 px-2 py-0.5 w-fit h-fit rounded bg-warning/20 text-warning text-[10px] font-medium"
      >
        <Laptop size={12} />
        {OS_MAP[flowd?.os] ?? flowd?.os ?? "linux"}
      </span>
    </div>
  </Card>

  <Card heading="Top repo">
    <div class="flex flex-col">
      <p class="text-2xl font-light">{flowd?.top_repo ?? "None"}</p>
      <span
        class="inline-flex items-center gap-1.5 px-2 py-0.5 w-fit h-fit rounded bg-warning/20 text-warning text-[10px] font-medium"
      >
        <GitBranch size={12} />
        {flowd?.top_branch ?? "main"}
      </span>
    </div>
  </Card>
</div>
