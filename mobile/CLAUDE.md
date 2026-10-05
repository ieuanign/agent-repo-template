# mobile conventions

The React Native app on Expo, one app for every user type, on iOS and Android phones and tablets
([ADR 0006](../docs/adr/0006-react-native-on-expo-for-mobile.md)). What it has and how to run it is
in [`README.md`](./README.md).

## Libraries (hard rules)

| Concern       | Use                                                              |
| ------------- | ---------------------------------------------------------------- |
| Framework     | an Expo SDK 58 development build                                 |
| Navigation    | expo-router, with a native stack                                 |
| Styling       | Uniwind, with `@agent-repo-template/design-tokens`               |
| Components    | react-native-reusables on Uniwind, copied into `src/components/` |
| Keyboard      | react-native-keyboard-controller                                 |
| Translation   | use-intl, with `@agent-repo-template/i18n`                       |
| Animation     | react-native-reanimated                                          |
| Icons         | Tabler, through the one `Icon` atom                              |
| Server state  | TanStack Query                                                   |
| Client state  | zustand, with `persist` on AsyncStorage                          |
| Forms         | TanStack Form                                                    |
| Fonts         | Inter, embedded by the expo-font config plugin                   |
| Lint          | oxlint                                                           |
| Format        | Prettier, one root config                                        |
| Types         | `tsc`, with no emit                                              |
| Unit tests    | jest-expo, React Native Testing Library with `userEvent`, MSW    |
| Configuration | `src/core/env.ts`                                                |

## Layout

```
app.config.ts       # the Expo config and its test, beside index.ts and global.css
src/app/            # expo-router routes only: layouts and route files
src/ui/<Screen>/    # one route's screen, mirroring src/app
src/components/     # the kit: a one-file component flat, a larger one in its own folder
src/core/           # constants, env, utilities and the TanStack Query client
src/lib/            # third-party setup
src/lib/i18n/       # the merged catalogues and the language resolver
src/lib/utils.ts    # cn, the kit's class-name merge
src/state/          # zustand stores
src/mocks/          # the MSW server and its default handlers, for Jest
src/test/           # test helpers that are not mocks
src/hooks/          # hooks used by more than one screen
messages/           # mobile's own id and en catalogues; shared text comes from @agent-repo-template/i18n
environments/       # dev, staging and production settings
assets/             # icons and the splash image
```

- Each folder is created when it first holds something.
- A route file renders its `src/ui` screen and wires params; the markup lives in the screen.
- Route files export the default component only, plus `ErrorBoundary` or `unstable_settings` where
  expo-router reads them.

## Components

Atomic design. A component is as small as it can usefully be, and its **tier** decides where it lives:

| Tier         | What it is                                                             | Where              |
| ------------ | ---------------------------------------------------------------------- | ------------------ |
| **Atom**     | One element, no domain knowledge — `Button`, `Icon`, `Text`, `Spinner` | `src/components/`  |
| **Molecule** | A few atoms, one job — a field with its label and error, a toast       | `src/components/`  |
| **Organism** | A composed section — a header, a project summary panel                 | `src/components/`  |
| **Screen**   | One route's whole view, assembling the above                           | `src/ui/<Screen>/` |

**A component with more than one file lives in its own folder.** One file stays flat
(`ContentColumn.tsx`); a second file, such as its test or its styles, moves it into
`src/components/<Name>/` as `index.tsx`. A screen always has its folder, where its test, constants and
utils collect.

**A component starts local and moves on its second caller.** What one screen renders lives in that
screen's folder; it earns a place in `src/components/` when a second screen wants it.

**An atom knows nothing about the domain.** `Button` takes a label. Domain vocabulary enters at the
organism tier.

**Composition and fetching stay apart.** A screen calls hooks and composes; everything in
`src/components/` takes props and renders.

