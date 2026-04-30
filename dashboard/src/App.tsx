import { useCallback, useEffect, useMemo, useState } from "preact/hooks";
import type { Data, Item } from "./types";
import { arcPath, fmtHM, fmtHour } from "./lib/format";
import { BranchIcon, FlowMark, ICONS, SvgIcon } from "./components/icons";

type Theme = "dark" | "light" | "system";
const THEMES: Theme[] = ["dark", "light", "system"];
const resolveTheme = (t: Theme) =>
  t === "system"
    ? matchMedia("(prefers-color-scheme: dark)").matches
      ? "dark"
      : "light"
    : t;

function Sparkline({ data }: { data: number[] }) {
  const pts = useMemo(() => {
    const max = Math.max(1, ...data);
    return data
      .map((v, i) => `${(i / 23) * 180},${32 - (v / max) * 28}`)
      .join(" ");
  }, [data]);
  return (
    <svg class="sparkline" viewBox="0 0 180 36">
      <polyline
        points={pts}
        fill="none"
        stroke="var(--accent)"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
        opacity=".9"
      />
      <line x1="0" y1="34" x2="180" y2="34" stroke="var(--hairline-soft)" />
    </svg>
  );
}

function Header({
  data,
  view,
  setView,
  theme,
  cycleTheme,
}: {
  data: Data;
  view: string;
  setView: (v: string) => void;
  theme: Theme;
  cycleTheme: () => void;
}) {
  return (
    <header class="header">
      <div class="brand">
        <FlowMark />
        <div>
          <div class="brand-name font-display">flowd</div>
          <div class="brand-sub">
            <span class="eyebrow">Generated</span>
            <span class="brand-stamp tnum">{data.generated}</span>
            {data.hasData && (
              <>
                <span>·</span>
                <span class="chip">
                  <span class="status-dot" />
                  {data.period}
                </span>
              </>
            )}
          </div>
        </div>
      </div>
      <div class="header-right">
        <div class="seg">
          {["today", "week"].map((v) => (
            <button
              key={v}
              class={view === v ? "is-active" : ""}
              title={
                v !== data.period
                  ? `Run fw dashboard ${v} to generate ${v} data`
                  : ""
              }
              onClick={() => setView(v)}
            >
              {v === "today" ? "Today" : "Week"}
            </button>
          ))}
        </div>
        <button class="icon-btn" title={`Theme: ${theme}`} onClick={cycleTheme}>
          <SvgIcon html={ICONS[theme]} size={18} />
        </button>
      </div>
    </header>
  );
}

function HeroStrip({ data, label }: { data: Data; label: string }) {
  const f = data.focus,
    c = data.code,
    top = data.byProject[0];
  return (
    <section class="hero-strip">
      <div class="card hero-card hero-primary">
        <div class="eyebrow">Focus {label}</div>
        <div class="hero-num font-display tnum">
          {Math.floor(f.totalMin / 60)}
          <span class="hero-unit">h</span>
          {f.totalMin % 60}
          <span class="hero-unit">m</span>
        </div>
        <div class="hero-sub">
          <span class="tnum">{f.blocks}</span> focus blocks{" "}
          <span class="dim">·</span> <span class="tnum">{f.switches}</span>{" "}
          context switches
        </div>
        <Sparkline data={data.hourly} />
      </div>
      <div class="card hero-card">
        <div class="eyebrow">Machine</div>
        <div class="hero-mid font-display">{data.machine}</div>
        <div class="hero-sub">{data.os}</div>
        <div class="machine-bar">
          <span style={{ width: "50%" }} />
        </div>
      </div>
      <div class="card hero-card">
        <div class="eyebrow">Top repo</div>
        <div class="hero-mid font-display">{data.topRepo.name}</div>
        {data.topRepo.branch && (
          <div class="hero-sub">
            <span class="branch-tag">
              <BranchIcon /> {data.topRepo.branch}
            </span>
          </div>
        )}
        {top && (
          <div class="hero-mini tnum">
            {top.name} · {top.min}m
          </div>
        )}
      </div>
      <div class="card hero-card">
        <div class="eyebrow">Code</div>
        <div class="hero-mid font-display tnum">{c.files} files</div>
        <div class="hero-sub">
          <span class="tnum diff-add">+{c.added}</span>
          <span class="dim"> / </span>
          <span class="tnum diff-rm">−{c.removed}</span>
        </div>
      </div>
    </section>
  );
}

