import { useTranslations } from "use-intl";

import { Button } from "@/components/Button";
import { ContentColumn } from "@/components/ContentColumn";
import { Text } from "@/components/Text";
import type { ErrorScreenProps } from "@/ui/Error/types";

export default function ErrorScreen({ retry }: ErrorScreenProps) {
  const t = useTranslations("error");

  return (
    <ContentColumn>
      <Text variant="h1">{t("heading")}</Text>
      <Button onPress={retry} className="mt-4">
        <Text>{t("retry")}</Text>
      </Button>
    </ContentColumn>
  );
}
