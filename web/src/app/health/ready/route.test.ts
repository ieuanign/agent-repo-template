import { GET } from "@/app/health/ready/route";

// The real connection() throws outside a request scope.
jest.mock("next/server", () => ({ connection: () => Promise.resolve() }));

describe("GET /health/ready", () => {
  it("answers 200 with an empty JSON body", async () => {
    const res = await GET();
    expect(res.status).toBe(200);
    expect(res.headers.get("content-type")).toMatch(/^application\/json/);
    expect(await res.json()).toEqual({});
  });
});
