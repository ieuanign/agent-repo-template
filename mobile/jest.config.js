const preset = require("jest-expo/jest-preset");

// MSW, Uniwind and use-intl ship ES modules, which the preset's allowlist leaves untransformed.
const esModules =
  "uniwind|@rn-primitives|use-intl|icu-minify|intl-messageformat|@formatjs|msw|rettime|@open-draft|@mswjs|until-async|outvariant|strict-event-emitter|headers-polyfill|is-node-process";
const [allowlist, ...ignored] = preset.transformIgnorePatterns;

module.exports = {
  preset: "jest-expo",
  setupFilesAfterEnv: ["<rootDir>/jest.setup.ts"],
  // The .css stub comes first: the root layout's @/../global.css also matches the @/ alias.
  moduleNameMapper: {
    "\\.css$": "<rootDir>/src/test/cssStub.js",
    "^@/(.*)$": "<rootDir>/src/$1",
    "^@messages/(.*)$": "<rootDir>/messages/$1",
  },
  transform: { "\\.mjs$": preset.transform["\\.[jt]sx?$"] },
  transformIgnorePatterns: [
    allowlist.replace(/\)\)$/, `|${esModules}))`),
    ...ignored,
  ],
};
