import type { components } from "../../../lib/api-types";
import type { StatusFacet, LastLoginFacet } from "../components/UsersFilterBar";

export type ManagedUser = components["schemas"]["ManagedUserCore"];

export type DerivedStatus = "active" | "pending" | "suspended" | "locked";

export function deriveStatus(u: ManagedUser): DerivedStatus {
  const now = Date.now();
  if (u.lockedUntil && new Date(u.lockedUntil).getTime() > now) return "locked";
  if (!u.isActive) return "suspended";
  if (u.mustChangePassword && !u.lastLoginAt) return "pending";
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
  if (!u.lastLoginAt) return false;
  const ageMs = Date.now() - new Date(u.lastLoginAt).getTime();
  return ageMs >= 0 && ageMs <= WINDOW_MS[facet];
}

export function matchesMfa(u: ManagedUser, facet?: "on" | "off"): boolean {
  if (!facet) return true;
  const enabled = u.mfaEnabled === true;
  return facet === "on" ? enabled : !enabled;
}

export function matchesArea(u: ManagedUser, area?: string): boolean {
  if (!area) return true;
  return u.areaMemberships.some((m) => m.areaCode === area);
}
