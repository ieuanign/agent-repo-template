import type { textVariants } from "@/components/Text/styles";
import type { VariantProps } from "class-variance-authority";
import type { Text as RNText } from "react-native";

type TextVariantProps = VariantProps<typeof textVariants>;

type TextVariant = NonNullable<TextVariantProps["variant"]>;

type TextProps = React.ComponentProps<typeof RNText> &
  React.RefAttributes<typeof RNText> &
  TextVariantProps & {
    asChild?: boolean;
  };

export type { TextProps, TextVariant };
