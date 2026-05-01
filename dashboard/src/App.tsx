import type { Data } from "./types";
import { Header } from "./components/header";

export function App({ data }: { data: Data }) {
  return (
    <main>
      <Header data={data} />
    </main>
  );
}
