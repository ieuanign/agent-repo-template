// An empty value is unset: a deployment that exported it blank configured nothing.
export function withDefault(
  value: string | undefined,
  fallback: string,
): string {
  return value !== undefined && value !== "" ? value : fallback;
}

// Read per call, never at import: one image is promoted across environments.
export function version(): string {
  return withDefault(process.env.VERSION, "dev");
}

export function gitSha(): string {
  return withDefault(process.env.GIT_SHA, "dev");
}

export function appEnv(): string {
  return withDefault(process.env.APP_ENV, "dev");
}

// Through Traefik, never backend:8080: during a rollout Compose DNS also answers with unhealthy copies.
export function apiBaseUrl(): string {
  return withDefault(process.env.API_BASE_URL, "http://traefik:4008/api");
}
