import { mount } from "svelte";
import App from "./App.svelte";
import "./styles.css";

// Pass the raw payload to App so it can re-transform reactively per period.
mount(App, {
  target: document.getElementById("app")!,
  props: { raw: window.__FLOWD_DATA__ || {} },
});
