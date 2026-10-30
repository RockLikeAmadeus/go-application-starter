import { defineConfig, minimal2023Preset } from "@vite-pwa/assets-generator/config";

// `pnpm icons` makes the app's PNG icons and favicon from public/icon.svg.
export default defineConfig({
  preset: minimal2023Preset,
  images: ["public/icon.svg"],
});
