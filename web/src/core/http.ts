import { createApiClient } from "@agent-repo-template/api-client";

import { apiBaseUrl } from "@/core/env";

// Built per call so API_BASE_URL is read at call time, never cached.
export function http() {
  return createApiClient(apiBaseUrl());
}
