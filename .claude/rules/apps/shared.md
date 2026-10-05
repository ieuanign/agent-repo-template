---
paths:
  - "web/**"
  - "mobile/**"
---

# web and mobile: what both follow

One copy of the rules the two apps share. Each app's own rules are in its `CLAUDE.md`. Where this file
says "view", mobile says "screen".

## Imports

- Imports are absolute: `@/` for `src/`, `@messages/` for the catalogues. Never a relative path, not
  even to a file beside the importer.

## File layout within a folder

The component file builds the component. Everything else has a home beside it:

```
src/ui/<View>/ or src/components/<Name>/
  index.tsx          # the view or component itself — a route renders its view's
  styles.ts          # the component's cva variants
  types.ts           # every type and interface the folder declares, props included
  hooks/             # the folder's own hooks
  constants.ts       # literals, lookup tables, tuning values
  utils.ts           # pure functions — formatting, parsing, clamping, deriving
  *.test.ts(x)       # colocated
```

- **Variants live in `styles.ts`**, so the component file names its classes and never declares them.
- **Constants live in `constants.ts`**; one shared across views moves to `src/core/`.
- **Pure functions live in `utils.ts`**, which is also what lets them be tested without rendering.
- **A folder's own hooks live in its `hooks/`**; one used by more than one view moves to
  `src/hooks/`.
- **Every `type` and `interface` a folder declares lives in its `types.ts`**, so a component file
  opens on the component. An inline annotation on a local helper's parameter stays where it is.

## Components

- **Icons come through one `Icon` atom**, `src/components/Icon`, the one place icons are imported. It
  re-exports the icons the app uses from its icon library and holds the app's own in
  `src/components/Icon/assets/*.tsx`. Neither app has it yet: the first icon creates it.
- **Props are camelCase.** An API field's name stops at the destructure that reads it.

## Catalogues

- Every user-facing string comes from a catalogue, one top-level namespace per view, named after its
  folder under `src/ui/` in camelCase (`ui/NotFound` → `notFound`). Namespaces and keys are both
  camelCase at every level.
- An app's own namespaces live in its `messages/id.json` and `messages/en.json`, each imported by a
  literal `@messages/` path; shared namespaces come from
  [`@agent-repo-template/i18n`](../../../packages/i18n/README.md). The app merges the two per language
  as `{...shared, ...own}`.
- Every key is in both languages; `tsc` fails when an `en` catalogue lacks a key its `id` has, in an
  app and in the package, and when an app's namespace has the same name as a shared one.
- Identical text is shared; text that differs stays in its app. A whole namespace starts in its app
  and moves to `@agent-repo-template/i18n` once the other app needs the same text; a namespace is never
  split, and moving one changes no call site. A shared namespace that loses one of its two apps moves
  back into the app still reading it, or is deleted if neither does, in the change that drops the
  last call.
- The namespace is always a string literal: `useTranslations("errorCode")`, or a sub-namespace such
  as `"errorCode.project"`. The key may be dynamic (`t(error.code)`). A root translator with a built
  path is never used. Review enforces it.
- `errorCode` arrives with the first view that shows a backend error, and lives in
  `@agent-repo-template/i18n` from its first use, because both apps show it.

## Display presets

Values are formatted through display presets, use-intl `formats` that live in
`@agent-repo-template/i18n` and are written with the first screen on either app that formats a value:

| Kind          | `id`           | `en`           | Pinned options                                                           |
| ------------- | -------------- | -------------- | ------------------------------------------------------------------------ |
| Rupiah        | `Rp 1.234.567` | `Rp 1,234,567` | `currency: "IDR"`, `currencyDisplay: "narrowSymbol"`, no fraction digits |
| Date          | `2 Okt 2026`   | `2 Oct 2026`   | medium date style                                                        |
| Time          | `10.04`        | `10:04`        | `hourCycle: "h23"`                                                       |
| Other numbers | `1.234`        | `1,234`        | grouping by language                                                     |

## Re-renders

- Subscribe to the narrowest thing: a derived boolean over the raw object, and nothing for a value
  only an event handler reads.
- Derived values compute during render; an effect that syncs one into state renders twice.
- Effect dependencies are primitives, not freshly built objects.
- A costly initial value takes the callback form: `useState(() => parse(raw))`.
