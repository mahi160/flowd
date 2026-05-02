import { defineConfig } from "vite-plus";
import { svelte } from "@sveltejs/vite-plugin-svelte";
import { viteSingleFile } from "vite-plugin-singlefile";

export default defineConfig({
  fmt: {},
  lint: { options: { typeAware: true, typeCheck: true } },
  plugins: [svelte(), viteSingleFile()],
  build: {
    emptyOutDir: true,
    outDir: "../internal/fw/static",
  },
});
