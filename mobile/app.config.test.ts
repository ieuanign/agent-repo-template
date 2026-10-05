import type { ConfigContext } from "expo/config";
import { AndroidConfig, type ExportedConfig } from "expo/config-plugins";

// A root file no tsconfig alias reaches; it sits beside this test because
// Expo's config loader resolves no alias, so it cannot move under src/.
import appConfig from "./app.config";

const config: ExportedConfig = appConfig({ config: {} } as ConfigContext);

// MainActivity as Expo's prebuild template writes it.
const templateManifest = () => ({
  manifest: {
    $: { "xmlns:android": "http://schemas.android.com/apk/res/android" },
    application: [
      {
        $: { "android:name": ".MainApplication" },
        activity: [
          {
            $: {
              "android:name": ".MainActivity",
              "android:configChanges":
                "keyboard|keyboardHidden|orientation|screenSize|screenLayout|uiMode|smallestScreenSize",
            },
          },
        ],
      },
    ],
  },
});

it("keeps MainActivity alive through the density and font-scale changes a fold makes", async () => {
  const manifestMod = config.mods?.android?.manifest;
  if (!manifestMod)
    throw new Error("app.config.ts registers no Android manifest mod");

  const result = await manifestMod({
    ...config,
    modResults: templateManifest(),
    modRequest: {},
  } as never);

  const activity = AndroidConfig.Manifest.getMainActivityOrThrow(
    result.modResults,
  );
  expect(activity.$["android:configChanges"]?.split("|")).toEqual(
    expect.arrayContaining(["uiMode", "density", "fontScale"]),
  );
});

it("supports iPad", () => {
  expect(config.ios?.supportsTablet).toBe(true);
});

it("leaves orientation unlocked", () => {
  expect(config.orientation).toBeUndefined();
});

it("follows the device's light or dark setting", () => {
  expect(config.userInterfaceStyle).toBe("automatic");
});
