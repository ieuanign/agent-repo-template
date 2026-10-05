# mobile

The React Native app on Expo, one app for every user type, on iOS and Android phones and tablets
([ADR 0006](../docs/adr/0006-react-native-on-expo-for-mobile.md)). It runs natively on the host
inside `devbox shell`, beside the Compose stack, and reaches backend through Traefik at
`localhost:4008`.

## Stack

| Concern       | Choice                                                                                                          |
| ------------- | --------------------------------------------------------------------------------------------------------------- |
| Framework     | an Expo SDK 58 development build, on React Native 0.88 and React 19.3                                           |
| Navigation    | expo-router 58, with a native stack                                                                             |
| Styling       | Uniwind on Tailwind CSS 4.3.3, with [`@agent-repo-template/design-tokens`](../packages/design-tokens/README.md) |
| Fonts         | Inter 400 and 500 from `@expo-google-fonts/inter`, embedded at build by the expo-font plugin                    |
| Server state  | TanStack Query                                                                                                  |
| Configuration | `environments/*.env` copied to `.env`; `EXPO_PUBLIC_*` values are inlined when Metro bundles                    |
| Lint          | oxlint, Prettier, `expo install --check` and `tsc` with no emit                                                 |
| Tests         | jest-expo, React Native Testing Library 14 with `userEvent`                                                     |
| Mocking       | MSW 2.15.0 through `msw/native`                                                                                 |

## Project structure

```
mobile/
├── app.config.ts           # the Expo config: identity, theme, fonts, splash and the fold plugin
├── app.config.test.ts      # checks the config's fold, tablet and theme settings
├── index.ts                # the entry, expo-router's
├── global.css              # Tailwind, Uniwind, then the shared tokens
├── jest.setup.ts           # copies dev.env into each test's env, mocks native modules, starts MSW
├── environments/           # dev, staging and production settings, copied to .env
├── assets/                 # the app icon, the adaptive icon layers and the splash image
├── scripts/                # lint steps run by `turbo run lint`
├── messages/               # mobile's own id and en catalogues
└── src/
    ├── app/                # routes only: the root layout and the (main) stack
    ├── ui/<Screen>/        # one screen per route, mirroring app/
    ├── components/         # the kit: ContentColumn, Text and Button
    ├── core/               # env, the http client, API_URLS, the TanStack Query client, reportError
    ├── lib/                # cn, and in i18n/ the merged catalogues and the language resolver
    ├── state/              # zustand stores: the saved language
    ├── mocks/              # the MSW server and its default handlers
    └── test/               # the query wrapper, and the stylesheet stub Jest maps `.css` imports to
```

`hooks/` is created when it first holds something.

The rules behind this layout are in [`CLAUDE.md`](./CLAUDE.md).

## Commands

Run from the repository root inside `devbox shell`.

| Command                      | Does                                                                               |
| ---------------------------- | ---------------------------------------------------------------------------------- |
| `make ios-build device=`     | Generates the iOS project, builds the dev client, installs it and starts Metro     |
| `make ios`                   | Starts Metro and opens the dev client already installed on the Simulator           |
| `make android-build device=` | Generates the Android project, builds the dev client, installs it and starts Metro |
| `make android`               | Starts Metro and opens the dev client already installed on the Emulator            |
| `make check`                 | Lints and tests every workspace, then smoke-tests the stack                        |
| `npm run lint -w mobile`     | oxlint, the Prettier check, `expo install --check`, then the route types and `tsc` |
| `npm run lint:fix -w mobile` | oxlint's autofix, then Prettier's write mode                                       |
| `npm test -w mobile`         | Runs Jest                                                                          |

The four `make` targets run against whatever stack is already up; start it with `make dev` when the
app needs backend.

## Host prerequisites

For mobile work only. devbox supplies Node, JDK 17, CocoaPods and Maestro.

- Xcode 26.4 or later, with an iOS Simulator runtime.
- Android Studio, with the Android SDK, the NDK, platform-tools, the emulator and an arm64 system
  image. `ANDROID_HOME` is set, and its `platform-tools` and `emulator` folders are on `PATH`.
