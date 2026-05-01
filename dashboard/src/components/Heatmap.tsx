import { useMemo } from "preact/hooks";
import { ParsedData } from "../types";

const HeatLegend = () => (
  <span class="heat-legend">
    less{" "}
    {[0.05, 0.3, 0.55, 0.8, 1].map((i) => (
      <span key={i} class="heat-cell" style={{ "--i": i } as any} />
    ))}{" "}
    more
  </span>
);

export const ActivityHeatmap = ({ data }: { data: ParsedData }) => {
  const { cells, peak, peakHour } = useMemo(() => {
    let p = 0,
      ph = 8;
    const c = Array.from({ length: 60 }, (_, i) => {
      const hour = 8 + Math.floor((i * 12) / 60);
      const v = data.hourly[Math.min(23, hour)] || 0;
      const noise = ((i * 13) % 7) / 6;
      const intensity = Math.max(
        0,
        Math.min(1, (v / 45) * (0.5 + noise * 0.6)),
      );
      if (intensity > p) {
        p = intensity;
        ph = hour;
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
    return { cells: c, peak: p, peakHour: ph };
  }, [data.hourly]);

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
};

export const WeekHeatmap = ({ data }: { data: ParsedData }) => {
  const days = Object.entries(data.weekHourlyByDay);
  const max = useMemo(() => Math.max(1, ...days.flatMap(([, h]) => h)), [days]);

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
};
