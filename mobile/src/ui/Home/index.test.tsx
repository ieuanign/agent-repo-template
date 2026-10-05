import type { Envelope } from "@agent-repo-template/api-client";
import { act, render, screen, userEvent } from "@testing-library/react-native";
import { delay, http, HttpResponse } from "msw";

import { API_URLS } from "@/core/apiUrls";
import { apiUrl } from "@/core/env";
import { server } from "@/mocks/server";
import { queryWrapper } from "@/test/queryWrapper";
import Home from "@/ui/Home";

test("shows the shared heading and tagline", async () => {
  await render(<Home />, { wrapper: queryWrapper() });

  expect(screen.getByText("Agent Repo Template")).toBeOnTheScreen();
  expect(
    screen.getByText(
      "Monorepo full-stack yang siap dikerjakan agen sejak klon pertama.",
    ),
  ).toBeOnTheScreen();
  // Settles the health query inside act, so its update does not land after the test.
  await screen.findByText("Versi backend: dev");
});

test("shows backend's version", async () => {
  await render(<Home />, { wrapper: queryWrapper() });

  expect(await screen.findByText("Versi backend: dev")).toBeOnTheScreen();
});

test("says it is reaching backend while the request is in flight", async () => {
  await render(<Home />, { wrapper: queryWrapper() });

  expect(screen.getByText("Menghubungkan ke backend…")).toBeOnTheScreen();
  expect(await screen.findByText("Versi backend: dev")).toBeOnTheScreen();
});

test("offers Retry when backend answers 503, then shows the version", async () => {
  const user = userEvent.setup();
  server.use(
    http.get(
      `${apiUrl}${API_URLS.liveness}`,
      ({ request }) =>
        HttpResponse.json<Envelope<null>>(
          {
            meta: {
              path: new URL(request.url).pathname,
              timestamp: new Date().toISOString(),
            },
            status: 503,
            code: "INTERNAL",
            message: "Something went wrong.",
            payload: null,
          },
          { status: 503 },
        ),
      { once: true },
    ),
  );
  await render(<Home />, { wrapper: queryWrapper() });

  expect(
    await screen.findByText("Backend tidak dapat dihubungi."),
  ).toBeOnTheScreen();
  await user.press(screen.getByRole("button", { name: "Coba lagi" }));

  expect(await screen.findByText("Versi backend: dev")).toBeOnTheScreen();
});

test("shows backend unreachable when a request hangs past 10 s", async () => {
  jest.useFakeTimers();
  try {
    server.use(
      http.get(`${apiUrl}${API_URLS.liveness}`, async () => {
        await delay("infinite");
      }),
    );
    await render(<Home />, { wrapper: queryWrapper() });

    await act(() => jest.advanceTimersByTimeAsync(10_000));

    expect(
      await screen.findByText("Backend tidak dapat dihubungi."),
    ).toBeOnTheScreen();
  } finally {
    jest.useRealTimers();
  }
});
