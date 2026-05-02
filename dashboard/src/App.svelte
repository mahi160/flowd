<script>
  import Header from "./lib/components/Header.svelte";
  import Focus from "./lib/widgets/Summary/Focus.svelte";
  import Machine from "./lib/widgets/Summary/Machine.svelte";
</script>

<div class="flex gap-4 min-h-screen flex-col p-8 container mx-auto">
  <Header />
  <main class="flex-1">
    <section class="grid grid-cols-[1.5fr_1fr_1fr_1fr] gap-8">
      <Focus />
      <div class="flex flex-col gap-4">
        <Machine />
      </div>
      <!-- <CodeWidget />
      <AiSessionWidget /> -->
    </section>
  </main>
</div>
