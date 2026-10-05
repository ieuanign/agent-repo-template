import type { Locale } from "@agent-repo-template/i18n";
import AsyncStorage from "@react-native-async-storage/async-storage";
import { create } from "zustand";
import { createJSONStorage, persist } from "zustand/middleware";

// Only a choice someone made; null follows the device.
interface LocaleState {
  locale: Locale | null;
}

export const useLocaleStore = create<LocaleState>()(
  persist((): LocaleState => ({ locale: null }), {
    name: "locale",
    storage: createJSONStorage(() => AsyncStorage),
  }),
);

export function setLocale(locale: Locale | null) {
  useLocaleStore.setState({ locale });
}