const HeatLegend = () => (
  <span class="heat-legend">
    less{" "}
    {[0.05, 0.3, 0.55, 0.8, 1].map((i) => (
      <span key={i} class="heat-cell" style={{ "--i": i } as any} />
    ))}{" "}
    more
  </span>
);
function ActivityHeatmap({ data }: { data: Data }) {
  let peak = 0,
    peakHour = 8;
  const cells = Array.from({ length: 60 }, (_, i) => {
    const hour = 8 + Math.floor((i * 12) / 60),
      v = data.hourly[Math.min(23, hour)] || 0,
      noise = ((i * 13) % 7) / 6;
    const intensity = Math.max(0, Math.min(1, (v / 45) * (0.5 + noise * 0.6)));
    if (intensity > peak) {
      peak = intensity;
      peakHour = hour;
    }
    return (
      <span
        key={i}
        class="heat-cell"
        style={{ "--i": intensity } as any}
        title={`~${hour}:00 — ≈${Math.round(intensity * 30)}m`}
      />
    );
  });
  return (
    <section class="card heatmap">
      <div class="card-head">
        <div class="card-title">Activity heatmap</div>
        <div class="card-sub">today · focus blocks by start time</div>
      </div>
      <div class="heat-axis">
        {["8a", "10a", "12p", "2p", "4p", "6p", "8p"].map((x) => (
          <span key={x}>{x}</span>
        ))}
      </div>
      <div class="heat-cells">{cells}</div>
      <div class="heat-foot">
        <HeatLegend />
        <span>{peak > 0 ? `peak ~${peakHour}:00` : ""}</span>
      </div>
    </section>
  );
}
function WeekHeatmap({ data }: { data: Data }) {
  const days = Object.entries(data.weekHourlyByDay),
    max = Math.max(1, ...days.flatMap(([, h]) => h));
  return (
    <section class="card heatmap">
      <div class="card-head">
        <div class="card-title">Activity heatmap</div>
        <div class="card-sub">last 7 days × hour of day</div>
      </div>
      <div class="week-heat-grid">
        {days.map(([day, hrs]) => (
          <div key={day} class="week-heat-row">
            <span class="week-heat-lbl">{day}</span>
            <div class="week-heat-cells">
              {hrs.map((v, h) => (
                <span
                  key={h}
                  class="heat-cell"
                  style={{ "--i": v / max } as any}
                  title={`${day} ${String(h).padStart(2, "0")}:00 · ${v}m`}
                />
              ))}
            </div>
          </div>
        ))}
      </div>
      <div class="heat-foot">
        <HeatLegend />
        <span>24h × 7 days</span>
      </div>
    </section>
  );
}

function Insights({ data }: { data: Data }) {
  const icon = (
    <svg width="15" height="15" viewBox="0 0 24 24" fill="none">
      <path
        d="M12 3v3M12 18v3M3 12h3M18 12h3M5.5 5.5l2 2M16.5 16.5l2 2M5.5 18.5l2-2M16.5 7.5l2-2"
        stroke="currentColor"
        strokeWidth="1.6"
        strokeLinecap="round"
      />
      <circle cx="12" cy="12" r="3" stroke="currentColor" strokeWidth="1.6" />
    </svg>
  );
  const Row = ({
    tag,
    cls = "",
    body,
    html = false,
  }: {
    tag: string;
    cls?: string;
    body: string;
    html?: boolean;
  }) => (
    <div class="insight-item">
      <span class={`insight-tag ${cls}`}>{tag}</span>
      <p dangerouslySetInnerHTML={html ? { __html: body } : undefined}>
        {html ? null : body}
      </p>
    </div>
  );
  if (data.aiRecap)
    return (
      <section class="card">
        <div class="insights-head">
          <div class="insights-icon">{icon}</div>
          <div class="card-title">AI recap</div>
          <span class="chip ml-auto">recap</span>
        </div>
        <Row tag="summary" cls="good" body={data.aiRecap} />
      </section>
    );
  if (data.aiPerBlock > 0)
    return (
      <section class="card">
        <div class="insights-head">
          <div class="insights-icon">{icon}</div>
          <div class="card-title">AI insights</div>
          <span class="chip ml-auto">
            {data.aiPerBlock} block{data.aiPerBlock === 1 ? "" : "s"}
          </span>
        </div>
        <Row
          tag="inline"
          body="Per-block AI summaries are inline in the timeline. Run <code>fw dashboard --ai-recap</code> for an aggregate."
          html
        />
      </section>
    );
  return (
    <section class="card">
      <div class="insights-head">
        <div class="insights-icon">{icon}</div>
        <div class="card-title">AI insights</div>
      </div>
      <Row
        tag="setup"
        body="Set <code>ai_enabled: true</code> and <code>ai_command</code> in your config to see AI insights here."
        html
      />
    </section>
  );
}

