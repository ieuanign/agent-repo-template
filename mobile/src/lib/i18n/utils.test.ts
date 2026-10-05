import { resolveLocale } from "@/lib/i18n/utils";

const device = (...codes: (string | null)[]) =>
  codes.map((languageCode) => ({ languageCode }));

describe("resolveLocale", () => {
  it("uses the saved choice over the device", () => {
    expect(resolveLocale("en", device("id"))).toBe("en");
    expect(resolveLocale("id", device("en"))).toBe("id");
  });

  it.each([
    [device("en", "id"), "en"],
    [device("fr", "id", "en"), "id"],
    [device(null, "en"), "en"],
    [device("in"), "id"],
  ])("follows the device list %j to %p", (locales, expected) => {
    expect(resolveLocale(null, locales)).toBe(expected);
  });

  it.each([[device("fr", "de")], [device()]])(
    "falls back to en for %j",
    (locales) => {
      expect(resolveLocale(null, locales)).toBe("en");
    },
  );
});
