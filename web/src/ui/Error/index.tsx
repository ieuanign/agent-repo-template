"use client";

import { useTranslations } from "next-intl";
import { Button } from "@/components/Button";

export function ErrorView({ reset }: { reset: () => void }) {
  const t = useTranslations("error");
  return (
    <main className="p-4">
      <h1 className="text-primary text-2xl font-semibold">{t("heading")}</h1>
      <Button type="button" onClick={reset} className="mt-4">
        {t("retry")}
      </Button>
    </main>
  );
}
