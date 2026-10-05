import { render, screen, userEvent } from "@testing-library/react-native";

import { Button } from "@/components/Button";
import { Text } from "@/components/Text";

async function renderButton(props: { disabled?: boolean } = {}) {
  const onPress = jest.fn();
  await render(
    <Button onPress={onPress} {...props}>
      <Text>Coba lagi</Text>
    </Button>,
  );
  return onPress;
}

test("renders its label", async () => {
  await renderButton();

  expect(screen.getByRole("button", { name: "Coba lagi" })).toBeOnTheScreen();
});

test("calls onPress when pressed", async () => {
  const user = userEvent.setup();
  const onPress = await renderButton();

  await user.press(screen.getByRole("button", { name: "Coba lagi" }));

  expect(onPress).toHaveBeenCalledTimes(1);
});

test("takes no press while disabled", async () => {
  const user = userEvent.setup();
  const onPress = await renderButton({ disabled: true });

  await user.press(screen.getByRole("button", { name: "Coba lagi" }));

  expect(onPress).not.toHaveBeenCalled();
});
