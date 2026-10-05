import { buttonTextVariants, buttonVariants } from "@/components/Button/styles";
import type { ButtonProps } from "@/components/Button/types";
import { TextClassContext } from "@/components/Text";
import { cn } from "@/lib/utils";
import { Pressable } from "react-native";

function Button({ className, variant, size, ...props }: ButtonProps) {
  return (
    <TextClassContext.Provider value={buttonTextVariants({ variant, size })}>
      <Pressable
        className={cn(
          props.disabled && "opacity-50",
          buttonVariants({ variant, size }),
          className,
        )}
        role="button"
        {...props}
      />
    </TextClassContext.Provider>
  );
}

export { Button };
export type { ButtonProps };