function Donut({
  title,
  subtitle,
  items,
}: {
  title: string;
  subtitle: string;
  items: Item[];
}) {
  const [hover, setHover] = useState<number | null>(null);
  const { segs, total } = useMemo(() => {
    let acc = -Math.PI / 2;
    const total = items.reduce((n, x) => n + x.min, 0);
    return {
      total,
      segs: items.map((item) => {
        const frac = total ? item.min / total : 0,
          s = acc,
          e = acc + frac * Math.PI * 2;
        acc = e;
        return { ...item, frac, s, e };
      }),
    };
  }, [items]);
  const active = hover == null ? null : segs[hover];
  return (
    <section class="card">
      <div class="card-head">
        <div class="card-title">{title}</div>
        {subtitle && <div class="card-sub">{subtitle}</div>}
      </div>
      {!items.length ? (
        <div>No data</div>
      ) : (
        <>
          <div class="donut-wrap">
            <svg class="donut" viewBox="0 0 220 220">
              {segs.map((seg, i) => (
                <path
                  key={seg.name}
                  d={arcPath(110, 110, 86, 56, seg.s, seg.e)}
                  fill={seg.color}
                  style={{ opacity: hover == null || hover === i ? 1 : 0.35 }}
                  onMouseEnter={() => setHover(i)}
                  onMouseLeave={() => setHover(null)}
                />
              ))}
              <circle
                cx="110"
                cy="110"
                r="55"
                fill="none"
                stroke="var(--hairline-soft)"
                strokeWidth="1"
              />
              <circle
                cx="110"
                cy="110"
                r="87"
                fill="none"
                stroke="var(--hairline-soft)"
                strokeWidth="1"
              />
            </svg>
            <div class="donut-center">
              {active ? (
                <>
                  <div class="donut-pct font-display tnum">
                    {Math.round(active.frac * 100)}%
                  </div>
                  <div class="donut-name">{active.name}</div>
                  <div class="donut-val tnum">{active.min}m</div>
                </>
              ) : (
                <>
                  <div class="donut-total font-display tnum">
                    {fmtHM(total)}
                  </div>
                  <div class="donut-cap eyebrow">total</div>
                </>
              )}
            </div>
          </div>
          <ul class="legend">
            {segs.map((seg, i) => (
              <li
                key={seg.name}
                class={`legend-row ${hover === i ? "is-hover" : ""}`}
                onMouseEnter={() => setHover(i)}
                onMouseLeave={() => setHover(null)}
              >
                <span class="legend-swatch" style={{ background: seg.color }} />
                <span class="legend-name">{seg.name}</span>
                <span class="legend-val tnum">{seg.min}m</span>
                <span class="legend-pct tnum">
                  {Math.round(seg.frac * 100)}%
                </span>
              </li>
            ))}
          </ul>
        </>
      )}
    </section>
  );
}

