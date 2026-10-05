import { render, screen, userEvent } from "@testing-library/react-native";
import { IntlProvider } from "use-intl";

import { messages } from "@/lib/i18n/messages";
import ErrorScreen from "@/ui/Error";

test("Retry calls retry once", async () => {
  const user = userEvent.setup();
  const retry = jest.fn();
  await render(
    <IntlProvider locale="id" messages={messages.id}>
      <ErrorScreen retry={retry} />
    </IntlProvider>,
  );

  expect(screen.getByText("Terjadi kesalahan")).toBeOnTheScreen();
  await user.press(screen.getByRole("button", { name: "Coba lagi" }));

  expect(retry).toHaveBeenCalledTimes(1);
});
