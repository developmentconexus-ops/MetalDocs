import type { components } from "../../../lib/api-types";
import type { StatusFacet, LastLoginFacet } from "../components/UsersFilterBar";

export type ManagedUser = components["schemas"]["ManagedUserCore"];

export type DerivedStatus = "active" | "pending" | "suspended" | "locked";

export function deriveStatus(u: ManagedUser): DerivedStatus {
  const now = Date.now();
  if (u.locked_until && new Date(u.locked_until).getTime() > now) return "locked";
  if (!u.is_active) return "suspended";
  if (u.must_change_password && !u.last_login_at) return "pending";
  return "active";
}

const STATUS_LABEL: Record<DerivedStatus, string> = {
  active: "Ativo",
  pending: "Pendente",
  suspended: "Suspenso",
  locked: "Bloqueado",
};

export function statusLabel(s: DerivedStatus): string {
  return STATUS_LABEL[s];
}

export function matchesStatusFacet(u: ManagedUser, facet?: StatusFacet): boolean {
  if (!facet) return true;
  return deriveStatus(u) === facet;
}

const WINDOW_MS: Record<LastLoginFacet, number> = {
  "24h": 24 * 60 * 60 * 1000,
  "7d": 7 * 24 * 60 * 60 * 1000,
  "30d": 30 * 24 * 60 * 60 * 1000,
  "90d": 90 * 24 * 60 * 60 * 1000,
};

export function matchesLastLogin(u: ManagedUser, facet?: LastLoginFacet): boolean {
  if (!facet) return true;
  if (!u.last_login_at) return false;
  const ageMs = Date.now() - new Date(u.last_login_at).getTime();
  return ageMs >= 0 && ageMs <= WINDOW_MS[facet];
}

export function matchesMfa(u: ManagedUser, facet?: "on" | "off"): boolean {
  if (!facet) return true;
  const enabled = u.mfa_enabled === true;
  return facet === "on" ? enabled : !enabled;
}

export function matchesArea(u: ManagedUser, area?: string): boolean {
  if (!area) return true;
  return u.area_memberships.some((m) => m.area_code === area);
}
