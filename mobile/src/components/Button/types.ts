import type { buttonVariants } from "@/components/Button/styles";
import type { VariantProps } from "class-variance-authority";
import type { Pressable } from "react-native";

type ButtonProps = React.ComponentProps<typeof Pressable> &
  React.RefAttributes<typeof Pressable> &
  VariantProps<typeof buttonVariants>;

export type { ButtonProps };