**Component and screen names are PascalCase** (`ContentColumn.tsx`, `Button/`), and a folder's own
component is its `index.tsx`. oxlint enforces it under `src/components/` and `src/ui/`. Routes keep expo-router's
lowercase names.

**A file copied from the kit is ours, and is edited before it is committed.** Every colour, radius
and font class names a token, the file goes into `src/components/` under a PascalCase name with its
`cva` variants split into `styles.ts`, and its Lucide imports are rewritten to Tabler. `cn` is `src/lib/utils.ts`.

**Button renders its label through `Text`**, and `Text` carries `font-sans`, so every label is Inter.

**Icons come through one `Icon` atom**, the one place icons are imported, each icon by its own
module path.

## File layout within a folder

The component file builds the component. Everything else has a home beside it:

```
src/ui/<Screen>/ or src/components/<Name>/
  index.tsx          # the screen or component itself — a route renders its screen's
  styles.ts          # the component's cva variants
  types.ts           # every type and interface the folder declares, props included
  hooks/             # the folder's own hooks
  constants.ts       # literals, lookup tables, tuning values
  utils.ts           # pure functions — formatting, clamping, deriving
  *.test.ts(x)       # colocated
```

- **Variants live in `styles.ts`**, so the component file names its classes and never declares them.
- **Constants live in `constants.ts`**; one shared across screens moves to `src/core/`.
- **Pure functions live in `utils.ts`**, which is also what lets them be tested without rendering.
- **A folder's own hooks live in its `hooks/`**; one used by more than one screen moves to
  `src/hooks/`.
- **Every `type` and `interface` a folder declares lives in its `types.ts`**, so a component file
  opens on the component. An inline annotation on a local helper's parameter stays where it is.

## Imports

- Imports are absolute: `@/` for `src/`, `@messages/` for the catalogues. Never a relative path,
  not even to a file beside the importer.
- Metro reads the aliases from `tsconfig.json`'s `paths`, and Jest from `moduleNameMapper` in
  `jest.config.js`, so both lists change together.

## Naming

- Props, route params and `testID`s are camelCase. A `testID` carrying an id joins it with a dot:
  `photo.${id}`.
- Maestro flows in `e2e-mobile/` find elements by visible text. A `testID` is added only where text
  cannot identify an element, named by the rule above.
- An API field's name stops at the destructure that reads it.
- Catalogue namespaces and keys are camelCase at every level (`home`, `error`, `notFound`).

## State

- **TanStack Query** owns server state.
- **zustand** owns client state. A store holds values, and its actions are exported beside it as
  plain functions calling the store's `setState`, so a component imports the action it needs and a
  caller outside a component reaches the same function.
- Component-local state stays in components.
- **The query is the one source of truth for whether a request is in flight.** A submitting button
  takes `loading={mutation.isPending}`; a reading screen takes its skeleton from `query.isPending`
  and its error surface from `isError`.
- **A descendant needing server data calls the hook itself.** TanStack Query dedupes by key, so the
  data is declared where it is read; a prop reaches the child that uses it and stops there. A child
  selects its own zustand slice the same way.
- **Where a query lives follows how often it is used:** used once, inline in the component that needs
  it; several times within one screen, a hook in `src/ui/<Screen>/hooks/`; across screens, a global
  hook in `src/hooks/`.

## Forms

TanStack Form is installed with the first form, and every form is one.

- **One `useForm` per form, however many screens it spans.** A wizard step is a set of fields on the
  form; a hook in the screen's `hooks/` builds it, and each step takes `form` and reads its own
  field.
- **Validation is declared on the field** (`validators={{ onChange }}`), and the rule itself is a
  pure function in `utils.ts`, so every negative case is tested without rendering.
- **The message a field shows is `field.state.meta.errors[0]`.**
- **A step is left only once its fields validate**: `form.validateField(name, "submit")` gates
  Continue.

## React Native practice

**Lists**

- Long lists are `FlatList`. Rows key off their id and take no inline object, array or arrow prop.
- Expensive work moves out of the row component.

