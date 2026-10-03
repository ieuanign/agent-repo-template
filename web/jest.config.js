import nextJest from "next/jest.js";

// `next` publishes no exports map, so the specifier above needs its extension.
const createJestConfig = nextJest({ dir: "./" });

const nextJestConfig = createJestConfig({
  testEnvironment: "node",
  setupFilesAfterEnv: ["<rootDir>/jest.setup.ts"],
  // SWC rewrites the tsconfig alias only in import statements; the setup file's require needs this.
  moduleNameMapper: { "^@/(.*)$": "<rootDir>/src/$1" },
});

// next-intl, msw and the packages they load ship only as ES modules, which next/jest never transforms.
export default async function jestConfig() {
  const config = await nextJestConfig();
  return {
    ...config,
    transformIgnorePatterns: [
      "/node_modules/(?!(next-intl|use-intl|intl-messageformat|@formatjs|msw|@mswjs|@open-draft|until-async|outvariant|strict-event-emitter|headers-polyfill|is-node-process|rettime)/)",
      "^.+\\.module\\.(css|sass|scss)$",
    ],
  };
}
