import { ParsedData } from "../types";
import { FlowMark, BranchIcon, ICONS, SvgIcon } from "./icons";

interface HeaderProps {
  data: ParsedData;
  view: string;
  setView: (v: string) => void;
  theme: "dark" | "light" | "system";
  cycleTheme: () => void;
}

export const Header = ({
  data,
  view,
  setView,
  theme,
  cycleTheme,
}: HeaderProps) => (
  <header class="header">
    <div class="brand">
      <FlowMark />
      <div>
        <div class="brand-name font-display">flowd</div>
        <div class="brand-sub">
          <span class="eyebrow">Generated</span>
          <span class="brand-stamp tnum">{data.generated}</span>
          {data.hasData && (
            <>
              <span>·</span>
              <span class="chip">
                <span class="status-dot" />
                {data.period}
              </span>
            </>
          )}
        </div>
      </div>
    </div>
    <div class="header-right">
      <div class="seg">
        {["today", "week"].map((v) => (
          <button
            key={v}
            class={view === v ? "is-active" : ""}
            title={
              v !== data.period
                ? `Run fw dashboard ${v} to generate ${v} data`
                : ""
            }
            onClick={() => setView(v)}
          >
            {v === "today" ? "Today" : "Week"}
          </button>
        ))}
      </div>
      <button class="icon-btn" title={`Theme: ${theme}`} onClick={cycleTheme}>
        <SvgIcon html={ICONS[theme]} size={18} />
      </button>
    </div>
  </header>
);