**Rendering**

- **Conditional JSX uses a ternary**: `{isError ? <Toast /> : null}`. A falsy `0` or `""` on the
  left of `&&` renders as a raw text node and crashes.
- Every string sits inside a `Text`.

**Touch, navigation, images**

- `Pressable` for touch, expo-router's native stack for navigation, `expo-image` for images.

**Animation**

- Animation runs through react-native-reanimated alone, its worklets carrying the frames.
- It changes only `transform` and `opacity`, which run on the UI thread.

**Re-renders**

- Subscribe to the narrowest thing: a derived boolean over the raw object, and nothing for a value
  only an event handler reads.
- Derived values compute during render.
- Effect dependencies are primitives.
- A costly initial value takes the callback form: `useState(() => parse(raw))`.
- Independent requests run together (`Promise.all`).
- The React Compiler is on, so `memo`, `useMemo` and `useCallback` are written by hand only for a
  measured problem it misses.

## Screens on every device

- Orientation is never locked, and tablets, iPad included, use the same layout, driven by window
  width ([ADR 0006](../docs/adr/0006-react-native-on-expo-for-mobile.md)).
- Every screen's root is `ContentColumn` (`src/components/ContentColumn.tsx`): a `SafeAreaView` on
  all four edges, then a `KeyboardAwareScrollView`, then the column
  `w-full max-w-2xl self-center px-4 grow`. It always scrolls, so landscape, a Fold's outer display
  and large font scales never clip. The root layout holds the `SafeAreaProvider` and
  `KeyboardProvider` it needs.
- A list screen uses a `FlatList` carrying the same column classes.
- Sizes come from flex, or from `useWindowDimensions()` when a value is needed. The 600dp and 840dp
  breakpoints are defined with the first two-pane screen.
- GlobalError is the one screen off `ContentColumn`, because it renders outside the providers (see
  Errors below).
- A fold never restarts the app: the inline plugin in `app.config.ts` adds `density` and
  `fontScale` to `MainActivity`'s `configChanges`.
- The fixed test devices:
  - Android Studio's resizable emulator;
  - the 7.6" Fold-in with outer display emulator, folded and unfolded;
  - one iPhone simulator;
  - one iPad simulator.

## Language

- mobile shows English (`en`) by default and Bahasa Indonesia (`id`) when asked. The language is, in
  order:
  1. the saved choice, when it is `en` or `id`;
  2. otherwise the first entry of expo-localization's `useLocales()` whose language is `en` or `id`,
     with Android's `in` read as `id`;
  3. otherwise `en`.
- `resolveLocale` in `src/lib/i18n/utils.ts` is that order as a pure function. The root layout feeds
  it `useLocales()`, so a language change on the device re-renders the app.
- The expo-localization plugin in `app.config.ts` declares `supportedLocales: ["en", "id"]`, so iOS
  and Android 13+ offer a per-app language.
- The saved choice lives in `src/state/locale.ts` and holds only a choice someone made: `id`, `en`,
  or `null` to follow the device. It never holds the detected language. `setLocale` is called by
  nothing yet; the module that owns the signed-in user and a language switcher will write it.
- The root layout holds the splash screen up until the saved choice has loaded from AsyncStorage, so
  the first frame is already in the right language.
- The language never comes from an IP lookup.
- The time zone is the device's, from expo-localization's `useCalendars()`, passed as
  `IntlProvider`'s `timeZone`. Which zone a domain record's times use is decided with the first screen that shows one.

**Catalogues**

- Every user-facing string comes from a catalogue, one top-level namespace per screen, named after
  its folder under `src/ui/` in camelCase (`ui/NotFound` → `notFound`). Namespaces and keys are
  camelCase at every level.
- mobile's own namespaces live in `messages/id.json` and `messages/en.json`, each imported by a
  literal `@messages/` path. Shared namespaces come from
  [`@agent-repo-template/i18n`](../packages/i18n/README.md).
