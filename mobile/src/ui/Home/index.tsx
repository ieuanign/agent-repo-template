import { useQuery } from "@tanstack/react-query";
import { View } from "react-native";
import { useTranslations } from "use-intl";

import { Button } from "@/components/Button";
import { ContentColumn } from "@/components/ContentColumn";
import { Text } from "@/components/Text";
import { API_URLS } from "@/core/apiUrls";
import { http } from "@/core/http";
import { TIMEOUT_MS } from "@/ui/Home/constants";

export default function Home() {
  const t = useTranslations("home");
  const tBackend = useTranslations("backend");
  const tError = useTranslations("error");
  const health = useQuery({
    queryKey: ["health"],
    queryFn: async () => {
      // AbortSignal.timeout is absent from React Native's polyfill, and fake timers cannot advance it.
      const controller = new AbortController();
      const timer = setTimeout(() => controller.abort(), TIMEOUT_MS);
      try {
        // Only data's presence means success: Traefik answers plain text until backend is routed.
        const { data } = await http.GET(API_URLS.liveness, {
          signal: controller.signal,
        });
        if (!data) throw new Error("Backend did not answer.");
        return data.payload;
      } finally {
        clearTimeout(timer);
      }
    },
  });

  return (
    <ContentColumn>
      <Text variant="h1">{t("heading")}</Text>
      <Text variant="lead">{t("tagline")}</Text>
      {health.data ? (
        <Text className="mt-4">
          {tBackend("version", { version: health.data.version })}
        </Text>
      ) : health.isError ? (
        <View className="mt-4 flex-row items-center gap-2">
          <Text>{tBackend("unreachable")}</Text>
          <Button onPress={() => health.refetch()}>
            <Text>{tError("retry")}</Text>
          </Button>
        </View>
      ) : (
        <Text className="mt-4">{tBackend("loading")}</Text>
      )}
    </ContentColumn>
  );
}
