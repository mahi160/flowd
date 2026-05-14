import tailwindcss from "@tailwindcss/vite";
import { svelte } from "@sveltejs/vite-plugin-svelte";
import { defineConfig } from "vite";
import { viteSingleFile } from "vite-plugin-singlefile";

export default defineConfig({
  plugins: [svelte(), viteSingleFile(), tailwindcss()],
  build: {
    emptyOutDir: true,
    outDir: "../internal/fw/static",
    minify: 'terser',
  },
  server: {
    port: 5173,
    strictPort: false,
  },
});
