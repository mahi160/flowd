import { createContext, type ComponentChildren } from "preact";
import type { IRawPayload } from "./types";
import { useContext, useEffect, useState } from "preact/hooks";
import { mockData } from "./mock";

interface DataContextType {
  data: IRawPayload | null;
  loading: boolean;
  error: Error | null;
}

const DataContext = createContext<DataContextType | undefined>(undefined);

export function DataProvider({ children }: { children: ComponentChildren }) {
  const [data, setData] = useState<IRawPayload | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    try {
      const globalData = (window as any).__FLOWD_DATA__;
      if (globalData) {
        setData(globalData);
      } else {
        setData(mockData);
      }
      setLoading(false);
    } catch (err) {
      setError(err instanceof Error ? err : new Error("Failed to load data"));
      setLoading(false);
    }
  }, []);

  return (
    <DataContext.Provider value={{ data, loading, error }}>
      {children}
    </DataContext.Provider>
  );
}

export function useData(): DataContextType {
  const context = useContext(DataContext);
  if (!context) {
    throw new Error("useData must be used within DataProvider");
  }
  return context;
}
