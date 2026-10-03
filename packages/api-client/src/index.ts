import createClient from "openapi-fetch";
import type { components, paths as generated } from "./schema";

export type { components, operations } from "./schema";

export type ErrorCode = components["schemas"]["ErrorCode"];

/** backend's wire envelope: `code` is null on success, `payload` null on failure. */
// openapi-fetch's Readable drops null-only properties, so `payload: null` is absent from reads.
export type Envelope<T> = {
  meta: { path: string; timestamp: string };
  status: number;
  code: ErrorCode | null;
  message: string;
  payload: T;
};

type Json<T> = {
  headers: Record<string, unknown>;
  content: { "application/json": T };
};

type Failure = Json<Omit<Envelope<null>, "code"> & { code: ErrorCode }>;

// The spec describes only `payload`; 204 and 304 carry no body, so no envelope.
type Responses<R> = {
  [S in keyof R]: S extends 204 | 304 | "204" | "304"
    ? R[S]
    : R[S] extends { content: { "application/json": infer B } }
      ? Json<Envelope<B>>
      : Json<Envelope<null>>;
} & { default: Failure };

type Retyped<Op> = Op extends { responses: infer R }
  ? Omit<Op, "responses"> & { responses: Responses<R> }
  : Op;

/** The generated paths with every response retyped as backend's envelope. */
export type paths = {
  [P in keyof generated]: {
    [M in keyof generated[P]]: Retyped<generated[P][M]>;
  };
};

/** Typed client over backend; responses reach the caller unmodified. */
export function createApiClient(baseUrl: string) {
  return createClient<paths>({ baseUrl });
}
