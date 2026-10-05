import type { buttonVariants } from "@/components/Button/styles";
import type { VariantProps } from "class-variance-authority";

type ButtonProps = React.ComponentProps<"button"> &
  VariantProps<typeof buttonVariants> & {
    asChild?: boolean;
  };

export type { ButtonProps };
