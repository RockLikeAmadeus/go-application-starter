import { expect, type Page } from "@playwright/test";
import { createBdd } from "playwright-bdd";

const { Given, When, Then } = createBdd();

// These steps run against a server with the dev sign-in on (the sign-in
// project in playwright.config.ts). Each scenario's browser context starts
// with no cookies, so it starts signed out.

async function expectSignInScreen(page: Page) {
  await expect(page.getByRole("heading", { name: "Sign in" })).toBeVisible();
  await expect(page.getByRole("link", { name: "Sign in" })).toBeVisible();
}

Given("I am signed out", async ({ page }) => {
  await page.goto("/");
  await expectSignInScreen(page);
});

When("I sign in as {string}", async ({ page }, name: string) => {
  await page.getByRole("link", { name: "Sign in" }).click();
  // The dev sign-in's form, which stands in for a real sign-in provider.
  await page.getByLabel("Name").fill(name);
  await page.getByRole("button", { name: "Sign in" }).click();
});

Then("I see the items page", async ({ page }) => {
  await expect(page.getByRole("heading", { name: "Items" })).toBeVisible();
});

// The sidebar's content is also in the page as the closed phone menu, so these
// steps look in the sidebar, which is the one shown on a desktop.
Then("I am shown as signed in as {string}", async ({ page }, name: string) => {
  await expect(
    page.getByRole("complementary").getByText(`Signed in as ${name}.`, { exact: true }),
  ).toBeVisible();
});

When("I sign out", async ({ page }) => {
  await page.getByRole("complementary").getByRole("button", { name: "Sign out" }).click();
});

Then("I am asked to sign in", async ({ page }) => {
  await expectSignInScreen(page);
});
