<script>
  import Card from "../components/Card.svelte";
  import { flowd } from "../store.svelte";
  import {
    FileText,
    ArrowDownToLine,
    ArrowUpFromLine,
    CircleDollarSign,
  } from "lucide-svelte";

  const { ai_sessions } = flowd;

  const formatter = new Intl.NumberFormat("en-US", {
    notation: "compact",
    compactDisplay: "short",
  });

  const summary = $derived(
    ai_sessions.reduce(
      (acc, session) => {
        acc.tokens_read += session.tokens_read;
        acc.tokens_write += session.tokens_write;
        acc.cost += session.cost;
        acc.tools.add(session.tool);
        return acc;
      },
      {
        tokens_read: 0,
        tokens_write: 0,
        cost: 0,
        tools: new Set(),
      },
    ),
  );
</script>

<Card heading="AI Session Summary">
  <div class="grid grid-cols-2 gap-4">
    <div class="col-span-2 flex items-end gap-2">
      <FileText class="w-5 h-5 text-primary mt-0.5" />
      <p class="font-medium text-foreground mt-1">
        {Array.from(summary.tools).join(", ")}
      </p>
    </div>

    <div class="flex items-end gap-2">
      <ArrowDownToLine class="w-5 h-5 text-primary mt-0.5" />
      <p class="font-bold text-foreground mt-1">
        {formatter.format(summary.tokens_read)}
      </p>
    </div>

    <div class="flex items-end gap-2">
      <ArrowUpFromLine class="w-5 h-5 text-primary mt-0.5" />
      <p class="font-bold text-foreground mt-1">
        {formatter.format(summary.tokens_write)}
      </p>
    </div>

    <div class="col-span-2 flex items-end gap-2">
      <CircleDollarSign class="w-5 h-5 text-accent mt-0.5" />
      <p class="font-bold text-accent mt-1">${summary.cost.toFixed(4)}</p>
    </div>
  </div>
</Card>
