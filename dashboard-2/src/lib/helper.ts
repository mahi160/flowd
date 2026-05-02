export function formatDuration(mins?: number) {
  if (!mins) return { d: 0, h: 0, m: 0 };

  return {
    d: Math.floor(mins / 1440),
    h: Math.floor((mins % 1440) / 60),
    m: Math.floor(mins % 60),
  };
}
