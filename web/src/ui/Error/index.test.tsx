/**
 * @jest-environment jsdom
 */
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { NextIntlClientProvider } from "next-intl";
import { messages } from "@agent-repo-template/i18n";
import { ErrorView } from ".";

test("the retry button calls reset once", async () => {
  const reset = jest.fn();
  render(
    <NextIntlClientProvider locale="id" messages={messages.id}>
      <ErrorView reset={reset} />
    </NextIntlClientProvider>,
  );

  await userEvent.click(screen.getByRole("button", { name: "Coba lagi" }));

  expect(reset).toHaveBeenCalledTimes(1);
});
