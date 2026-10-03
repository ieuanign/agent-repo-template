import { expect, test } from "@playwright/test";

test("the HMR websocket connects through the edge and receives a frame", async ({
  page,
  baseURL,
}) => {
  // Listen before navigating, and for frames as soon as the socket opens: goto resolves only after load.
  const socket = page
    .waitForEvent("websocket")
    .then((ws) => ({ ws, frame: ws.waitForEvent("framereceived") }));
  await page.goto("/");
  const { ws, frame } = await socket;

  const url = new URL(ws.url());
  expect(url.host).toBe(new URL(baseURL!).host);
  expect(url.pathname).toBe("/_next/hmr");
  // Rejects with "Socket error" when next dev refuses the browser's Origin.
  await frame;
});
