import AsyncStorage from "@react-native-async-storage/async-storage";

import { setLocale, useLocaleStore } from "@/state/locale";

beforeEach(async () => {
  await AsyncStorage.clear();
  setLocale(null);
});

it("follows the device until a choice is saved", () => {
  expect(useLocaleStore.getState().locale).toBeNull();
});

it("saves a choice", () => {
  setLocale("en");
  expect(useLocaleStore.getState().locale).toBe("en");
});

it("keeps the saved choice across a restart", async () => {
  setLocale("en");
  await new Promise((resolve) => setTimeout(resolve, 0));

  // A restart reloads the store, while the device's storage stays as it was.
  jest.resetModules();
  jest.doMock("@react-native-async-storage/async-storage", () => AsyncStorage);
  const { useLocaleStore: restarted } =
    require("@/state/locale") as typeof import("@/state/locale");
  await restarted.persist.rehydrate();

  expect(restarted.getState().locale).toBe("en");
});
