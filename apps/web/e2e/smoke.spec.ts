import { expect, test } from "@playwright/test";

test("the app shell renders", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByText("Apotekerza")).toBeVisible();
});
