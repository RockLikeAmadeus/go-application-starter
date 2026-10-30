import { defineConfig, devices } from "@playwright/test";
import { cucumberReporter, defineBddConfig } from "playwright-bdd";

// E2E tests run against the real Go server with the built SPA embedded, so
// they exercise the same artifact that ships in the Docker image.
// Build first: `task e2e` does `pnpm build` before running this.
const PORT = Number(process.env.E2E_PORT ?? 4173);
const baseURL = process.env.E2E_BASE_URL ?? `http://localhost:${PORT}`;

const testDir = defineBddConfig({
  features: "e2e/features/**/*.feature",
  steps: "e2e/steps/**/*.ts",
});

export default defineConfig({
  testDir,
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  reporter: [
    ["list"],
    ["html", { open: "never" }],
    cucumberReporter("html", { outputFile: "playwright-report/cucumber.html" }),
  ],
  use: {
    baseURL,
    trace: "on-first-retry",
    // Animations off: tests stay fast and deterministic, and this covers the
    // reduced-motion path users can choose.
    reducedMotion: "reduce",
  },
  // @phone scenarios run only on the emulated phones; everything else runs on
  // desktop Chromium. WebKit on an iPhone is the closest we get to Safari on Linux.
  projects: [
    { name: "chromium", use: { ...devices["Desktop Chrome"] }, grepInvert: /@phone/ },
    { name: "phone", use: { ...devices["Pixel 7"] }, grep: /@phone/ },
    { name: "iphone", use: { ...devices["iPhone 15"] }, grep: /@phone/ },
  ],
  // Skip starting a server when pointing at an existing deployment (smoke tests).
  webServer: process.env.E2E_BASE_URL
    ? undefined
    : {
        command: "go run -tags embedweb ./cmd/server",
        cwd: "..",
        url: `${baseURL}/readyz`,
        reuseExistingServer: !process.env.CI,
        env: { PORT: String(PORT), STARTER_STORAGE: "memory" },
        timeout: 120_000,
      },
});
