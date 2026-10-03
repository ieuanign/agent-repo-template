import assert from "node:assert";
import { readFileSync } from "node:fs";
import { test } from "node:test";

const css = readFileSync(new URL("./theme.css", import.meta.url), "utf8");

// Blocks hold no nested braces, so each ends at its first "}".
function colourKeys(text, opener, valuePattern = "[^;]+") {
  const start = text.indexOf(opener);
  assert.notStrictEqual(start, -1, `missing ${opener}`);
  const block = text.slice(start, text.indexOf("}", start));
  const keyPattern = new RegExp(`--color-([\\w-]+):\\s*${valuePattern};`, "g");
  return new Set([...block.matchAll(keyPattern)].map((match) => match[1]));
}

function keyErrors(text) {
  const registered = colourKeys(text, "@theme {", "unset");
  const themes = {
    light: colourKeys(text, "@variant light {"),
    dark: colourKeys(text, "@variant dark {"),
  };
  const errors = [];
  for (const [theme, keys] of Object.entries(themes)) {
    for (const name of keys) {
      if (!registered.has(name))
        errors.push(
          `${theme}: --color-${name} is not registered as unset in @theme`,
        );
    }
  }
  for (const [theme, other] of [
    ["light", "dark"],
    ["dark", "light"],
  ]) {
    for (const name of themes[theme]) {
      if (!themes[other].has(name))
        errors.push(`--color-${name} is in ${theme} but not in ${other}`);
    }
  }
  return errors;
}

test("the real theme.css has no key errors", () => {
  assert.deepStrictEqual(keyErrors(css), []);
});

test("a theme key without an unset registration is reported", () => {
  const mutated = css.replace("  --color-ring: unset;\n", "");
  assert.notStrictEqual(mutated, css);
  assert.deepStrictEqual(keyErrors(mutated), [
    "light: --color-ring is not registered as unset in @theme",
    "dark: --color-ring is not registered as unset in @theme",
  ]);
});

test("a key held by one theme only is reported", () => {
  const mutated = css.replace("      --color-input: #ffffff26;\n", "");
  assert.notStrictEqual(mutated, css);
  assert.deepStrictEqual(keyErrors(mutated), [
    "--color-input is in light but not in dark",
  ]);
});
