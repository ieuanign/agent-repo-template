import { resolveLocale } from "@/lib/i18n/utils";

describe("resolveLocale", () => {
  it("defaults to en with no cookie and no header", () => {
    expect(resolveLocale(undefined, undefined)).toBe("en");
  });

  it("prefers a valid cookie over the header", () => {
    expect(resolveLocale("en", "id")).toBe("en");
    expect(resolveLocale("id", "en")).toBe("id");
  });

  it.each(["fr", "ID", ""])("ignores the cookie value %p", (cookie) => {
    expect(resolveLocale(cookie, "id")).toBe("id");
  });

  it.each([
    ["en-US,en;q=0.9", "en"],
    ["id-ID,id;q=0.9,en-US;q=0.8,en;q=0.7", "id"],
    ["id;q=0.5,en;q=0.8", "en"],
    ["id;q=0", "en"],
    ["fr-FR,de", "en"],
    ["in-ID", "id"],
  ])("matches Accept-Language %p to %p", (header, expected) => {
    expect(resolveLocale(undefined, header)).toBe(expected);
  });

  it.each(["*", "!!,;q=x,@@", ";;;", "en;q=abc"])(
    "falls back to en on %p without throwing",
    (header) => {
      expect(resolveLocale(undefined, header)).toBe("en");
    },
  );
});
