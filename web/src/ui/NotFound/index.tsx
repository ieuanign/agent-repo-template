import { useTranslations } from "next-intl";

export function NotFoundView() {
  const t = useTranslations("notFound");
  return (
    <main className="p-4">
      <h1 className="text-primary text-2xl font-semibold">{t("heading")}</h1>
      <p>{t("body")}</p>
    </main>
  );
}
