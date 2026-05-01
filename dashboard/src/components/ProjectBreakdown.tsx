import { ParsedData } from "../types";

export const ProjectBreakdown = ({ data }: { data: ParsedData }) => {
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
              <span style={{ width: `${(p.min / top) * 100}%`, background: p.color }} />
            </div>
            <span class="tnum lang-val">{p.min}m</span>
          </li>
        ))}
      </ul>
    </section>
  );
};
