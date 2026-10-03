import { resolveLocale } from "@/lib/i18n/utils";

describe("resolveLocale", () => {
  it("defaults to id with no cookie and no header", () => {
    expect(resolveLocale(undefined, undefined)).toBe("id");
  });

  it("prefers a valid cookie over the header", () => {
    expect(resolveLocale("en", "id")).toBe("en");
    expect(resolveLocale("id", "en")).toBe("id");
  });

  it.each(["fr", "EN", ""])("ignores the cookie value %p", (cookie) => {
    expect(resolveLocale(cookie, "en")).toBe("en");
  });

  it.each([
    ["en-US,en;q=0.9", "en"],
    ["id-ID,id;q=0.9,en-US;q=0.8,en;q=0.7", "id"],
    ["id;q=0.5,en;q=0.8", "en"],
    ["en;q=0", "id"],
    ["fr-FR,de", "id"],
    ["in-ID", "id"],
  ])("matches Accept-Language %p to %p", (header, expected) => {
    expect(resolveLocale(undefined, header)).toBe(expected);
  });

  it.each(["*", "!!,;q=x,@@", ";;;", "en;q=abc"])(
    "falls back to id on %p without throwing",
    (header) => {
      expect(resolveLocale(undefined, header)).toBe("id");
    },
  );
});
