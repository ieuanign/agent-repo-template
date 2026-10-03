import { API_URLS } from "@/core/apiUrls";
import { http } from "@/core/http";
import { reportError } from "@/core/reporting";

// Global setTimeout, not AbortSignal.timeout: Jest's fake timers cannot advance the latter.
const TIMEOUT_MS = 10_000;
// Waits between tries: 11 tries over about six minutes, since backend starts after migrate and seed.
const WAITS_MS = [2, 4, 8, 16, 32, 60, 60, 60, 60, 60].map((s) => s * 1000);

// Never rejects, so an unawaited call cannot stop the process.
export async function reportBackend(): Promise<void> {
  let cause: unknown;
  try {
    for (let attempt = 0; attempt <= WAITS_MS.length; attempt++) {
      if (attempt > 0) {
        await new Promise((resolve) =>
          setTimeout(resolve, WAITS_MS[attempt - 1]),
        );
      }
      const controller = new AbortController();
      const timer = setTimeout(() => controller.abort(), TIMEOUT_MS);
      try {
        // Only data's presence decides success: Traefik answers plain text until backend is routed.
        const { data, error } = await http().GET(API_URLS.liveness, {
          signal: controller.signal,
        });
        if (data) {
          const { version, environment } = data.payload;
          console.log(
            JSON.stringify({
              msg: "backend reachable",
              service: "backend",
              version,
              environment,
            }),
          );
          return;
        }
        cause = error;
      } catch (err) {
        cause = err;
      } finally {
        clearTimeout(timer);
      }
    }
  } catch (err) {
    cause = err;
  }
  reportError(new Error("backend unreachable", { cause }));
}
