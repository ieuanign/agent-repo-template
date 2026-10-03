# React Native on Expo for mobile

mobile is one React Native app on Expo, for every user type, on iOS and Android phones and tablets. On foldable and multi-screen phones it meets the lowest tier of Android's [large-screen app quality](https://developer.android.com/docs/quality-guidelines/large-screen-app-quality) guidelines, Tier 3 "Adaptive ready": it fills the window, keeps its state through a fold or unfold, and lays itself out by window width. Both frameworks we compared, React Native on Expo and Flutter, do this by default, so at this tier they need the same work, and React Native is the one that reuses what the repository already has. `theme.css` from [`packages/design-tokens`](../../packages/design-tokens/README.md) compiles as it is, [`packages/api-client`](../../packages/api-client) works unchanged, Changesets versions the app through `mobile/package.json` ([ADR 0003](./0003-independent-changesets-versioning.md)), web's translation catalogues load through `use-intl`, and devbox needs no new toolchain. The repository also stays on three languages: TypeScript, Go and Python.

## Considered Options

- **Flutter**: rejected. It leads only above Tier 3. It reports hinges and posture out of the box through `MediaQuery.displayFeatures`, and its shadcn-style kits are fuller. In exchange it adds Dart as a fourth language, plus glue that each needs its own staleness check:
  - a generator from `theme.css` to a Dart theme;
  - a Dart API client with its own envelope layer;
  - a bridge from `package.json`'s version to `--build-name`;
  - a conversion of the catalogues to ARB;
  - a Flutter toolchain;
  - Shorebird for over-the-air updates, which cannot be self-hosted.

  Revisit if hinge-aware or multi-pane layouts become a core product requirement, such as a field-work screen designed around a half-open phone.

## Consequences

- **Every screen is one fluid column** with a maximum content width, centred on wide screens. A screen gets a second pane only when the product asks for one, switching at Android's [window size class](https://developer.android.com/develop/ui/compose/layouts/adaptive/use-window-size-classes) breakpoints of 600dp and 840dp.
- **Orientation is never locked.** Android 16 ignores orientation locks on screens 600dp or wider for apps targeting API 36 ([Android 16 behaviour changes](https://developer.android.com/about/versions/16/behavior-changes-16)), which [Google Play requires](https://developer.android.com/google/play/requirements/target-sdk). A lock would then hold on a Fold's cover screen but not on its inner one. iPad multitasking also requires every orientation.
- **Tablets are supported.** iPad (`supportsTablet: true`) and Android tablets use the same layout, driven by window width.
- **A fold never restarts the app.** A config plugin adds `density` and `fontScale` to Android's `configChanges`, which Expo leaves out.
- **No hinge or posture code until a product screen needs it** (Tier 1). Then it is a local Expo module over Jetpack WindowManager that reports every fold, not just the first.
- **Developers test on a fixed list, not on every phone**:
  - Android Studio's resizable emulator;
  - the 7.6" Fold-in with outer display emulator;
  - one iPhone simulator;
  - one iPad simulator.

  Maestro runs on the same list.
- **HarmonyOS NEXT is not a target.** Huawei's foldables sold outside China run EMUI and take the Android build.
- **iPhone Duo needs Expo SDK 58**, which is in beta as of October 2026. The SDK version is chosen when the mobile skeleton is specified.
