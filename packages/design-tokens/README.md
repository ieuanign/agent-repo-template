# @agent-repo-template/design-tokens

The single source of design tokens for `web` (Tailwind v4) and `mobile` (Uniwind).

- One file, `theme.css`, holding a `@theme {}` block and a light and a dark theme (see
  [Rules](#rules-for-themecss)). The colour values are placeholders, except that light keeps web's
  original four; the spacing base is not (see [Spacing](#spacing)).
- Private and unversioned (`"private": true`, version `0.0.0`); Changesets leaves it alone.
- One script, `test`, running `node --test` on `theme.test.mjs`. `turbo run lint test` picks it up,
  so `make check` and the pre-commit hook run it.

## Consuming it

Declare it in the consumer's own `package.json` `dependencies`:

```json
"@agent-repo-template/design-tokens": "*"
```

npm links the workspace, and `turbo prune <workspace> --docker` keeps this package in the pruned
build context only when it is declared.

The bare specifier resolves through this package's `exports` (`./theme.css`) on both sides, so no
alias, `@source` entry or file copy is needed.

### Web (Next.js, Tailwind v4)

In `web/src/app/globals.css`, after `@import "tailwindcss";`:

```css
@import "tailwindcss";
@import "@agent-repo-template/design-tokens";
```

The same file declares the two custom variants the themes are written in, `@custom-variant light`
and `@custom-variant dark`. Each matches its own class on `<html>`, or, with no theme class there,
the device's `prefers-color-scheme`. Tailwind cannot compile `theme.css` without them.

### Mobile (Expo, Uniwind)

In `mobile/global.css` (the `cssEntryFile` given to `withUniwindConfig`), importing `tailwindcss`,
then `uniwind`, then the tokens:

```css
@import "tailwindcss";
@import "uniwind";
@import "@agent-repo-template/design-tokens";
```

Uniwind reads the `light` and `dark` variants as its own themes.

## Rules for `theme.css`

Mobile compiles this same file for native, so values must be React Native-safe:

- Exactly two top-level blocks:
  - `@theme { … }` registers every colour as `--color-<name>: unset;`, beside the font, radius and
    spacing tokens;
  - `@layer theme { :root { @variant light { … } @variant dark { … } } }` holds the colour values.
- Every colour is registered as `unset` in `@theme`; without it web emits no class for that colour.
- `light` and `dark` hold the same keys. The `test` script fails on an unregistered key or on a key
  in one theme only.
- Tailwind's own namespaces: `--color-*`, `--font-*`, `--radius-*`, `--spacing-*`, and the bare
  `--spacing` base.
- Hex colours (8-digit for alpha), `rem` lengths (never `px`), plain font-family names. No `oklch()`,
  `color-mix()`, `light-dark()`, `calc()`, `env()` or media queries.

## Spacing

Spacing sits on a 4px grid. The base is `--spacing: 0.25rem`, which is 4px on web, and on mobile
too, where Uniwind treats `rem` as 16px.

Every numeric spacing class is its number times the base: `p-4` is 1rem (16px), `gap-6` is 1.5rem
(24px), `mt-2` is 0.5rem (8px). The base matches Tailwind's default; this file states it anyway so
the grid is a decision recorded here rather than an upstream default.

Named spacing steps are added only when the design names them, and they take role names such as
`gutter` or `section`.

Tailwind's size words, `3xs` to `7xl` (including `xs`, `sm`, `md`, `lg` and `xl`), are never spacing
names. Tailwind's `w-*`, `min-w-*`, `max-w-*` and `basis-*` utilities read `--spacing-*` before
`--container-*`, so a spacing step named `md` would turn `max-w-md` from 28rem into that step.
