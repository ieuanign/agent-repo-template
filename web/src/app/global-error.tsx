"use client";

import { useEffect } from "react";
import { reportError } from "@/core/reporting";
import { GlobalErrorView } from "@/ui/GlobalError";
import { TITLE } from "@/ui/GlobalError/constants";
import "./globals.css";

export default function GlobalError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    reportError(error);
  }, [error]);

  return (
    <html lang="en">
      <body>
        <title>{TITLE}</title>
        <GlobalErrorView reset={reset} />
      </body>
    </html>
  );
}
