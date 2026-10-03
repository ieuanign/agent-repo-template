# @agent-repo-template/i18n

The locales and the catalogue text web and mobile both show.

- `messages/id.json` and `messages/en.json` hold the shared namespaces, `error` and `home`. Each app
  keeps the text only it shows in its own catalogue and merges the two.
- Private and unversioned (`"private": true`, version `0.0.0`); Changesets leaves it alone.
- Consumed as TypeScript source (`exports` → `./src/index.ts`): no build, no `dist/`, no tests.

## Exports

- `LOCALES` — the supported languages, `["en", "id"]`.
- `DEFAULT_LOCALE` — `"en"`.
- `Locale` — the type of one entry in `LOCALES`.
- `messages` — the shared catalogues keyed by `Locale`. `en` is typed against `id`, so an `en` missing
  a key fails `tsc`.

## Consuming it

Declare it in the consumer's `package.json` `dependencies` as `"@agent-repo-template/i18n": "*"`. Next.js
consumers list it in `transpilePackages`, since it ships TypeScript.

## Lint

`npm run lint -w packages/i18n` runs `tsc --noEmit`.
