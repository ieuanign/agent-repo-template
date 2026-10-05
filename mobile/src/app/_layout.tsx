// Uniwind needs this import in the root component; the file sits at the mobile
// root, outside every alias, so Tailwind scans everything beneath it.
import "@/../global.css";

import { QueryClientProvider } from "@tanstack/react-query";
import { useCalendars, useLocales } from "expo-localization";
import {
  DarkTheme,
  DefaultTheme,
  type ErrorBoundaryProps,
  Slot,
  ThemeProvider,
} from "expo-router";
import * as SplashScreen from "expo-splash-screen";
import { useEffect, useSyncExternalStore } from "react";
import { useColorScheme } from "react-native";
import { KeyboardProvider } from "react-native-keyboard-controller";
import { SafeAreaProvider } from "react-native-safe-area-context";
import { useCSSVariable } from "uniwind";
import { IntlProvider } from "use-intl";

import { queryClient } from "@/core/queryClient";
import { reportError } from "@/core/reporting";
import { messages } from "@/lib/i18n/messages";
import { resolveLocale } from "@/lib/i18n/utils";
import { useLocaleStore } from "@/state/locale";
import GlobalError from "@/ui/GlobalError";

// Held until the saved language has loaded, so the first frame is already in it.
SplashScreen.preventAutoHideAsync();

export function ErrorBoundary({ error, retry }: ErrorBoundaryProps) {
  useEffect(() => {
    reportError(error);
  }, [error]);

  return <GlobalError retry={retry} />;
}

function useNavigationTheme() {
  const base = useColorScheme() === "dark" ? DarkTheme : DefaultTheme;
  const [background, card, text, border, primary, notification] =
    useCSSVariable([
      "--color-background",
      "--color-card",
      "--color-foreground",
      "--color-border",
      "--color-primary",
      "--color-destructive",
    ]) as string[];
  return {
    ...base,
    colors: { background, card, text, border, primary, notification },
  };
}

export default function RootLayout() {
  const hydrated = useSyncExternalStore(
    useLocaleStore.persist.onFinishHydration,
    useLocaleStore.persist.hasHydrated,
  );
  const locale = resolveLocale(
    useLocaleStore((state) => state.locale),
    useLocales(),
  );
  const timeZone = useCalendars()[0].timeZone ?? undefined;
  const theme = useNavigationTheme();

  useEffect(() => {
    if (hydrated) SplashScreen.hideAsync();
  }, [hydrated]);

  if (!hydrated) return null;

  return (
    <SafeAreaProvider>
      <KeyboardProvider>
        <QueryClientProvider client={queryClient}>
          <IntlProvider
            locale={locale}
            messages={messages[locale]}
            timeZone={timeZone}
          >
            <ThemeProvider value={theme}>
              <Slot />
            </ThemeProvider>
          </IntlProvider>
        </QueryClientProvider>
      </KeyboardProvider>
    </SafeAreaProvider>
  );
}
