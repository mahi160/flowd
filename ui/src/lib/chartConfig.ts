import type { ChartOptions } from 'chart.js';

const X_TICKS = {
  color: '#94a3b8',
  font: { size: 10, family: 'inherit' },
  maxRotation: 0,
} as const;

export function barChartOptions(
  tooltipLabel: (ctx: any) => string,
): ChartOptions<'bar'> {
  return {
    animation: false,
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: { display: false },
      tooltip: { callbacks: { label: tooltipLabel } },
    },
    scales: {
      x: { grid: { display: false }, ticks: X_TICKS, border: { display: false } },
      y: { display: false },
    },
  };
}
