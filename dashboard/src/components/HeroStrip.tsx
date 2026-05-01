import { ParsedData } from "../types";
import { BranchIcon, FlowMark } from "./icons";

const Sparkline = ({ data }: { data: number[] }) => {
  const pts = data
    .map((v, i) => `${(i / 23) * 180},${32 - (v / Math.max(1, ...data)) * 28}`)
    .join(" ");
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
};

export const HeroStrip = ({
  data,
  label,
}: {
  data: ParsedData;
  label: string;
}) => {
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
};
