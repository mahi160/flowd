import { useMemo, useState } from "preact/hooks";
import { Item } from "../types";
import { fmtHM, arcPath } from "../lib/format";

export const Donut = ({
  title,
  subtitle,
  items,
}: {
  title: string;
  subtitle: string;
  items: Item[];
}) => {
  const [hover, setHover] = useState<number | null>(null);

  const { segs, total } = useMemo(() => {
    let acc = -Math.PI / 2;
    const total = items.reduce((n, x) => n + x.min, 0);
    return {
      total,
      segs: items.map((item) => {
        const frac = total ? item.min / total : 0;
        const s = acc,
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
                  style={{
                    opacity: hover == null || hover === i ? 1 : 0.35,
                    transition: "opacity 0.2s",
                  }}
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
};
