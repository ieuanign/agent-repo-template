import type { Messages } from "next-intl";
import {
  DEFAULT_LOCALE,
  LOCALES,
  type Locale,
} from "@agent-repo-template/i18n";

function isLocale(value: string): value is Locale {
  return (LOCALES as readonly string[]).includes(value);
}

// Intl.Locale canonicalises case and legacy tags (in -> id) and throws on `*` and malformed tags.
function primaryLanguage(tag: string): string | undefined {
  try {
    return new Intl.Locale(tag).language;
  } catch {
    return undefined;
  }
}

function parseAcceptLanguage(header: string): { tag: string; q: number }[] {
  const ranges: { tag: string; q: number }[] = [];
  for (const part of header.split(",")) {
    const [tag, ...params] = part.split(";").map((s) => s.trim());
    let q = 1;
    for (const param of params) {
      const [key, value] = param.split("=").map((s) => s.trim());
      if (key === "q") q = Number(value);
    }
    if (tag && Number.isFinite(q) && q > 0 && q <= 1) ranges.push({ tag, q });
  }
  // Array.prototype.sort is stable, so equal q keeps header order.
  return ranges.sort((a, b) => b.q - a.q);
}

export function resolveLocale(
  cookie: string | undefined,
  acceptLanguage: string | undefined,
): Locale {
  if (cookie && isLocale(cookie)) return cookie;
  for (const { tag } of parseAcceptLanguage(acceptLanguage ?? "")) {
    const language = primaryLanguage(tag);
    if (language && isLocale(language)) return language;
  }
  return DEFAULT_LOCALE;
}

export function pick<K extends keyof Messages>(
  messages: Messages,
  namespaces: readonly K[],
): Pick<Messages, K> {
  return Object.fromEntries(
    namespaces.map((namespace) => [namespace, messages[namespace]]),
  ) as Pick<Messages, K>;
}
