import type { ReactNode } from "react";
import { View } from "react-native";
import { KeyboardAwareScrollView as RNKeyboardAwareScrollView } from "react-native-keyboard-controller";
import { SafeAreaView as RNSafeAreaView } from "react-native-safe-area-context";
import { withUniwind } from "uniwind";

// Third-party components take className only through withUniwind.
const SafeAreaView = withUniwind(RNSafeAreaView);
const KeyboardAwareScrollView = withUniwind(RNKeyboardAwareScrollView);

// Always scrolls, so landscape, a Fold's outer display and large font scales never clip.
export function ContentColumn({ children }: { children: ReactNode }) {
  return (
    <SafeAreaView
      edges={["top", "right", "bottom", "left"]}
      className="flex-1 bg-background"
    >
      <KeyboardAwareScrollView contentContainerClassName="grow">
        <View className="w-full max-w-2xl self-center px-4 grow">
          {children}
        </View>
      </KeyboardAwareScrollView>
    </SafeAreaView>
  );
}
