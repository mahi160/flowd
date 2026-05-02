import { useMemo } from "preact/hooks";
import { Sparkline } from "../charts/sparkline";
import { Card } from "../components/card";
import { useData } from "../context";
import { formatDuration } from "../helper";
import type { IRawPayload } from "../types";

export function FocusWidget() {
  const { data } = useData();
  console.log(data);

  return (
    <Card heading="Focus" description="">
      <FocusTime mins={data?.total_focus_min} />
      <span className="font-mono text-foreground/50 text-[10px]">
        {data?.total_blocks} focus blocks . {data?.total_switches} context
        switches
      </span>
      <FocusLineChart heatmap={data?.heatmap} period={data?.period} />
    </Card>
  );
}

const FocusTime = ({ mins }: { mins?: number }) => {
  const time = formatDuration(mins);
  return (
    <div className="flex gap-2">
      {Object.entries(time).map(
        ([key, value]) =>
          !!value && (
            <div key={key} className="flex items-baseline gap-1">
              <h1 className="text-5xl font-light text-primary font-display tabular-nums">
                {value}
              </h1>
              <span className="font-mono text-foreground/50">{key}</span>
            </div>
          ),
      )}
    </div>
  );
};

const FocusLineChart = ({
  heatmap,
  period,
}: {
  heatmap: IRawPayload["heatmap"];
  period: IRawPayload["period"];
}) => {
  const data = useMemo(() => {
    const heat = heatmap || [];
    const isToday = period !== "week";
    const hourly = Array(24).fill(0) as number[];

    for (const c of heat) {
      const h = isToday ? Math.floor(c.hour / 2) : c.hour;
      if (h >= 0 && h < 24) hourly[h] += c.minute;
    }
    return hourly;
  }, [heatmap]);
  return (
    <div className="mt-4">
      <Sparkline data={data} />
      <hr className="border-foreground/15" />
    </div>
  );
};
