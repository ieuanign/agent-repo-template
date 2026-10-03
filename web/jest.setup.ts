// jest-environment-jsdom lacks the Fetch globals msw needs at import, so jsdom files run without it.
if (typeof Request !== "undefined") {
  // At module scope: a test file's imports evaluate before any beforeAll.
  const { server } =
    require("@/mocks/server") as typeof import("@/mocks/server");
  server.listen({ onUnhandledRequest: "error" });
  beforeEach(() => server.resetHandlers());
  afterAll(() => server.close());
}
