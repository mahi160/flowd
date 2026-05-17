import { mount } from "svelte";
import App from "./App.svelte";
import { transform } from "./lib/transform";
import "./styles.css";

const data = transform(window.__FLOWD_DATA__ || {});

mount(App, { target: document.getElementById("app")!, props: { data } });
