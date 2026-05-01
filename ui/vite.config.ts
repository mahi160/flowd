import { defineConfig } from "vite-plus";
import { svelte } from "@sveltejs/vite-plugin-svelte";
import tailwindcss from "@tailwindcss/vite";
import { viteSingleFile } from "vite-plugin-singlefile";

export default defineConfig({
  fmt: {},
  lint: { options: { typeAware: true, typeCheck: true } },
  plugins: [svelte(), tailwindcss(), viteSingleFile()],
  build: {
    target: "esnext",
    outDir: "../internal/fw/static",
    emptyOutDir: true,
  },
});
