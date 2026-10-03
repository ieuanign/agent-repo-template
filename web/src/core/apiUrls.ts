import type { paths } from "@agent-repo-template/api-client";

export const API_URLS = {
  liveness: "/health",
} as const satisfies Record<string, keyof paths>;
