<script>
  import Card from "../components/Card.svelte";
  import { Layers, Terminal, FileCode } from "lucide-svelte";

  let { languages = [], projects = [], commands = [] } = $props();
</script>

<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
  <Card heading="Languages">
    <div class="flex items-center gap-2 mb-4 text-gray-500">
      <FileCode class="w-4 h-4" />
    </div>
    <ul class="space-y-2">
      {#each languages.slice(0, 6) as lang}
        <li class="flex justify-between text-sm">
          <span>{lang.name}</span>
          <span class="font-mono text-gray-500">{lang.min}m</span>
        </li>
      {/each}
    </ul>
  </Card>

  <Card heading="By Project">
    <div class="flex items-center gap-2 mb-4 text-gray-500">
      <Layers class="w-4 h-4" />
    </div>
    <ul class="space-y-2">
      {#each projects.slice(0, 6) as proj}
        <li class="flex justify-between text-sm">
          <span>{proj.name}</span>
          <span class="font-mono text-gray-500">{proj.min}m</span>
        </li>
      {/each}
    </ul>
  </Card>

  <Card heading="By Tool">
    <div class="flex items-center gap-2 mb-4 text-gray-500">
      <Terminal class="w-4 h-4" />
    </div>
    <ul class="space-y-2">
      {#each commands.slice(0, 6) as cmd}
        <li class="flex justify-between text-sm">
          <span class="font-mono">{cmd.name}</span>
          <span class="font-mono text-gray-500">{cmd.min}m</span>
        </li>
      {/each}
    </ul>
  </Card>
</div>
