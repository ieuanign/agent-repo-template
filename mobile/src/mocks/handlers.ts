import type { components, Envelope } from "@agent-repo-template/api-client";
import { http, HttpResponse } from "msw";

import { API_URLS } from "@/core/apiUrls";
import { apiUrl } from "@/core/env";

type Health = components["schemas"]["Health"];

// Full envelopes, as backend's envelope middleware writes them.
export const handlers = [
  http.get(`${apiUrl}${API_URLS.liveness}`, ({ request }) =>
    HttpResponse.json<Envelope<Health>>({
      meta: {
        path: new URL(request.url).pathname,
        timestamp: new Date().toISOString(),
      },
      status: 200,
      code: null,
      message: "Backend is up.",
      payload: { version: "dev", environment: "dev" },
    }),
  ),
];
