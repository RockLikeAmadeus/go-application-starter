import { expect } from "@playwright/test";
import { createBdd } from "playwright-bdd";

const { Given, When, Then } = createBdd();

// Each scenario gets a fresh browser context, so "first time" means no
// stored choice. The theme is observed through the root's computed
// color-scheme, which is what the browser itself uses for native controls.

Given(/^my system prefers a (dark|light) scheme$/, async ({ page }, scheme: "dark" | "light") => {
  await page.emulateMedia({ colorScheme: scheme });
});

When("I open the app for the first time", async ({ page }) => {
  await page.goto("/");
});

When("I come back to the app later", async ({ page }) => {
  await page.reload();
});

When(/^I switch to the (dark|light) theme$/, async ({ page }, theme: string) => {
  await page.getByRole("switch", { name: "Dark mode" }).setChecked(theme === "dark");
});

Then(/^the app uses the (dark|light) theme$/, async ({ page }, theme: string) => {
  await expect(page.getByRole("switch", { name: "Dark mode" })).toBeChecked({
    checked: theme === "dark",
  });
  await expect(page.locator("html")).toHaveCSS("color-scheme", theme);
});
