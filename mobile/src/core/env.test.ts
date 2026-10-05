describe("apiUrl", () => {
  const saved = process.env.EXPO_PUBLIC_API_URL;
  afterEach(() => {
    process.env.EXPO_PUBLIC_API_URL = saved;
  });

  it("reads EXPO_PUBLIC_API_URL", () => {
    jest.isolateModules(() => {
      expect(require("@/core/env").apiUrl).toBe("http://localhost:4008/api");
    });
  });

  it("throws naming the key when it is empty", () => {
    process.env.EXPO_PUBLIC_API_URL = "";
    expect(() => jest.isolateModules(() => require("@/core/env"))).toThrow(
      "EXPO_PUBLIC_API_URL",
    );
  });
});
