import {
  DEFAULT_LOCALE,
  LOCALES,
  type Locale,
} from "@agent-repo-template/i18n";

function isLocale(value: string): value is Locale {
  return (LOCALES as readonly string[]).includes(value);
}

// Hermes has no Intl.Locale, so Android's legacy `in` for Indonesian is mapped by hand.
export function resolveLocale(
  saved: Locale | null,
  device: readonly { languageCode: string | null }[],
): Locale {
  if (saved) return saved;
  for (const { languageCode } of device) {
    const language = languageCode === "in" ? "id" : languageCode;
    if (language && isLocale(language)) return language;
  }
  return DEFAULT_LOCALE;
}
