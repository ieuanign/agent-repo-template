import { screen } from "@testing-library/react-native";
import { renderRouter } from "expo-router/testing-library";

jest.mock("expo-localization", () => ({
  useLocales: () => [{ languageCode: "id" }],
  useCalendars: () => [{ timeZone: "Asia/Jakarta" }],
}));

test("an unmatched URL shows the not-found screen", async () => {
  await renderRouter("./src/app", { initialUrl: "/no-such-screen" });

  expect(await screen.findByText("Layar tidak ditemukan")).toBeOnTheScreen();
});
