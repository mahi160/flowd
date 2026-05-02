import { Laptop } from "lucide-react";
import { Card } from "../components/card";
import { useData } from "../context";

const OS_MAP: Record<string, string> = {
  darwin: "macos",
  linux: "linux",
};
export function MachineWidget() {
  const { data } = useData();
  return (
    <Card heading="Machine" description="">
      <p className="font-display text-2xl font-light">{data?.machine}</p>
      <span
        className="summary-widget-badge"
        style={{ "--bg": "var(--color-warning)" }}
      >
        <Laptop size={12} />
        {OS_MAP[data?.os || "linux"]}
      </span>
    </Card>
  );
}
