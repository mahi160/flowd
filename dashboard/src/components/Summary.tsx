import { ParsedData } from "../types";
import { fmtHM } from "../lib/format";

export const Summary = ({ data }: { data: ParsedData }) => {
  const top3 = (arr: any[]) => arr.slice(0, 3).map((x) => `${x.name} ${x.min}m`).join(" · ") || "—";
  const row = (term: string, value: any) => (
    <div key={term}>
      <dt>{term}</dt>
      <dd>{value}</dd>
    </div>
  );

  return (
    <section class="card">
      <div class="card-head">
        <div class="card-title">Summary</div>
        <div class="card-sub">{data.period} · narrative</div>
      </div>
      <div class="summary-scroll">
        <dl class="summary-list">
          {row("Focus", <><b class="tnum">{data.focus.totalMin}m</b> across <b class="tnum">{data.focus.blocks}</b> blocks</>)}
          {data.byProject.length > 0 && row("Projects", top3(data.byProject))}
          {data.byCommand.length > 0 && row("Tools", top3(data.byCommand))}
          {row("Code", <><span class="tnum">{data.code.files} files</span> (<span class="tnum diff-add">+{data.code.added}</span> <span class="tnum diff-rm">−{data.code.removed}</span>)</>)}
          {data.topRepo.name !== "—" && row("Top repo", <><code>{data.topRepo.name}</code>{data.topRepo.branch && <> on <code>{data.topRepo.branch}</code></>}</>)}
        </dl>
      </div>
    </section>
  );
};
