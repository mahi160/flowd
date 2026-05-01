import { Header } from "./lib/components/header";
import { DataProvider } from "./lib/context";

export function App() {
  return (
    <DataProvider>
      <div className="flex min-h-screen flex-col p-8 container mx-auto">
        <Header />
        <main className="flex-1" />
      </div>
    </DataProvider>
  );
}
