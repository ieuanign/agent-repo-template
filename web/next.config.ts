import type { NextConfig } from "next";
import createNextIntlPlugin from "next-intl/plugin";

// Relative: the default path is outside src/lib, and Turbopack rejects absolute paths.
const withNextIntl = createNextIntlPlugin("./src/lib/i18n/request.ts");

const nextConfig: NextConfig = {
  output: "standalone",
  // The e2e browser's origin is http://traefik:4008; next dev refuses /_next/* (HMR included) from
  // origins it does not list. No effect on the standalone server.
  allowedDevOrigins: ["traefik"],
  transpilePackages: [
    "@agent-repo-template/api-client",
    "@agent-repo-template/i18n",
  ],
};

export default withNextIntl(nextConfig);
