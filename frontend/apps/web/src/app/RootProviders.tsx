import type { ReactNode } from "react";

type RootProvidersProps = {
  children: ReactNode;
};

export function RootProviders({ children }: RootProvidersProps) {
  return <>{children}</>;
}
