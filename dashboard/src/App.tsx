import { Header } from "./lib/components/header";
import { DataProvider } from "./lib/context";
import { CodeWidget } from "./lib/widgets/code";
import { FocusWidget } from "./lib/widgets/focus";
import { MachineWidget } from "./lib/widgets/machine";
import { RepoWidget } from "./lib/widgets/repo";

export function App() {
  return (
    <DataProvider>
      <div className="flex gap-4 min-h-screen flex-col p-8 container mx-auto">
        <Header />
        <main className="flex-1">
          <section className="grid grid-cols-[1.5fr_1fr_1fr_1fr] gap-8">
            <FocusWidget />
            <div className="flex flex-col gap-4">
              <MachineWidget />
              <RepoWidget />
            </div>
            <CodeWidget />
          </section>
        </main>
      </div>
    </DataProvider>
  );
}
