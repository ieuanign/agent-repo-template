import type { Envelope } from "@agent-repo-template/api-client";
import { http, HttpResponse } from "msw";

import { API_URLS } from "@/core/apiUrls";
import { reportBackend } from "@/core/backendReport";
import { apiBaseUrl } from "@/core/env";
import { server } from "@/mocks/server";

const liveness = () => `${apiBaseUrl()}${API_URLS.liveness}`;

describe("reportBackend", () => {
  let log: jest.SpyInstance;
  let warn: jest.SpyInstance;
  beforeEach(() => {
    log = jest.spyOn(console, "log").mockImplementation(() => {});
    warn = jest.spyOn(console, "error").mockImplementation(() => {});
  });
  afterEach(() => {
    jest.useRealTimers();
    jest.restoreAllMocks();
  });

  it("logs backend's version and environment once when it answers", async () => {
    await reportBackend();

    expect(log).toHaveBeenCalledTimes(1);
    expect(JSON.parse(log.mock.calls[0][0])).toEqual({
      msg: "backend reachable",
      service: "backend",
      version: "dev",
      environment: "dev",
    });
    expect(warn).not.toHaveBeenCalled();
  });

  it("warns once when backend answers with a failure envelope", async () => {
    jest.useFakeTimers({ doNotFake: ["nextTick", "queueMicrotask"] });
    server.use(
      http.get(liveness(), () =>
        HttpResponse.json<Envelope<null>>(
          {
            meta: { path: "/api/health", timestamp: "2026-09-30T00:00:00Z" },
            status: 503,
            code: "INTERNAL",
            message: "Something went wrong.",
            payload: null,
          },
          { status: 503 },
        ),
      ),
    );

    const report = reportBackend();
    await jest.advanceTimersByTimeAsync(400_000);
    await report;

    expect(warn).toHaveBeenCalledTimes(1);
    expect(log).not.toHaveBeenCalled();
  });

  it("warns once when backend never answers within 10 seconds", async () => {
    jest.useFakeTimers({ doNotFake: ["nextTick", "queueMicrotask"] });
    server.use(http.get(liveness(), () => new Promise<never>(() => {})));

    const report = reportBackend();
    await jest.advanceTimersByTimeAsync(500_000);
    await report;

    expect(warn).toHaveBeenCalledTimes(1);
    expect(log).not.toHaveBeenCalled();
  });

  it("retries with doubling waits until backend answers", async () => {
    jest.useFakeTimers({ doNotFake: ["nextTick", "queueMicrotask"], now: 0 });
    const failures = [
      () => HttpResponse.error(),
      () => HttpResponse.json({ status: 503 }, { status: 503 }),
      () => HttpResponse.text("404 page not found", { status: 404 }),
    ];
    const seen: number[] = [];
    server.use(
      http.get(liveness(), () => {
        seen.push(Date.now());
        // Returning nothing falls through to the default success handler.
        return failures[seen.length - 1]?.();
      }),
    );

    const report = reportBackend();
    for (const [i, at] of [2_000, 6_000, 14_000].entries()) {
      await jest.advanceTimersByTimeAsync(at - 1 - Date.now());
      expect(seen).toHaveLength(i + 1);
      await jest.advanceTimersByTimeAsync(1);
      expect(seen).toHaveLength(i + 2);
    }
    await report;
    await jest.advanceTimersByTimeAsync(600_000);

    expect(seen).toEqual([0, 2_000, 6_000, 14_000]);
    expect(log).toHaveBeenCalledTimes(1);
    expect(JSON.parse(log.mock.calls[0][0])).toMatchObject({
      version: "dev",
      environment: "dev",
    });
    expect(warn).not.toHaveBeenCalled();
  });

  it("tries 11 times, then warns once, when every try fails", async () => {
    jest.useFakeTimers({ doNotFake: ["nextTick", "queueMicrotask"], now: 0 });
    const seen: number[] = [];
    server.use(
      http.get(liveness(), () => {
        seen.push(Date.now());
        return HttpResponse.error();
      }),
    );

    const report = reportBackend();
    await jest.advanceTimersByTimeAsync(362_000);
    await report;
    await jest.advanceTimersByTimeAsync(600_000);

    expect(seen).toEqual([
      0, 2_000, 6_000, 14_000, 30_000, 62_000, 122_000, 182_000, 242_000,
      302_000, 362_000,
    ]);
    expect(warn).toHaveBeenCalledTimes(1);
    expect(log).not.toHaveBeenCalled();
  });

  it("abandons a hung try after 10 seconds, then waits before retrying", async () => {
    jest.useFakeTimers({ doNotFake: ["nextTick", "queueMicrotask"], now: 0 });
    const seen: number[] = [];
    server.use(
      http.get(liveness(), () => {
        seen.push(Date.now());
        return seen.length === 1 ? new Promise<never>(() => {}) : undefined;
      }),
    );

    const report = reportBackend();
    await jest.advanceTimersByTimeAsync(11_999);
    expect(seen).toEqual([0]);
    await jest.advanceTimersByTimeAsync(1);
    await report;

    expect(seen).toEqual([0, 12_000]);
    expect(log).toHaveBeenCalledTimes(1);
    expect(warn).not.toHaveBeenCalled();
  });
});
