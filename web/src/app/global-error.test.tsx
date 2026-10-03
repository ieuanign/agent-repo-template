/**
 * @jest-environment jsdom
 */
import { render, screen } from "@testing-library/react";
import GlobalError from "@/app/global-error";

test("shows its copy in Bahasa Indonesia and in English", () => {
  jest.spyOn(console, "error").mockImplementation(() => {});
  // The page renders its own <html>, which cannot sit inside the default <div> container.
  render(<GlobalError error={new Error("boom")} reset={() => {}} />, {
    container: document,
  });

  screen.getByRole("heading", { name: "Terjadi kesalahan" });
  screen.getByText("Something went wrong");
  screen.getByText("Coba lagi");
  screen.getByText("Try again");
  expect(document.title).toBe("Terjadi kesalahan | Something went wrong");
});
