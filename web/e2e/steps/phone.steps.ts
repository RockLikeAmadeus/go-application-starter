import { expect } from "@playwright/test";
import { createBdd } from "playwright-bdd";

const { Given, When, Then } = createBdd();

// The device comes from the Playwright project (@phone scenarios run on
// emulated phones), so "on my phone" only guards against a misconfigured
// project. Style and theme steps live in style.steps.ts and theme.steps.ts.

const menuButton = { name: "Menu" } as const;

Given("I open the app on my phone", async ({ page }) => {
  expect(page.viewportSize()?.width, "phone scenarios need a phone-sized project").toBeLessThan(
    768,
  );
  await page.goto("/");
});

When("I open the menu", async ({ page }) => {
  await page.getByRole("button", menuButton).click();
});

When("I go to the home page from it", async ({ page }) => {
  await page.getByRole("dialog").getByRole("link", { name: "Starter" }).click();
});

When("I press Escape", async ({ page }) => {
  await page.keyboard.press("Escape");
});

When("I tap outside the menu", async ({ page }) => {
  // The drawer slides in from the left, so the right edge is backdrop.
  const { width, height } = page.viewportSize()!;
  await page.mouse.click(width - 5, height / 2);
});

Then("I can open the menu", async ({ page }) => {
  await expect(page.getByRole("button", menuButton)).toBeVisible();
});

Then("there is no menu", async ({ page }) => {
  await expect(page.getByRole("button", menuButton)).toBeHidden();
});

Then("I see the navigation", async ({ page }) => {
  await expect(page.getByRole("navigation")).toBeVisible();
});

Then("the navigation is hidden", async ({ page }) => {
  await expect(page.getByRole("navigation")).toBeHidden();
});

Then("I see the Appearance controls", async ({ page }) => {
  await expect(page.getByRole("group", { name: "Appearance" })).toBeVisible();
});

Then("the page does not scroll sideways", async ({ page }) => {
  // A string expression, because the e2e tsconfig has no DOM types.
  const overflow = await page.evaluate<number>(
    "document.documentElement.scrollWidth - document.documentElement.clientWidth",
  );
  expect(overflow).toBeLessThanOrEqual(0);
});
