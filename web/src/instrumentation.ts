import { reportBackend } from "@/core/backendReport";
import { appEnv, gitSha, version } from "@/core/env";

// The only place the commit is disclosed.
export function register() {
  console.log(
    JSON.stringify({
      msg: "started",
      service: "web",
      version: version(),
      commit: gitSha(),
      environment: appEnv(),
    }),
  );
  // Not awaited: boot and readiness never wait on backend.
  if (process.env.NEXT_RUNTIME === "nodejs") void reportBackend();
}
