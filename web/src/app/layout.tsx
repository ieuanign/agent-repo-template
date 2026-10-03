import type { Metadata } from "next";
import { Inter } from "next/font/google";
import { NextIntlClientProvider } from "next-intl";
import { getLocale, getMessages } from "next-intl/server";
import { pick } from "@/lib/i18n/utils";
import "./globals.css";

const inter = Inter({
  subsets: ["latin"],
  variable: "--font-inter",
});

// Everything sits behind sign-in, so no page is for search engines.
export const metadata: Metadata = {
  robots: { index: false },
};

export default async function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const [locale, messages] = await Promise.all([getLocale(), getMessages()]);
  return (
    <html lang={locale} className={inter.variable}>
      <body>
        {/* Only namespaces client components read; without the prop next-intl ships every one. */}
        <NextIntlClientProvider messages={pick(messages, ["error"])}>
          {children}
        </NextIntlClientProvider>
      </body>
    </html>
  );
}
