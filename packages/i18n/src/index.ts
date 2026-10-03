import en from "../messages/en.json";
import id from "../messages/id.json";

export const LOCALES = ["en", "id"] as const;

export type Locale = (typeof LOCALES)[number];

export const DEFAULT_LOCALE: Locale = "en";

// Typed against id.json, so an en.json missing one of its keys fails tsc.
export const messages: Record<Locale, typeof id> = { id, en };