- `src/lib/i18n/messages.ts` merges the two per language as `{...shared, ...own}`. `tsc` fails when
  `en` lacks a key `id` has, and when a mobile namespace has a shared one's name.
- Identical text is shared; text that differs stays in its app. A whole namespace starts in its app
  and moves to `@agent-repo-template/i18n` once the other app needs the same text; a namespace is never
  split, and moving one changes no call site. A shared namespace that loses one of its two apps moves
  back into the app still reading it, or is deleted if neither does, in the change that drops the
  last call. web follows the same rules ([`../web/CLAUDE.md`](../web/CLAUDE.md)).
- The namespace is always a string literal: `useTranslations("errorCode")`, or a sub-namespace such
  as `"errorCode.project"`. The key may be dynamic (`t(error.code)`). A root translator with a built
  path is never used. Review enforces it.
- `errorCode` arrives with the first screen that shows a backend error, and is shared from its first
  use.

**Hermes and `Intl`**

- `@formatjs/intl-pluralrules` is the one polyfill: `polyfill-force` with the `id` and `en` data,
  imported first in `index.ts`.
- Numbers and dates use Hermes's own `NumberFormat` and `DateTimeFormat`, through the display
  presets.
- `Intl.Locale`, `RelativeTimeFormat`, `ListFormat`, `DisplayNames` and
  `DateTimeFormat.prototype.formatRange` are absent on Hermes. `jest.setup.ts` deletes them, so a
  test fails the way a phone would. A feature that needs one polyfills it then.

**Display presets**

Values are formatted through use-intl `formats` that live in `@agent-repo-template/i18n` and are written
with the first screen on either app that formats a value:

| Kind          | `id`           | `en`           | Pinned options                                                           |
| ------------- | -------------- | -------------- | ------------------------------------------------------------------------ |
| Rupiah        | `Rp 1.234.567` | `Rp 1,234,567` | `currency: "IDR"`, `currencyDisplay: "narrowSymbol"`, no fraction digits |
| Date          | `2 Okt 2026`   | `2 Oct 2026`   | medium date style                                                        |
| Time          | `10.04`        | `10:04`        | `hourCycle: "h23"`                                                       |
| Other numbers | `1.234`        | `1,234`        | grouping by language                                                     |

## Errors

However a screen fails, the user stays on a screen of mobile's own. Each route file below renders
its screen from `src/ui/`:

| Route file                         | Covers                                        | Screen           |
| ---------------------------------- | --------------------------------------------- | ---------------- |
| `_layout`'s `ErrorBoundary`        | an error in the root layout itself            | `ui/GlobalError` |
| `(main)/_layout`'s `ErrorBoundary` | an error thrown by a screen in the main stack | `ui/Error`       |
| `+not-found`                       | every URL no route matches                    | `ui/NotFound`    |

- `ui/GlobalError` renders outside every provider, so its copy lives in its `constants.ts` in both
  languages, English first, and it roots on a plain `SafeAreaView` and `ScrollView` carrying
  the column classes.
- `ui/Error` reads the shared `error` namespace; `ui/NotFound` reads mobile's `notFound`.
- Both retry buttons call the `retry` expo-router passes.
- These screens show fixed copy only, never an error's message or stack.
- `reportError` in `src/core/reporting.ts` is the one place a caught error leaves the app. Both
  boundaries call it from an effect keyed on the error.

## Tokens and theme

- Styles name only the tokens from `@agent-repo-template/design-tokens`. `global.css` imports
  `tailwindcss`, then `uniwind`, then the tokens, and holds nothing else: no reset, because one after
  Uniwind's import drops utilities silently. Review enforces the tokens-only rule.
- No hex colour literal appears under `mobile/`, in code, config or docs.
- The app follows the device's light or dark setting through `userInterfaceStyle: "automatic"`,
  with `expo-system-ui` installed so Android development builds honour it.
- The navigation theme's colours come from the tokens through Uniwind's `useCSSVariable`, in the
  root layout.
