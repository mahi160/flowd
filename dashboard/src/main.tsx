import { render } from "preact";
import { App } from "./app.tsx";
import "./index.css";

declare global {
  interface Window {
    __FLOWD_DATA__?: Record<string, unknown>;
  }
}

const root = document.getElementById("app");
if (root) {
  render(<App data={(window as Window).__FLOWD_DATA__ || {}} />, root);
}
