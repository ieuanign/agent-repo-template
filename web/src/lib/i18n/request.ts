import { cookies, headers } from "next/headers";
import { getRequestConfig } from "next-intl/server";
import { messages as shared, type Locale } from "@agent-repo-template/i18n";
import en from "@messages/en.json";
import id from "@messages/id.json";
import { LOCALE_COOKIE } from "@/lib/i18n/constants";
import { resolveLocale } from "@/lib/i18n/utils";

type Shared = (typeof shared)["id"];
type Own = typeof id;

declare module "next-intl" {
  interface AppConfig {
    Locale: Locale;
    Messages: Shared & Own;
  }
}

// Typed against id.json, so an en.json missing one of its keys fails tsc. The never-typed shared
// names make a web namespace named like a shared one fail tsc instead of silently replacing it.
const own: Record<Locale, Own & Partial<Record<keyof Shared, never>>> = {
  id,
  en,
};

const catalogues: Record<Locale, Shared & Own> = {
  id: { ...shared.id, ...own.id },
  en: { ...shared.en, ...own.en },
};

export default getRequestConfig(async () => {
  const [cookieStore, headerStore] = await Promise.all([cookies(), headers()]);
  const locale = resolveLocale(
    cookieStore.get(LOCALE_COOKIE)?.value,
    headerStore.get("accept-language") ?? undefined,
  );
  return { locale, messages: catalogues[locale] };
});
