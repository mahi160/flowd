import type { IFlowdData } from "./types";

function loadData(): IFlowdData | null {
  const el = document.getElementById("flowd-data");

  if (!el?.textContent) {
    return null;
  }

  try {
    return JSON.parse(el.textContent);
  } catch {
    return null;
  }
}
function loadDevMock(): IFlowdData | null {
  const raw = import.meta.env.DEV ? import.meta.env.VITE_FLOWD_MOCK_DATA : null;
  if (!raw) return null;

  try {
    return JSON.parse(raw);
  } catch {
    return null;
  }
}
export const flowdData = $state<IFlowdData>(
  loadData() || loadDevMock() || ({} as IFlowdData),
);
