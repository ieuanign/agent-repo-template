import { expect, test } from "@playwright/test";

test("an unknown path answers 404 with the localised not-found page", async ({
  page,
  context,
  baseURL,
}) => {
  await context.addCookies([{ name: "locale", value: "id", url: baseURL }]);

  const response = await page.goto("/this-page-does-not-exist");

  expect(response?.status()).toBe(404);
  await expect(page.locator("html")).toHaveAttribute("lang", "id");
  await expect(
    page.getByRole("heading", { name: "Halaman tidak ditemukan" }),
  ).toBeVisible();
  await expect(
    page.getByText("Halaman yang Anda cari tidak ada atau telah dipindahkan."),
  ).toBeVisible();
});
