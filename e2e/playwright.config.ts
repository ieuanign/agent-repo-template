import { defineConfig, devices } from "@playwright/test";

export default defineConfig({
  testDir: "./tests",
  use: { baseURL: process.env.BASE_URL },
  projects: [{ name: "chromium", use: { ...devices["Desktop Chrome"] } }],
});
