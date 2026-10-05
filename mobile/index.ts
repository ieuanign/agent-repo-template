// Hermes has no Intl.PluralRules; use-intl needs it before any screen renders.
import "@formatjs/intl-pluralrules/polyfill-force.js";
import "@formatjs/intl-pluralrules/locale-data/id.js";
import "@formatjs/intl-pluralrules/locale-data/en.js";
import "expo-router/entry";
