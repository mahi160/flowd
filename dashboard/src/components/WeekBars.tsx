import { ParsedData } from "../types";
import { fmtHM } from "../lib/format";

export const WeekBars = ({ data }: { data: ParsedData }) => {
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
              <span class="week-fill" style={{ "--h": `${(d.min / max) * 100}%` } as any}>
                {d.min > 0 && <span class="week-fill-val tnum">{fmtHM(d.min)}</span>}
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
};
