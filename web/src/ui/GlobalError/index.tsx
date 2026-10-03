"use client";

import { Button } from "@/components/Button";
import { COPY } from "@/ui/GlobalError/constants";

export function GlobalErrorView({ reset }: { reset: () => void }) {
  return (
    <main className="p-4">
      <h1 className="text-primary text-2xl font-semibold">{COPY.id.heading}</h1>
      <p lang="en" className="text-foreground">
        {COPY.en.heading}
      </p>
      {/* h-auto overrides the size's fixed height so both stacked lines fit. */}
      <Button
        type="button"
        onClick={reset}
        className="mt-4 h-auto flex-col py-1"
      >
        <span>{COPY.id.retry}</span>
        <span lang="en">{COPY.en.retry}</span>
      </Button>
    </main>
  );
}
