import { expect, test } from "@playwright/test";

const ID_TAGLINE = "Monorepo full-stack yang siap dikerjakan agen sejak klon pertama.";
const EN_TAGLINE = "A full-stack monorepo an agent can work in from the first clone.";

test("defaults to English with no cookie and no Accept-Language", async ({
  page,
}) => {
  // Playwright's default locale fixture would otherwise send `en-US`.
  await page.route("**/*", (route) => {
    const headers = { ...route.request().headers() };
    delete headers["accept-language"];
    return route.continue({ headers });
  });

  const response = await page.goto("/");

  expect(response?.status()).toBe(200);
  await expect(page.locator("html")).toHaveAttribute("lang", "en");
  await expect(page.getByText(EN_TAGLINE)).toBeVisible();
});

test.describe("with Accept-Language: id", () => {
  test.use({ locale: "id" });

  test("follows the header", async ({ page }) => {
    await page.goto("/");

    await expect(page.locator("html")).toHaveAttribute("lang", "id");
    await expect(page.getByText(ID_TAGLINE)).toBeVisible();
  });

  test("prefers the locale cookie over the header", async ({
    page,
    context,
    baseURL,
  }) => {
    await context.addCookies([{ name: "locale", value: "en", url: baseURL }]);

    await page.goto("/");

    await expect(page.locator("html")).toHaveAttribute("lang", "en");
    await expect(page.getByText(EN_TAGLINE)).toBeVisible();
  });
});
