import { readFile } from "node:fs/promises";
import { type Download, expect, type Page } from "@playwright/test";
import { createBdd } from "playwright-bdd";

const { Given, When, Then } = createBdd();

// The file each page downloaded last, for the step that reads it.
const exported = new WeakMap<Page, Download>();

// Scenarios run in parallel against one server, so they use distinct item
// names and assert on presence rather than on the whole list.

Given("I am on the items page", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByRole("heading", { name: "Items" })).toBeVisible();
});

When("I add an item named {string}", async ({ page }, name: string) => {
  await page.getByLabel("New item name").fill(name);
  await page.getByRole("button", { name: "Add item" }).click();
});

Given("I have added an item named {string}", async ({ page }, name: string) => {
  await page.getByLabel("New item name").fill(name);
  await page.getByRole("button", { name: "Add item" }).click();
  await expect(page.getByRole("table", { name: "Items" })).toContainText(name);
});

When("I rename {string} to {string}", async ({ page }, from: string, to: string) => {
  await page.getByRole("button", { name: `Rename ${from}` }).click();
  await page.getByLabel(`New name for ${from}`).fill(to);
  await page.getByRole("button", { name: "Save" }).click();
});

When("I delete {string}", async ({ page }, name: string) => {
  await page.getByRole("button", { name: `Delete ${name}` }).click();
});

Then("I see {string} in the item list", async ({ page }, name: string) => {
  await expect(
    page.getByRole("table", { name: "Items" }).getByText(name, { exact: true }),
  ).toBeVisible();
});

Then("I do not see {string} in the item list", async ({ page }, name: string) => {
  await expect(page.getByText(name, { exact: true })).toHaveCount(0);
});

Then("I see the error {string}", async ({ page }, message: string) => {
  await expect(page.getByRole("alert")).toContainText(message);
});

When("I export the items", async ({ page }) => {
  const download = page.waitForEvent("download");
  await page.getByRole("button", { name: "Export", exact: true }).click();
  // Kept on the page so the next step can read the file the browser saved.
  exported.set(page, await download);
});

Then("the downloaded file lists {string}", async ({ page }, name: string) => {
  const download = exported.get(page);
  if (!download) throw new Error("no file was downloaded");
  expect(download.suggestedFilename()).toBe("items.csv");
  const csv = await readFile(await download.path(), "utf8");
  expect(csv.split("\n")[0]).toBe("name,createdAt");
  expect(csv).toContain(name);
});
