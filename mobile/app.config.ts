import { readFileSync } from "node:fs";

import type { ConfigContext, ExpoConfig } from "expo/config";
import {
  AndroidConfig,
  type ConfigPlugin,
  withAndroidManifest,
} from "expo/config-plugins";

const themeCss = readFileSync(
  require.resolve("@agent-repo-template/design-tokens"),
  "utf8",
);

const background = (variant: "light" | "dark"): string => {
  const value = themeCss
    .split(`@variant ${variant}`)[1]
    ?.match(/--color-background:\s*([^;]+);/)?.[1];
  if (!value) {
    throw new Error(`theme.css has no ${variant} --color-background`);
  }
  return value.trim();
};

// npm hoists the font package to the repository root, out of reach of a path relative to mobile/.
const inter400 =
  require.resolve("@expo-google-fonts/inter/400Regular/Inter_400Regular.ttf");
const inter500 =
  require.resolve("@expo-google-fonts/inter/500Medium/Inter_500Medium.ttf");

// A fold changes density and font scale; without them Android restarts the activity.
const withFoldConfigChanges: ConfigPlugin = (config) =>
  withAndroidManifest(config, (config) => {
    const activity = AndroidConfig.Manifest.getMainActivityOrThrow(
      config.modResults,
    );
    const changes = new Set(
      activity.$["android:configChanges"]?.split("|") ?? [],
    );
    changes.add("density").add("fontScale");
    activity.$["android:configChanges"] = [...changes].join("|");
    return config;
  });

export default ({ config }: ConfigContext): ExpoConfig =>
  withFoldConfigChanges({
    ...config,
    name: "template",
    slug: "agent-repo-template",
    // No version here: Expo reads it from package.json,
    // so Changesets stays its one source.
    scheme: "agent-repo-template",
    userInterfaceStyle: "automatic",
    icon: "./assets/icon.png",
    ios: {
      bundleIdentifier: "com.agent.template",
      supportsTablet: true,
    },
    android: {
      package: "com.agent.template",
      adaptiveIcon: {
        foregroundImage: "./assets/android-icon-foreground.png",
        backgroundImage: "./assets/android-icon-background.png",
        monochromeImage: "./assets/android-icon-monochrome.png",
      },
    },
    experiments: { typedRoutes: true, reactCompiler: true },
    plugins: [
      "expo-router",
      [
        "expo-font",
        {
          ios: { fonts: [inter400, inter500] },
          android: {
            fonts: [
              {
                fontFamily: "Inter",
                fontDefinitions: [
                  { path: inter400, weight: 400 },
                  { path: inter500, weight: 500 },
                ],
              },
            ],
          },
        },
      ],
      [
        "expo-splash-screen",
        {
          image: "./assets/splash-icon.png",
          resizeMode: "contain",
          imageWidth: 76,
          backgroundColor: background("light"),
          dark: { backgroundColor: background("dark") },
        },
      ],
      // Lists the languages iOS and Android 13+ offer in the per-app language setting.
      ["expo-localization", { supportedLocales: ["en", "id"] }],
    ],
  });
