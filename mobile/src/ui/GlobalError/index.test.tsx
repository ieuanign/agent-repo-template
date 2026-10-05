import { render, screen, userEvent } from "@testing-library/react-native";

import GlobalError from "@/ui/GlobalError";

test("shows the heading and retry in both languages", async () => {
  const user = userEvent.setup();
  const retry = jest.fn();
  await render(<GlobalError retry={retry} />);

  expect(screen.getByText("Something went wrong")).toBeOnTheScreen();
  expect(screen.getByText("Terjadi kesalahan")).toBeOnTheScreen();
  await user.press(
    screen.getByRole("button", { name: /Try again.*Coba lagi/ }),
  );

  expect(retry).toHaveBeenCalledTimes(1);
});
