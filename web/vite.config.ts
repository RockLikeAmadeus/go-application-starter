/// <reference types="vitest/config" />
import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";
import { VitePWA } from "vite-plugin-pwa";

// The installed app's window and splash colour: the light theme's background
// (frost, in src/styles/palette.css). index.html's theme-color meta repeats it.
const appBackground = "#e6eaee";

// In development the Go server runs on :8080 and Vite proxies API calls to it.
// In production the Go server serves the built files directly.
export default defineConfig({
  plugins: [
    react(),
    tailwindcss(),
    VitePWA({
      injectRegister: false,
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
    }),
  ],
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
