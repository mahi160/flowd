import { useCallback, useEffect, useMemo, useState } from "preact/hooks";
import type { ParsedData } from "./types";
import { Header } from "./components/Header";
import { HeroStrip } from "./components/HeroStrip";
import { ActivityHeatmap, WeekHeatmap } from "./components/Heatmap";
import { Insights } from "./components/Insights";
import { Donut } from "./components/Donut";
import { Languages } from "./components/Languages";
import { HourPattern } from "./components/HourPattern";
import { StreakCard } from "./components/StreakCard";
import { ProjectBreakdown } from "./components/ProjectBreakdown";
import { Timeline } from "./components/Timeline";
import { Summary } from "./components/Summary";
import { WeekBars } from "./components/WeekBars";
import { FlowMark } from "./components/icons";

type Theme = "dark" | "light" | "system";
const THEMES: Theme[] = ["dark", "light", "system"];
const resolveTheme = (t: Theme) =>
  t === "system"
    ? matchMedia("(prefers-color-scheme: dark)").matches
      ? "dark"
      : "light"
    : t;

function TodayView({ data }: { data: ParsedData }) {
  return (
    <main class="dashboard">
      <HeroStrip data={data} label="today" />
      <div class="grid-act">
        <ActivityHeatmap data={data} />
        <Insights data={data} />
      </div>
      <div class="grid-3">
        <Donut title="By project" subtitle="today" items={data.byProject} />
        <Donut
          title="By command"
          subtitle="today · top tools"
          items={data.byCommand.slice(0, 8)}
        />
        <Languages data={data} />
      </div>
      <div class="grid-mid">
        <HourPattern data={data} />
        <StreakCard data={data} />
        <ProjectBreakdown data={data} />
      </div>
      <div class="grid-bot">
        <Timeline data={data} />
        <Summary data={data} />
      </div>
    </main>
  );
}
function WeekView({ data }: { data: ParsedData }) {
  return (
    <main class="dashboard">
      <HeroStrip data={data} label="this week" />
      {data.weekDays.length > 0 && <WeekBars data={data} />}
      <WeekHeatmap data={data} />
      <div class="grid-3">
        <Donut title="By project" subtitle="this week" items={data.byProject} />
        <Donut
          title="By command"
          subtitle="this week"
          items={data.byCommand.slice(0, 8)}
        />
        <Languages data={data} />
      </div>
      <div class="grid-bot">
        <Timeline data={data} />
        <Summary data={data} />
      </div>
    </main>
  );
}
const EmptyState = () => (
  <main class="dashboard">
    <div class="card empty-state">
      <FlowMark size={48} />
      <h2>No activity yet</h2>
      <p>Start the daemon to begin tracking.</p>
      <p>
        <code>fw start</code>
      </p>
    </div>
  </main>
);
const WrongPeriod = ({ data, view }: { data: ParsedData; view: string }) => (
  <main class="dashboard">
    <div class="card empty-state">
      <h2>{view[0].toUpperCase() + view.slice(1)} view</h2>
      <p>This dashboard was generated for the {data.period} period.</p>
      <p>
        <code>fw dashboard {view}</code>
      </p>
    </div>
  </main>
);

export function App({ data }: { data: ParsedData }) {
  const [view, setView] = useState(data.period);
  const [theme, setTheme] = useState<Theme>(
    () => (localStorage.getItem("fw-theme") as Theme) || "system",
  );
  useEffect(() => {
    const apply = () => {
      document.documentElement.dataset.theme = resolveTheme(theme);
      localStorage.setItem("fw-theme", theme);
    };
    apply();
    const mq = matchMedia("(prefers-color-scheme: dark)");
    mq.addEventListener?.("change", apply);
    return () => mq.removeEventListener?.("change", apply);
  }, [theme]);
  const cycleTheme = useCallback(
    () => setTheme((t) => THEMES[(THEMES.indexOf(t) + 1) % THEMES.length]),
    [],
  );
  return (
    <div class="page">
      <Header
        data={data}
        view={view}
        setView={setView}
        theme={theme}
        cycleTheme={cycleTheme}
      />
      {!data.hasData ? (
        <EmptyState />
      ) : view === data.period ? (
        view === "week" ? (
          <WeekView data={data} />
        ) : (
          <TodayView data={data} />
        )
      ) : (
        <WrongPeriod data={data} view={view} />
      )}
      <footer class="foot">
        flowd<span class="dim"> — local activity tracker · self-hosted</span>
      </footer>
    </div>
  );
}
