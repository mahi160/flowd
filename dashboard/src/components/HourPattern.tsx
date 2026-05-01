import { ParsedData } from "../types";
import { fmtHour } from "../lib/format";

export const HourPattern = ({ data }: { data: ParsedData }) => {
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
};
