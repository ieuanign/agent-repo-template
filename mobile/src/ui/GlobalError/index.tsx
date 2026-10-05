import { ScrollView, View } from "react-native";
import { SafeAreaView as RNSafeAreaView } from "react-native-safe-area-context";
import { withUniwind } from "uniwind";

import { Button } from "@/components/Button";
import { Text } from "@/components/Text";
import { COPY } from "@/ui/GlobalError/constants";

const SafeAreaView = withUniwind(RNSafeAreaView);

// Renders outside every provider, so no translator and no ContentColumn (it needs KeyboardProvider).
export default function GlobalError({ retry }: { retry: () => void }) {
  return (
    <SafeAreaView
      edges={["top", "right", "bottom", "left"]}
      className="flex-1 bg-background"
    >
      <ScrollView contentContainerClassName="grow">
        <View className="w-full max-w-2xl self-center px-4 grow">
          <Text variant="h1">{COPY.en.heading}</Text>
          <Text>{COPY.id.heading}</Text>
          {/* h-auto overrides the size's fixed height so both stacked lines fit. */}
          <Button onPress={retry} className="mt-4 h-auto flex-col py-1">
            <Text>{COPY.en.retry}</Text>
            <Text>{COPY.id.retry}</Text>
          </Button>
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}
