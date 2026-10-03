import { useTranslations } from "next-intl";

export function HomeView() {
  const t = useTranslations("home");
  return (
    <main className="p-4">
      <h1 className="text-primary text-2xl font-semibold">{t("heading")}</h1>
      <p>{t("tagline")}</p>
    </main>
  );
}
