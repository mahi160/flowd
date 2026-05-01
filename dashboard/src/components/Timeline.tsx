import { ParsedData } from "../types";
import { BranchIcon } from "./icons";

export const Timeline = ({ data }: { data: ParsedData }) => {
  const entries = [...data.timeline].reverse(), max = Math.max(1, ...entries.map(e => e.focus));
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
                <div class="tl-time tnum">{e.from} → {e.to}</div>
                <div class="tl-rail">
                  <span class="tl-dot" style={{ "--c": e.project ? "var(--accent)" : "var(--ink-5)" } as any} />
                  {i < entries.length - 1 && <span class="tl-line" />}
                </div>
                <div class="tl-body">
                  {!e.project ? (
                    <div class="tl-head"><span class="dim">— idle</span></div>
                  ) : (
                    <>
                      <div class="tl-head">
                        <span class="tl-project">{e.project}</span>
                        {e.branch && <span class="branch-tag"><BranchIcon size={10} /> {e.branch}</span>}
                      </div>
                      <div class="tl-meta">
                        <span class="tnum">{e.focus}m</span> focus
                        {e.switches > 0 && <><span class="dim"> · </span><span class="tnum">{e.switches}</span> context switches</>}
                      </div>
                      <div class="tl-bar"><span style={{ width: `${(e.focus / max) * 100}%` }} /></div>
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
};
