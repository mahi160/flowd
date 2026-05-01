import { ParsedData } from "../types";

export const StreakCard = ({ data }: { data: ParsedData }) => {
  const s = data.streakDays, sub = s === 0 ? "start coding today" : s >= 14 ? "on fire 🔥" : s >= 7 ? "keep it up 🌿" : "building momentum";
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
          return <span key={cell.d} class="streak-pad" title={cell.d === 29 ? "today" : `${29 - cell.d}d ago`} style={{ background: `color-mix(in oklch, var(--moss) ${i * 85}%, var(--bg-inset))`, border: `1px solid color-mix(in oklch, var(--moss) ${i * 40}%, var(--hairline-soft))`, outline: cell.d === 29 ? "2px solid color-mix(in oklch, var(--moss) 70%, transparent)" : "", boxShadow: i > .15 ? `inset 0 1px 0 rgba(255,255,255,${(i * .12).toFixed(2)}), 0 0 ${(5 + i * 9).toFixed(1)}px ${(i * 2.5).toFixed(1)}px color-mix(in oklch, var(--moss) ${Math.round(i * 55)}%, transparent)` : "" }} />;
        })}
      </div>
      <div class="streak-axis"><span>30d ago</span><span>today</span></div>
    </section>
  );
};
