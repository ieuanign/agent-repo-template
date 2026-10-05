// Jest hands each test file its own process.env copy; process.loadEnvFile writes only Node's real one.
Object.assign(
  process.env,
  require("node:util").parseEnv(
    require("node:fs").readFileSync("environments/dev.env", "utf8"),
  ),
);

// Hermes lacks these, so a test fails the way a phone would; index.ts polyfills PluralRules alone.
const intl = Intl as unknown as Record<string, unknown>;
for (const name of [
  "PluralRules",
  "Locale",
  "RelativeTimeFormat",
  "ListFormat",
  "DisplayNames",
]) {
  delete intl[name];
}
delete (Intl.DateTimeFormat.prototype as Partial<Intl.DateTimeFormat>)
  .formatRange;
require("@formatjs/intl-pluralrules/polyfill-force.js");
require("@formatjs/intl-pluralrules/locale-data/id.js");
require("@formatjs/intl-pluralrules/locale-data/en.js");

jest.mock("@react-native-async-storage/async-storage", () =>
  require("@react-native-async-storage/async-storage/jest/async-storage-mock"),
);
jest.mock("react-native-keyboard-controller", () =>
  require("react-native-keyboard-controller/jest"),
);
jest.mock(
  "react-native-safe-area-context",
  () => require("react-native-safe-area-context/jest/mock").default,
);

// At module scope: a test file's imports evaluate before any beforeAll.
const { server } = require("@/mocks/server") as typeof import("@/mocks/server");
server.listen({ onUnhandledRequest: "error" });
beforeEach(() => server.resetHandlers());
afterAll(() => server.close());
