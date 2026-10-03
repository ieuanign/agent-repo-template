/**
 * @jest-environment jsdom
 */
import { render, screen } from "@testing-library/react";
import GlobalError from "@/app/global-error";

test("shows its copy in English and in Bahasa Indonesia", () => {
  jest.spyOn(console, "error").mockImplementation(() => {});
  // The page renders its own <html>, which cannot sit inside the default <div> container.
  render(<GlobalError error={new Error("boom")} reset={() => {}} />, {
    container: document,
  });

  screen.getByRole("heading", { name: "Something went wrong" });
  screen.getByText("Terjadi kesalahan");
  screen.getByText("Try again");
  screen.getByText("Coba lagi");
  expect(document.title).toBe("Something went wrong | Terjadi kesalahan");
});
