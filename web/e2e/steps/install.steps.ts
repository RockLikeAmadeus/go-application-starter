import { expect } from "@playwright/test";
import { createBdd } from "playwright-bdd";

const { Given, When, Then } = createBdd();

// "Installable" is checked the way a browser decides it: a linked manifest
// with a name, a standalone display, a start URL and loadable PNG icons.
// "Without a connection" is the browser context going offline, so only the
// service worker can answer.

Then("the app can be installed as {string}", async ({ page, request }, name: string) => {
  const link = page.locator('link[rel="manifest"]');
  await expect(link, "the page links a web app manifest").toHaveCount(1);
  const href = await link.getAttribute("href");
  const response = await request.get(new URL(href!, page.url()).href);
  expect(response.ok()).toBe(true);
  expect(response.headers()["content-type"]).toContain("application/manifest+json");

  const manifest = await response.json();
  expect(manifest).toMatchObject({ name, short_name: name, display: "standalone", start_url: "/" });
  for (const size of ["192x192", "512x512"]) {
    const icon = manifest.icons?.find(
      (i: { sizes: string; type: string }) => i.sizes === size && i.type === "image/png",
    );
    expect(icon, `a ${size} PNG icon`).toBeTruthy();
    const image = await request.get(new URL(icon.src, response.url()).href);
    expect(image.headers()["content-type"]).toBe("image/png");
  }
});

Given("I have opened the app before", async ({ page }) => {
  await page.goto("/");
  // The first visit installs the service worker; wait until it controls the page.
  await page.waitForFunction("navigator.serviceWorker?.controller != null", null, {
    timeout: 10_000,
  });
});

When("I open the app without a connection", async ({ page, context }) => {
  await context.setOffline(true);
  await page.reload();
});

When("the connection comes back", async ({ context }) => {
  await context.setOffline(false);
});

const offlineNotice = { hasText: /offline/i } as const;

Then("the app says it is offline", async ({ page }) => {
  await expect(page.getByRole("status").filter(offlineNotice)).toBeVisible();
});

Then("the app no longer says it is offline", async ({ page }) => {
  await expect(page.getByRole("status").filter(offlineNotice)).toHaveCount(0);
});
