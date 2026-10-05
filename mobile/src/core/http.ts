import { createApiClient } from "@agent-repo-template/api-client";

import { apiUrl } from "@/core/env";

// Module scope: Metro fixes EXPO_PUBLIC_* into the bundle, so nothing differs between calls.
export const http = createApiClient(apiUrl);
