/// <reference types="vitest/config" />
import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";
import { VitePWA } from "vite-plugin-pwa";

// The installed app's window and splash colour: the light theme's background
// (frost, in src/styles/palette.css). index.html's theme-color meta repeats it.
const appBackground = "#e6eaee";

// The installable app: its manifest, and a service worker that caches the
// built app (not the API) and updates itself quietly. See SPEC.md §4.5.
const pwa = VitePWA({
  registerType: "autoUpdate",
  workbox: {
    // Offline, page requests get the cached app, except for these: the API
    // must always reach the server (the app never shows a stored copy of
    // data), and so must the health probes and sign-in.
    navigateFallbackDenylist: [/^\/api\//, /^\/auth\//, /^\/healthz$/, /^\/readyz$/],
    // The Wails runtime is only for the Android app, which runs no service
    // worker, so browsers shouldn't download it when they install the app.
    globIgnores: ["**/wails-runtime-*.js"],
  },
  manifest: {
    name: "Starter",
    short_name: "Starter",
    display: "standalone",
    start_url: "/",
    theme_color: appBackground,
    background_color: appBackground,
    icons: [
      { src: "pwa-64x64.png", sizes: "64x64", type: "image/png" },
      { src: "pwa-192x192.png", sizes: "192x192", type: "image/png" },
      { src: "pwa-512x512.png", sizes: "512x512", type: "image/png" },
      {
        src: "maskable-icon-512x512.png",
        sizes: "512x512",
        type: "image/png",
        purpose: "maskable",
      },
    ],
  },
});

// In development the Go server runs on :8080 and Vite proxies API calls to it.
// In production the Go server serves the built files directly.
export default defineConfig({
  plugins: [react(), tailwindcss(), pwa],
  build: {
    rolldownOptions: {
      output: {
        // A fixed name for the Wails runtime's chunk, so the service worker
        // can leave it out (see globIgnores).
        codeSplitting: { groups: [{ name: "wails-runtime", test: /@wailsio[\\/]runtime/ }] },
      },
    },
  },
  server: {
    port: 5173,
    proxy: {
      "/api": process.env.API_URL ?? "http://localhost:8080",
    },
  },
  test: {
    environment: "jsdom",
    setupFiles: ["./src/test/setup.ts"],
    include: ["src/**/*.test.{ts,tsx}"],
    css: false,
  },
});