- The fixed test devices:
  - Android Studio's resizable emulator;
  - the 7.6" Fold-in with outer display emulator, folded and unfolded;
  - one iPhone simulator;
  - one iPad simulator.

## Running on the Simulator, the Emulator and a phone

- `make ios-build` installs on a booted Simulator, and `make android-build` on the running Emulator.
  `device=` picks one by name or UDID, as `make ios-build device="<name or UDID>"`.
- After the first build, `make ios` and `make android` start Metro and open the installed dev client.
  A JavaScript change needs only those.
- The iOS targets run Xcode through [`scripts/xcode-env.sh`](../scripts/xcode-env.sh), which gives
  `xcodebuild` Apple's toolchain inside `devbox shell`.
- The Android targets run `adb reverse tcp:4008 tcp:4008` first, so `localhost:4008` on the device
  reaches the stack on the Mac. An Android phone on USB works through the same reverse.

## On a phone

A phone on Wi-Fi reaches backend at the Mac's LAN address:

1. Put the phone on the same Wi-Fi as the Mac.
2. Read the Mac's address with `ipconfig getifaddr en0`.
3. Write it into the gitignored `mobile/.env.local`, which overrides `.env` for one developer:
   `EXPO_PUBLIC_API_URL=http://<lan-ip>:4008/api`.
4. Restart Metro, then allow the Local Network prompt the first time the app opens, so it can reach
   Metro and backend.

## End-to-end tests

The Maestro suite in [`e2e-mobile/`](../e2e-mobile/) checks the built app end to end: it opens the
development build, waits for Home to show backend's version in either language, and takes a
screenshot. Today it runs the development build against `make dev`; with CI/CD it will run every
environment's build.

Run `make dev`, then `make ios-build` or `make android-build` (`make ios` or `make android` once the
app is installed), then `make e2e-mobile platform=ios` or `make e2e-mobile platform=android`.

- The suite runs on the host, not in Compose, and outside `make check`.
- `device=` takes a Simulator UDID or an Emulator serial. Without it, the target uses the one booted
  device, and lists them when there is not exactly one.
- The target checks, in order, that backend answers at `localhost:4008`, that Metro answers at
  `localhost:8081` and that the app is installed, and stops naming the command that fixes the first
  one missing. On Android it then re-runs both `adb reverse` rules before Maestro starts.
- Each run clears `e2e-mobile/.output/<platform>/` and writes under it one timestamped folder, which
  holds a `home` folder with `takeScreenshot/home.png`, the video `startRecording/home.mp4`,
  `commands.json`, `logs/maestro.log`, and `screenshots/final.png` when the flow fails. The video
  is recorded on a failed run too.
- To attach a run's evidence, drag the screenshot and the mp4 into a pull request comment; GitHub
  plays the mp4.
- Flows find elements by visible text; the selector rule is in [`CLAUDE.md`](./CLAUDE.md).

## Environment

- `environments/dev.env`, `staging.env` and `production.env` are committed, one per `APP_ENV`
  value. Every `make` target copies `dev.env` to the gitignored `mobile/.env`, which Expo reads for
  Metro and for `app.config.ts`.
- The keys:

  | Key                   | Holds                                               | dev                         | staging, production          |
  | --------------------- | --------------------------------------------------- | --------------------------- | ---------------------------- |
  | `EXPO_PUBLIC_API_URL` | backend's base URL through Traefik, `/api` included | `http://localhost:4008/api` | empty until hosting fills it |

  It is inlined when Metro bundles, so a change needs a Metro restart.

- Every value ships in the app bundle, so every value is public.
- Two moments fix a value into the app:
  - native config is read at prebuild, so a change to it needs `make ios-build` or
    `make android-build`;
  - `EXPO_PUBLIC_*` values and the tokens are fixed when Metro bundles, so a change to either needs
    a Metro restart.

## Continuous Native Generation

`ios/` and `android/` are generated by `expo prebuild` from `app.config.ts` and gitignored. Every
build target runs prebuild before `expo run:<platform>`, because `run:*` generates a project only
when its folder is missing. Pass `--clean` by hand to regenerate a project from scratch.
