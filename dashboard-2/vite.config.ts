import { defineConfig } from "vite-plus";
import preact from "@preact/preset-vite";
import { viteSingleFile } from "vite-plugin-singlefile";
import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
  fmt: {},
  lint: { options: { typeAware: true, typeCheck: true } },
  plugins: [preact(), tailwindcss(), viteSingleFile()],

  build: {
    outDir: "../internal/fw/static",
    emptyOutDir: true,
  },
});
