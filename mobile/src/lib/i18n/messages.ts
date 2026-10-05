import { messages as shared, type Locale } from "@agent-repo-template/i18n";
import en from "@messages/en.json";
import id from "@messages/id.json";

type Shared = (typeof shared)["id"];
type Own = typeof id;

declare module "use-intl" {
  interface AppConfig {
    Locale: Locale;
    Messages: Shared & Own;
  }
}

// Typed against id.json, so an en.json missing one of its keys fails tsc. The never-typed shared
// names make a mobile namespace named like a shared one fail tsc instead of silently replacing it.
const own: Record<Locale, Own & Partial<Record<keyof Shared, never>>> = {
  id,
  en,
};

export const messages: Record<Locale, Shared & Own> = {
  id: { ...shared.id, ...own.id },
  en: { ...shared.en, ...own.en },
};
