import { ParsedData } from "../types";
import { fmtHM } from "../lib/format";

export const Languages = ({ data }: { data: ParsedData }) => {
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
};
