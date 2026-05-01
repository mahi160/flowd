import { Card } from "../components/card";
import { useData } from "../context";

export function CodeWidget() {
  const { data } = useData();
  return (
    <Card header="Code" description="">
      <p className="font-display text-2xl font-light tabular-nums">
        {data?.files_changed} files
      </p>
      <div className="flex gap-2 items-center text-xs font-mono text-foreground/30 tabular-nums">
        <span className="text-accent">+{data?.lines_added}</span>/
        <span className="text-danger">-{data?.lines_removed}</span>
      </div>
      <div>
        <span>js</span>
      </div>
    </Card>
  );
}
