import { render } from "preact";
import { App } from "./App";
import { transform } from "./lib/data";
import "./styles.css";

render(<App data={transform(window.__FLOWD_DATA__ || {})} />, document.getElementById("app")!);
