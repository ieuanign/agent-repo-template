"use client";

import { Button } from "@/components/Button";
import { COPY } from "@/ui/GlobalError/constants";
import type { GlobalErrorViewProps } from "@/ui/GlobalError/types";

export function GlobalErrorView({ reset }: GlobalErrorViewProps) {
  return (
    <main className="p-4">
      <h1 className="text-primary text-2xl font-semibold">{COPY.en.heading}</h1>
      <p lang="id" className="text-foreground">
        {COPY.id.heading}
      </p>
      {/* h-auto overrides the size's fixed height so both stacked lines fit. */}
      <Button
        type="button"
        onClick={reset}
        className="mt-4 h-auto flex-col py-1"
      >
        <span>{COPY.en.retry}</span>
        <span lang="id">{COPY.id.retry}</span>
      </Button>
    </main>
  );
}