function Languages({ data }: { data: Data }) {
  const langs = data.byLanguage.slice(0, 8),
    top = langs[0]?.min || 1,
    total = langs.reduce((n, l) => n + l.min, 0);
  return (
    <section class="card">
      <div class="card-head">
        <div class="card-title">Languages</div>
        <div class="card-sub">by time in session</div>
      </div>
      {!langs.length ? (
        <p>No language data — inferred from git diffs.</p>
      ) : (
        <>
          <div class="lang-bar">
            {langs.map((l) => (
              <span
                key={l.name}
                class="lang-seg"
                style={{ flex: l.min, background: l.color }}
                title={`${l.name} · ${l.min}m`}
              />
            ))}
          </div>
          <ul class="lang-list">
            {langs.map((l) => (
              <li key={l.name}>
                <span class="legend-swatch" style={{ background: l.color }} />
                <span class="lang-name">{l.name}</span>
                <div class="lang-bar-mini">
                  <span
                    style={{
                      width: `${(l.min / top) * 100}%`,
                      background: l.color,
                    }}
                  />
                </div>
                <span class="tnum lang-val">{l.min}m</span>
              </li>
            ))}
          </ul>
          <div class="lang-foot">
            <span class="eyebrow">total</span>
            <span class="tnum lang-total font-display">
              {fmtHM(total)} <span class="dim">tracked</span>
            </span>
          </div>
        </>
      )}
    </section>
  );
}
function HourPattern({ data }: { data: Data }) {
  const max = Math.max(1, ...data.hourly),
    peak = data.hourly.indexOf(Math.max(...data.hourly)),
    first = data.hourly.findIndex((v) => v > 0),
    last = data.hourly.reduce((a, v, i) => (v > 0 ? i : a), -1);
  const stat = (k: string, v: string) => (
    <div key={k}>
      <span class="eyebrow">{k}</span>
      <span class="tnum">{v}</span>
    </div>
  );
  return (
    <section class="card">
      <div class="card-head">
        <div class="card-title">Best hours</div>
        <div class="card-sub">focus by hour of day</div>
      </div>
      <div class="hour-bars">
        {data.hourly.map((v, i) => (
          <div key={i} class="hour-bar">
            <span
              class="hour-fill"
              style={{ "--h": `${(v / max) * 100}%` } as any}
            />
            {i % 4 === 0 && <span class="hour-tick">{fmtHour(i)}</span>}
          </div>
        ))}
      </div>
      <div class="hour-stats">
        {stat("peak", fmtHour(peak))}
        {first >= 0 && stat("start", fmtHour(first))}
        {last >= 0 && stat("end", fmtHour(last))}
      </div>
    </section>
  );
}
function StreakCard({ data }: { data: Data }) {
  const s = data.streakDays,
    sub =
      s === 0
        ? "start coding today"
        : s >= 14
          ? "on fire 🔥"
          : s >= 7
            ? "keep it up 🌿"
            : "building momentum";
  return (
    <section class="card">
      <div class="card-head">
        <div class="card-title">Streak</div>
        <div class="card-sub">last 30 days</div>
      </div>
      <div class="streak-hero">
        <span class="streak-big font-display tnum">{s}</span>
        <div class="streak-hero-right">
          <span class="streak-hero-label">day streak</span>
          <span class="streak-hero-sub eyebrow">{sub}</span>
        </div>
      </div>
      <div class="streak-grid">
        {data.streakCells.map((cell) => {
          const i = cell.v / 4;
          return (
            <span
              key={cell.d}
              class="streak-pad"
              title={cell.d === 29 ? "today" : `${29 - cell.d}d ago`}
              style={{
                background: `color-mix(in oklch, var(--moss) ${i * 85}%, var(--bg-inset))`,
                border: `1px solid color-mix(in oklch, var(--moss) ${i * 40}%, var(--hairline-soft))`,
                outline:
                  cell.d === 29
                    ? "2px solid color-mix(in oklch, var(--moss) 70%, transparent)"
                    : "",
                boxShadow:
                  i > 0.15
                    ? `inset 0 1px 0 rgba(255,255,255,${(i * 0.12).toFixed(2)}), 0 0 ${(5 + i * 9).toFixed(1)}px ${(i * 2.5).toFixed(1)}px color-mix(in oklch, var(--moss) ${Math.round(i * 55)}%, transparent)`
                    : "",
              }}
            />
          );
        })}
      </div>
      <div class="streak-axis">
        <span>30d ago</span>
        <span>today</span>
      </div>
    </section>
  );
}
function ProjectBreakdown({ data }: { data: Data }) {
  const top = data.byProject[0]?.min || 1;
  return (
    <section class="card">
      <div class="card-head">
        <div class="card-title">By project</div>
        <div class="card-sub">time breakdown</div>
      </div>
      <ul class="lang-list project-list">
        {data.byProject.slice(0, 8).map((p) => (
          <li key={p.name}>
            <span class="legend-swatch" style={{ background: p.color }} />
            <span class="lang-name">{p.name}</span>
            <div class="lang-bar-mini">
              <span
                style={{
                  width: `${(p.min / top) * 100}%`,
                  background: p.color,
                }}
              />
            </div>
            <span class="tnum lang-val">{p.min}m</span>
          </li>
        ))}
      </ul>
    </section>
  );
}
function Timeline({ data }: { data: Data }) {
  const entries = [...data.timeline].reverse(),
    max = Math.max(1, ...entries.map((e) => e.focus));
  return (
    <section class="card">
      <div class="card-head">
        <div class="card-title">Timeline</div>
        <div class="card-sub">focus blocks · newest first</div>
      </div>
      {!entries.length ? (
        <p>No blocks recorded yet.</p>
      ) : (
        <div class="timeline-scroll">
          <ol class="timeline">
            {entries.map((e, i) => (
              <li key={i} class={`tl-row ${!e.project ? "tl-idle" : ""}`}>
                <div class="tl-time tnum">
                  {e.from} → {e.to}
                </div>
                <div class="tl-rail">
                  <span
                    class="tl-dot"
                    style={
                      {
                        "--c": e.project ? "var(--accent)" : "var(--ink-5)",
                      } as any
                    }
                  />
                  {i < entries.length - 1 && <span class="tl-line" />}
                </div>
                <div class="tl-body">
                  {!e.project ? (
                    <div class="tl-head">
                      <span class="dim">— idle</span>
                    </div>
                  ) : (
                    <>
                      <div class="tl-head">
                        <span class="tl-project">{e.project}</span>
                        {e.branch && (
                          <span class="branch-tag">
                            <BranchIcon size={10} /> {e.branch}
                          </span>
                        )}
                      </div>
                      <div class="tl-meta">
                        <span class="tnum">{e.focus}m</span> focus
                        {e.switches > 0 && (
                          <>
                            <span class="dim"> · </span>
                            <span class="tnum">{e.switches}</span> context
                            switches
                          </>
                        )}
                      </div>
                      <div class="tl-bar">
                        <span style={{ width: `${(e.focus / max) * 100}%` }} />
                      </div>
                      {e.ai && <div class="tl-ai">{e.ai}</div>}
                    </>
                  )}
                </div>
              </li>
            ))}
          </ol>
        </div>
      )}
    </section>
  );
}
function Summary({ data }: { data: Data }) {
  const top3 = (arr: Item[]) =>
    arr
      .slice(0, 3)
      .map((x) => `${x.name} ${x.min}m`)
      .join(" · ") || "—";
  const row = (term: string, value: any) => (
    <div key={term}>
      <dt>{term}</dt>
      <dd>{value}</dd>
    </div>
  );
  return (
    <section class="card">
      <div class="card-head">
        <div class="card-title">Summary</div>
        <div class="card-sub">{data.period} · narrative</div>
      </div>
      <div class="summary-scroll">
        <dl class="summary-list">
          {row(
            "Focus",
            <>
              <b class="tnum">{data.focus.totalMin}m</b> across{" "}
              <b class="tnum">{data.focus.blocks}</b> blocks
            </>,
          )}
          {data.byProject.length > 0 && row("Projects", top3(data.byProject))}
          {data.byCommand.length > 0 && row("Tools", top3(data.byCommand))}
          {row(
            "Code",
            <>
              <span class="tnum">{data.code.files} files</span> (
              <span class="tnum diff-add">+{data.code.added}</span>{" "}
              <span class="tnum diff-rm">−{data.code.removed}</span>)
            </>,
          )}
          {data.topRepo.name !== "—" &&
            row(
              "Top repo",
              <>
                <code>{data.topRepo.name}</code>
                {data.topRepo.branch && (
                  <>
                    {" "}
                    on <code>{data.topRepo.branch}</code>
                  </>
                )}
              </>,
            )}
        </dl>
      </div>
    </section>
  );
}
function WeekBars({ data }: { data: Data }) {
  const max = Math.max(1, ...data.weekDays.map((d) => d.min));
  return (
    <section class="card">
      <div class="card-head">
        <div class="card-title">Week at a glance</div>
        <div class="card-sub">focus per day</div>
      </div>
      <div class="week-bars">
        {data.weekDays.map((d) => (
          <div key={`${d.day}-${d.date}`} class="week-bar">
            <div class="week-track">
              <span
                class="week-fill"
                style={{ "--h": `${(d.min / max) * 100}%` } as any}
              >
                {d.min > 0 && (
                  <span class="week-fill-val tnum">{fmtHM(d.min)}</span>
                )}
              </span>
            </div>
            <div class="week-label">
              <div class="week-day">{d.day}</div>
              <div class="week-date tnum">{d.date}</div>
            </div>
          </div>
        ))}
      </div>
    </section>
  );
}

function TodayView({ data }: { data: Data }) {
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
function WeekView({ data }: { data: Data }) {
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
const WrongPeriod = ({ data, view }: { data: Data; view: string }) => (
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

export function App({ data }: { data: Data }) {
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