- The splash screen's light and dark background colours are read from `theme.css` by
  `app.config.ts` at config time.
- Inter is embedded at 400 and 500, so `font-sans font-medium` is Inter Medium. Another weight is
  one more expo-font plugin entry; a new family also needs a `--font-<role>` token.
- `uniwind-types.d.ts` is generated and committed, for `className` typing.
- Editing the tokens needs a Metro restart.

## Configuration

- `src/core/env.ts` is the only app file that reads `process.env`; oxlint's `node/no-process-env`
  enforces it. Each value is a literal `process.env.EXPO_PUBLIC_X` read, with no fallback, and a
  missing or empty value throws at import, naming the key.
- The rule's override in `.oxlintrc.json` lists two test-side files besides it: `jest.setup.ts`
  copies `dev.env` into each test file's `process.env`, and `src/core/env.test.ts` sets the key to
  exercise the throw.
- Every value ships in the app bundle, so every value is public.
- Native config is fixed at prebuild; `EXPO_PUBLIC_*` values are fixed when Metro bundles.

## Backend calls

- `http` in `src/core/http.ts` is the one client, built once from `apiUrl` by `createApiClient`.
- Every path comes from `API_URLS` in `src/core/apiUrls.ts`, typed against the generated `paths`;
  a path is never written as a literal, tests and handlers included.
- Callers read `payload` themselves; nothing unwraps a response. Only `data`'s presence means
  success, and a query function rejects when it is absent.
- A failure the screen expects is handled inline on that screen; an unexpected one is thrown to the
  `(main)` boundary.
- The query placement rule in State above holds: inline when used once, the screen's `hooks/` for
  several uses on one screen, `src/hooks/` across screens.
- MSW handlers answer with backend's full envelope, typed with `Envelope<T>` from
  `@agent-repo-template/api-client`.
- Plain HTTP is allowed by Expo's development builds alone. Never set `expo-build-properties`'
  `android.usesCleartextTraffic` and never add a cleartext manifest plugin: both open release builds
  too.

## Native projects

- `ios/` and `android/` are generated by Continuous Native Generation and gitignored.
- Every build target runs `expo prebuild` before `expo run:<platform>`. `--clean` is passed by hand.
- A native change (a native module, a config plugin, `app.config.ts`) needs `make ios-build` or
  `make android-build`; a JavaScript change needs only Metro.

## Tests and lint

- `npm run lint -w mobile` runs `scripts/lint.sh`: oxlint, `prettier --check mobile` from the
  repository root, `expo install --check` offline, `expo customize tsconfig.json` for the route
  types, then `tsc --noEmit`. It writes only gitignored files.
- `lint:fix` runs `oxlint --fix`, then `prettier --write mobile` from the repository root. The
  pre-commit hook runs it before its gate.
- Expo-governed packages are installed with `npx expo install`, so `expo install --check` passes.
- Jest runs through the `jest-expo` preset. Tests are colocated as `*.test.ts(x)`.
- `jest.setup.ts` copies `environments/dev.env` into each test file's `process.env`, deletes the `Intl` members Hermes lacks and loads
  the plural polyfill, mocks AsyncStorage, react-native-keyboard-controller
  and react-native-safe-area-context, and starts one MSW server from `src/mocks/server.ts` at module
  scope with `onUnhandledRequest: "error"`, resetting its handlers before each test and closing it
  after the file.
- MSW mocks the network; a test overrides a default handler with `server.use`.
- Interaction goes through `userEvent` first.
- Uniwind applies no styles under Jest, so tests assert behaviour and text.
- A translated component is wrapped in use-intl's `IntlProvider` with the real `id` catalogue from
  `src/lib/i18n/messages.ts`, never a mock of use-intl.
- A route test renders through `await renderRouter(...)` from `expo-router/testing-library`.
- A test whose outcome depends on the device language mocks expo-localization's hooks.
