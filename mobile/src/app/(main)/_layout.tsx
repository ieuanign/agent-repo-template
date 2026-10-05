import { Stack, type ErrorBoundaryProps } from "expo-router";
import { useEffect } from "react";

import { reportError } from "@/core/reporting";
import ErrorScreen from "@/ui/Error";

export function ErrorBoundary({ error, retry }: ErrorBoundaryProps) {
  useEffect(() => {
    reportError(error);
  }, [error]);

  return <ErrorScreen retry={retry} />;
}

export default function MainLayout() {
  return <Stack screenOptions={{ headerShown: false }} />;
}
