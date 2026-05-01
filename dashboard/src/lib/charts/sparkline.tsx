import { useMemo } from "preact/hooks";

export function Sparkline({ data }: { data: number[] }) {
  const pts = useMemo(() => {
    const max = Math.max(1, ...data);
    return data
      .map((v, i) => `${(i / 23) * 180},${32 - (v / max) * 28}`)
      .join(" ");
  }, [data]);
  return (
    <svg class="sparkline" viewBox="0 0 180 36">
      <polyline
        points={pts}
        fill="none"
        stroke="var(--primary)"
        strokeWidth="1"
        strokeLinecap="round"
        strokeLinejoin="round"
        opacity=".8"
      />
    </svg>
  );
}
