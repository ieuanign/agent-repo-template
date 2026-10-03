import { GET } from "@/app/health/route";

// The real connection() throws outside a request scope.
jest.mock("next/server", () => ({ connection: () => Promise.resolve() }));

describe("GET /health", () => {
  const saved = { ...process.env };
  beforeEach(() => {
    process.env.VERSION = "1.2.0";
    process.env.APP_ENV = "staging";
    process.env.GIT_SHA = "abc1234";
  });
  afterEach(() => {
    process.env = { ...saved };
  });

  it("answers 200 with only version and environment as JSON", async () => {
    const res = await GET();
    expect(res.status).toBe(200);
    expect(res.headers.get("content-type")).toMatch(/^application\/json/);
    expect(await res.json()).toEqual({
      version: "1.2.0",
      environment: "staging",
    });
  });

  it("reads the environment on every call", async () => {
    await GET();
    process.env.VERSION = "1.3.0";
    process.env.APP_ENV = "production";
    expect(await (await GET()).json()).toEqual({
      version: "1.3.0",
      environment: "production",
    });
  });

  it("reports dev when VERSION is unset", async () => {
    delete process.env.VERSION;
    expect((await (await GET()).json()).version).toBe("dev");
  });

  it("reports dev when VERSION is empty", async () => {
    process.env.VERSION = "";
    expect((await (await GET()).json()).version).toBe("dev");
  });
});
