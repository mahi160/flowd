import { transform } from "./data.js";
import { initTheme } from "./theme.js";
import { el, FlowMark } from "./utils.js";
import { Header } from "./components/header.js";
import { HeroStrip } from "./components/hero.js";
import { ActivityHeatmap, WeekHeatmap } from "./components/heatmap.js";
import { Insights } from "./components/insights.js";
import { Donut } from "./components/donut.js";
import { Languages } from "./components/languages.js";
import { HourPattern } from "./components/hours.js";
import { StreakCard } from "./components/streak.js";
import { Timeline } from "./components/timeline.js";
import { Summary } from "./components/summary.js";
import { WeekBars } from "./components/week.js";

// ── Data ──────────────────────────────────────────────────────
const DATA = transform(window.__RAW);

// ── Views ─────────────────────────────────────────────────────
function TodayView() {
  const main = el("main", "dashboard");
  main.append(
    HeroStrip(DATA, "today"),

    el("div", "grid-act", ActivityHeatmap(DATA), Insights(DATA)),

    el(
      "div",
      "grid-3",
      Donut("By project", "today", DATA.byProject),
      Donut("By command", "today · top tools", DATA.byCommand.slice(0, 8)),
      Languages(DATA),
    ),

    el(
      "div",
      "grid-mid",
      HourPattern(DATA),
      StreakCard(DATA),
      ProjectBreakdown(DATA),
    ),

    el("div", "grid-bot", Timeline(DATA), Summary(DATA)),
  );
  return main;
}

function WeekView() {
  const main = el("main", "dashboard");
  main.append(
    HeroStrip(DATA, "this week"),
    DATA.weekDays.length ? WeekBars(DATA) : el("div"),
    WeekHeatmap(DATA),

    el(
      "div",
      "grid-3",
      Donut("By project", "this week", DATA.byProject),
      Donut("By command", "this week", DATA.byCommand.slice(0, 8)),
      Languages(DATA),
    ),

    el("div", "grid-bot", Timeline(DATA), Summary(DATA)),
  );
  return main;
}

function ProjectBreakdown(DATA) {
  const card = el("div", "card");
  card.appendChild(
    el(
      "div",
      "card-head",
      el("div", "card-title", "By project"),
      el("div", "card-sub", "time breakdown"),
    ),
  );
  const list = el("ul", "lang-list");
  list.style.gap = "12px";
  const topMin = DATA.byProject[0]?.min || 1;
  DATA.byProject.slice(0, 8).forEach((p) => {
    const swatch = el("span", "legend-swatch");
    swatch.style.background = p.color;
    const fill = el("span");
    fill.style.width = `${(p.min / topMin) * 100}%`;
    fill.style.background = p.color;
    const bar = el("div", "lang-bar-mini", fill);
    const li = el(
      "li",
      null,
      swatch,
      el("span", "lang-name", p.name),
      bar,
      el("span", "tnum lang-val", p.min + "m"),
    );
    li.style.gridTemplateColumns = "10px 1fr 1fr auto";
    list.appendChild(li);
  });
  card.appendChild(list);
  return card;
}

function EmptyState() {
  return el(
    "div",
    "card empty-state",
    FlowMark(48),
    el("h2", null, "No activity yet"),
    el("p", null, "Start the daemon to begin tracking."),
    el("p", null, el("code", null, "fw start")),
  );
}

function WrongPeriod(requested) {
  const main = el("main", "dashboard");
  main.appendChild(
    el(
      "div",
      "card empty-state",
      el("h2", null, `${requested[0].toUpperCase() + requested.slice(1)} view`),
      el(
        "p",
        null,
        `This dashboard was generated for the ${DATA.period} period.`,
      ),
      el("p", null, el("code", null, `fw dashboard ${requested}`)),
    ),
  );
  return main;
}

function Footer() {
  return el(
    "footer",
    "foot",
    "flowd",
    el("span", "dim", " — local activity tracker · self-hosted"),
  );
}

// ── App ───────────────────────────────────────────────────────
let currentView = DATA.period;
const root = document.getElementById("app");

function render() {
  root.innerHTML = "";
  const page = el("div", "page");

  page.appendChild(
    Header(DATA, currentView, (v) => {
      currentView = v;
      render();
    }),
  );

  if (!DATA.hasData) {
    page.appendChild(EmptyState());
  } else if (currentView === DATA.period) {
    page.appendChild(currentView === "week" ? WeekView() : TodayView());
  } else {
    page.appendChild(WrongPeriod(currentView));
  }

  page.appendChild(Footer());
  root.appendChild(page);

  // Re-sync theme button icon after full re-render
  initTheme();
}

initTheme();
render();
