import { apiBaseUrl, appEnv, gitSha, version, withDefault } from "@/core/env";

describe("withDefault", () => {
  it("returns the value when set", () => {
    expect(withDefault("1.2.0", "dev")).toBe("1.2.0");
  });

  it("falls back when unset", () => {
    expect(withDefault(undefined, "dev")).toBe("dev");
  });

  it("falls back when empty", () => {
    expect(withDefault("", "dev")).toBe("dev");
  });
});

describe("settings", () => {
  const saved = { ...process.env };
  afterEach(() => {
    process.env = { ...saved };
  });

  it("reads each value when asked, not at import", () => {
    process.env.VERSION = "1.2.0";
    process.env.GIT_SHA = "abc1234";
    process.env.APP_ENV = "staging";
    process.env.API_BASE_URL = "http://traefik:4100/api";
    expect(version()).toBe("1.2.0");
    expect(gitSha()).toBe("abc1234");
    expect(appEnv()).toBe("staging");
    expect(apiBaseUrl()).toBe("http://traefik:4100/api");
  });

  it("defaults VERSION and GIT_SHA to dev when unset", () => {
    delete process.env.VERSION;
    delete process.env.GIT_SHA;
    expect(version()).toBe("dev");
    expect(gitSha()).toBe("dev");
  });

  it("defaults VERSION and GIT_SHA to dev when empty", () => {
    process.env.VERSION = "";
    process.env.GIT_SHA = "";
    expect(version()).toBe("dev");
    expect(gitSha()).toBe("dev");
  });

  it("defaults API_BASE_URL to Traefik when unset", () => {
    delete process.env.API_BASE_URL;
    expect(apiBaseUrl()).toBe("http://traefik:4008/api");
  });

  it("defaults API_BASE_URL to Traefik when empty", () => {
    process.env.API_BASE_URL = "";
    expect(apiBaseUrl()).toBe("http://traefik:4008/api");
  });
});
