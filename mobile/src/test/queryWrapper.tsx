import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { ReactNode } from "react";
import { IntlProvider } from "use-intl";

import { messages } from "@/lib/i18n/messages";

// A fresh client per render, so no answer leaks from one test into the next.
export function queryWrapper() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false, gcTime: 0 } },
  });
  return function Wrapper({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={client}>
        <IntlProvider locale="id" messages={messages.id}>
          {children}
        </IntlProvider>
      </QueryClientProvider>
    );
  };
}
