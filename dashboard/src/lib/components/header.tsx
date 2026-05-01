import { useData } from "../context";
import { FlowdLogo } from "./icons";
import { ThemeToggle } from "./theme";

export function Header() {
  const { data } = useData();
  const periods = ["today", "week"];

  return (
    <header className="flex items-center justify-between">
      <section className="flex items-center gap-4">
        <FlowdLogo />

        <div className="flex flex-col gap-2">
          <span className="font-display text-3xl leading-none tracking-widest text-primary">
            flowd
          </span>

          <div className="flex items-center gap-3 font-mono text-xs text-foreground/50">
            <span className="uppercase tracking-[0.25em] text-[10px]">
              Generated
            </span>
            <span>|</span>
            <span className="tracking-tight">{data?.generated}</span>
          </div>
        </div>
      </section>

      <section className="flex items-center gap-3 text-xs font-medium">
        <div className="flex rounded-xl bg-surface p-1 shadow-sm">
          {periods.map((p) => (
            <button
              data-active={p === data?.period ? "" : undefined}
              className="capitalize rounded-lg data-active:bg-primary/8 px-4 py-1.5"
            >
              {p}
            </button>
          ))}
        </div>
        <ThemeToggle />
      </section>
    </header>
  );
}
