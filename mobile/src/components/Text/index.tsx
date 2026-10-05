import { ARIA_LEVEL, ROLE } from "@/components/Text/constants";
import { textVariants } from "@/components/Text/styles";
import type { TextProps } from "@/components/Text/types";
import { cn } from "@/lib/utils";
import { Slot } from "@rn-primitives/slot";
import * as React from "react";
import { Text as RNText } from "react-native";

const TextClassContext = React.createContext<string | undefined>(undefined);

function Text({
  className,
  asChild = false,
  variant = "default",
  ...props
}: TextProps) {
  const textClass = React.useContext(TextClassContext);
  const Component = asChild ? Slot : RNText;
  return (
    <Component
      className={cn(textVariants({ variant }), textClass, className)}
      role={variant ? ROLE[variant] : undefined}
      aria-level={variant ? ARIA_LEVEL[variant] : undefined}
      {...props}
    />
  );
}

export { Text, TextClassContext };
