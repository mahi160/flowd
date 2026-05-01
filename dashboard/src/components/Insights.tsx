import { ParsedData } from "../types";

export const Insights = ({ data }: { data: ParsedData }) => {
  const icon = (
    <svg width="15" height="15" viewBox="0 0 24 24" fill="none">
      <path
        d="M12 3v3M12 18v3M3 12h3M18 12h3M5.5 5.5l2 2M16.5 16.5l2 2M5.5 18.5l2-2M16.5 7.5l2-2"
        stroke="currentColor"
        strokeWidth="1.6"
        strokeLinecap="round"
      />
      <circle cx="12" cy="12" r="3" stroke="currentColor" strokeWidth="1.6" />
    </svg>
  );

  const Row = ({
    tag,
    cls = "",
    body,
    html = false,
  }: {
    tag: string;
    cls?: string;
    body: string;
    html?: boolean;
  }) => (
    <div class="insight-item">
      <span class={`insight-tag ${cls}`}>{tag}</span>
      <p dangerouslySetInnerHTML={html ? { __html: body } : undefined}>
        {html ? null : body}
      </p>
    </div>
  );

  if (data.aiRecap)
    return (
      <section class="card">
        <div class="insights-head">
          <div class="insights-icon">{icon}</div>
          <div class="card-title">AI recap</div>
          <span class="chip ml-auto">recap</span>
        </div>
        <Row tag="summary" cls="good" body={data.aiRecap} />
      </section>
    );

  if (data.aiPerBlock > 0)
    return (
      <section class="card">
        <div class="insights-head">
          <div class="insights-icon">{icon}</div>
          <div class="card-title">AI insights</div>
          <span class="chip ml-auto">
            {data.aiPerBlock} block{data.aiPerBlock === 1 ? "" : "s"}
          </span>
        </div>
        <Row
          tag="inline"
          body="Per-block AI summaries are inline in the timeline. Run <code>fw dashboard --ai-recap</code> for an aggregate."
          html
        />
      </section>
    );

  return (
    <section class="card">
      <div class="insights-head">
        <div class="insights-icon">{icon}</div>
        <div class="card-title">AI insights</div>
      </div>
      <Row
        tag="setup"
        body="Set <code>ai_enabled: true</code> and <code>ai_command</code> in your config to see AI insights here."
        html
      />
    </section>
  );
};
