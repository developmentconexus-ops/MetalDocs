import { useAuthStore } from "../../../store/auth.store";

export function useHasCapability(cap: string): boolean {
  const caps = useAuthStore((s) => s.user?.capabilities ?? []);
  return caps.includes(cap);
}
