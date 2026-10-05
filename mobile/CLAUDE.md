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
| Icons         | an icon library, through the one `Icon` atom                     |
| Server state  | TanStack Query                                                   |
| Client state  | zustand, with `persist` on AsyncStorage                          |
| Forms         | TanStack Form                                                    |
| Fonts         | Inter, embedded by the expo-font config plugin                   |
| Lint          | oxlint                                                           |
| Format        | Prettier, one root config                                        |
| Types         | `tsc`, with no emit                                              |
| Unit tests    | jest-expo, React Native Testing Library with `userEvent`, MSW    |
| Configuration | `src/core/env.ts`                                                |

TanStack Form, `expo-image` and an icon library are not installed yet; their rows and rules bind the
first use.

## Layout

The tree is in [`README.md`](./README.md#project-structure).

- Each folder is created when it first holds something.
- `src/core/` holds constants, env, utilities and the TanStack Query client; `src/lib/` holds
  third-party setup; `src/test/` holds test helpers that are not mocks.
- A route file renders its `src/ui` screen and wires params; the markup lives in the screen.
- Route files export the default component only, plus `ErrorBoundary` or `unstable_settings` where
  expo-router reads them.

## Components

Atomic design. A component is as small as it can usefully be, and its **tier** decides where it lives:

| Tier         | What it is                                                              | Where              |
| ------------ | ----------------------------------------------------------------------- | ------------------ |
| **Atom**     | One element, no domain knowledge — `Button`, `Text`, an icon, a spinner | `src/components/`  |
| **Molecule** | A few atoms, one job — a field with its label and error, a toast        | `src/components/`  |
| **Organism** | A composed section — a header, a project summary panel                  | `src/components/`  |
| **Screen**   | One route's whole view, assembling the above                            | `src/ui/<Screen>/` |

**A component with more than one file lives in its own folder.** One file stays flat
(`ContentColumn.tsx`); a second file, such as its test or its styles, moves it into
`src/components/<Name>/` as `index.tsx`. A screen always has its folder, where its test, constants and
utils collect.

**A component starts local and moves on its second caller.** What one screen renders lives in that
screen's folder; it earns a place in `src/components/` when a second screen wants it.

**An atom knows nothing about the domain.** Domain vocabulary enters at the organism tier.

**Composition and fetching stay apart.** A screen calls hooks and composes; everything in
`src/components/` takes props and renders.

**Component and screen names are PascalCase** (`ContentColumn.tsx`, `Button/`), and a folder's own
component is its `index.tsx`. oxlint enforces it under `src/components/` and `src/ui/`. Routes keep expo-router's
lowercase names.

**A file copied from the kit is ours, and is edited before it is committed.** Every colour, radius
and font class names a token, the file goes into `src/components/` under a PascalCase name with its
`cva` variants split into `styles.ts`, and its Lucide imports are rewritten to go through `Icon`. `cn` is `src/lib/utils.ts`.

**A Button's label is a `Text` child**, and `Text` carries `font-sans`, so every label is Inter.

**`Icon` imports each icon by its own module path.**

## Imports

- Metro reads the `@/` and `@messages/` aliases from `tsconfig.json`'s `paths`, and Jest from `moduleNameMapper` in
  `jest.config.js`, so both lists change together.

## Naming

- Props, route params and `testID`s are camelCase. A `testID` carrying an id joins it with a dot:
  `photo.${id}`.
- Maestro flows in `e2e-mobile/` find elements by visible text. A `testID` is added only where text
  cannot identify an element, named by the rule above.

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

- `src/lib/i18n/messages.ts` merges the shared and own catalogues per language.

**Hermes and `Intl`**

- `@formatjs/intl-pluralrules` is the one polyfill: `polyfill-force` with the `id` and `en` data,
  imported first in `index.ts`.
- Numbers and dates use Hermes's own `NumberFormat` and `DateTimeFormat`, through the display
  presets.
- `Intl.Locale`, `RelativeTimeFormat`, `ListFormat`, `DisplayNames` and
  `DateTimeFormat.prototype.formatRange` are absent on Hermes. `jest.setup.ts` deletes them, so a
  test fails the way a phone would. A feature that needs one polyfills it then.

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

- Lint runs Prettier from the repository root, and writes only gitignored files.
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

## Shared with web

[`.claude/rules/apps/shared.md`](../.claude/rules/apps/shared.md) loads with every file under `mobile/`.
It holds what both apps follow: absolute imports, the file layout within a folder, the `Icon` atom,
camelCase props, the catalogue rules, the display presets and four re-render rules.
