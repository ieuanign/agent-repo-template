"use client";

import { useEffect } from "react";
import { reportError } from "@/core/reporting";
import { ErrorView } from "@/ui/Error";

export default function ErrorPage({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    reportError(error);
  }, [error]);

  return <ErrorView reset={reset} />;
}
