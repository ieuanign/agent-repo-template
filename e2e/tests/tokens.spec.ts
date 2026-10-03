import { readFile } from "node:fs/promises";
import { expect, test } from "@playwright/test";

const schemes = ["light", "dark"] as const;

async function themeColour(
  scheme: (typeof schemes)[number],
  name: string,
): Promise<string> {
  const css = await readFile(
    new URL(import.meta.resolve("@agent-repo-template/design-tokens")),
    "utf8",
  );
  // Scoped to the scheme's block: @theme registers every colour as unset, and light comes first.
  const block = css.match(
    new RegExp(`@variant ${scheme}\\s*\\{([^}]*)\\}`),
  )?.[1];
  const hex = block?.match(
    new RegExp(`--color-${name}:\\s*#([0-9a-f]{6})\\s*;`, "i"),
  )?.[1];
  if (!hex)
    throw new Error(
      `--color-${name} is not a six-digit hex colour under @variant ${scheme} in theme.css`,
    );
  const [r, g, b] = [0, 2, 4].map((i) => parseInt(hex.slice(i, i + 2), 16));
  return `rgb(${r}, ${g}, ${b})`;
}

for (const scheme of schemes) {
  test.describe(`${scheme} colour scheme`, () => {
    test.use({ colorScheme: scheme });

    test.beforeEach(async ({ page }) => {
      await page.goto("/");
    });

    test("the heading is coloured with the primary token", async ({ page }) => {
      await expect(page.getByRole("heading", { level: 1 })).toHaveCSS(
        "color",
        await themeColour(scheme, "primary"),
      );
    });

    test("the body is painted with the background token", async ({ page }) => {
      await expect(page.locator("body")).toHaveCSS(
        "background-color",
        await themeColour(scheme, "background"),
      );
    });

    test("main is padded 16px on every side", async ({ page }) => {
      const main = page.locator("main");
      for (const side of ["top", "right", "bottom", "left"]) {
        await expect(main).toHaveCSS(`padding-${side}`, "16px");
      }
    });
  });
}
