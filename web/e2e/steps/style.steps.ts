import { expect } from "@playwright/test";
import { createBdd } from "playwright-bdd";

const { When, Then } = createBdd();

// Opening the app, coming back later and the theme steps live in
// theme.steps.ts. The style is observed through the picker and the root's
// data-style attribute, which all style CSS keys off.

When(/^I switch to the (\w+) style$/, async ({ page }, style: string) => {
  await page.getByRole("combobox", { name: "Style" }).selectOption({ label: style });
});

Then(/^the app uses the (\w+) style$/, async ({ page }, style: string) => {
  await expect(page.getByRole("combobox", { name: "Style" })).toHaveValue(style.toLowerCase());
  await expect(page.locator("html")).toHaveAttribute("data-style", style.toLowerCase());
});
