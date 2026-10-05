import { useTranslations } from "use-intl";

import { ContentColumn } from "@/components/ContentColumn";
import { Text } from "@/components/Text";

export default function NotFound() {
  const t = useTranslations("notFound");

  return (
    <ContentColumn>
      <Text variant="h1">{t("heading")}</Text>
    </ContentColumn>
  );
}
