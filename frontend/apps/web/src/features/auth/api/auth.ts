import type { CurrentUser, UserRole } from "../../../lib/types";
import type { components, operations } from "../../../lib/api-types";
import { request } from "../../../lib/api/client";

// FE-20: response types derived from the generated operations instead of a
// hand-rolled `WireCurrentUser` alias. Verified 1:1 against
// internal/modules/auth/delivery/http/handler.go (handleLogin/handleMe/
// handleChangePassword write authdomain.CurrentUser directly — a flat,
// non-enveloped body — matching components["schemas"]["CurrentUser"]).
//
// KNOWN CONTRACT GAP (flagged, not fixed here — api/openapi is out of scope
// for this task): api/openapi/v1/openapi.yaml's CurrentUser.roles enum only
// lists 5 of the 8 canonical roles (missing signer/area_admin/qms_admin),
// while internal/modules/iam/domain/model.go's Role enum and the OpenAPI
// Role schema both correctly have all 8. A user whose roles include one of
// the missing 3 would still receive it over the wire (Go has no such
// restriction) but the generated union would reject it at compile time. We
// keep the runtime `normalizeRoles` allowlist filter (all 8 values) rather
// than trusting the narrower generated union, and read `roles`/`capabilities`
// through a local cast at the one call site below.
type WireCurrentUser = components["schemas"]["CurrentUser"];
type LoginResponse = operations["login"]["responses"][200]["content"]["application/json"];
type ChangePasswordResponse = operations["changePassword"]["responses"][200]["content"]["application/json"];

const allowedRoles = new Set<UserRole>(["system_admin", "editor", "approver", "author", "viewer", "signer", "area_admin", "qms_admin"]);

function normalizeRoles(value: unknown): UserRole[] {
  if (!Array.isArray(value)) {
    return [];
  }
  return value.filter((item): item is UserRole => typeof item === "string" && allowedRoles.has(item as UserRole));
}

function normalizeCapabilities(value: unknown): string[] {
  if (!Array.isArray(value)) {
    return [];
  }
  return value.filter((item): item is string => typeof item === "string");
}

function normalizeCurrentUser(value: WireCurrentUser): CurrentUser {
  return {
    userId: value?.user_id ?? "",
    tenantId: value?.tenant_id ?? "",
    tenantName: value?.tenant_name ?? "",
    username: value?.username ?? "",
    email: value?.email ?? "",
    displayName: value?.display_name ?? value?.username ?? "",
    mustChangePassword: Boolean(value?.must_change_password),
    // roles is cast through `unknown` because the generated enum is narrower
    // than the runtime contract (see the contract-gap note above); the
    // allowlist filter in normalizeRoles is the real type guard here.
    roles: normalizeRoles(value?.roles as unknown),
    capabilities: normalizeCapabilities(value?.capabilities),
  };
}

export async function login(body: { identifier: string; password: string }) {
  const response = await request<LoginResponse>("/auth/login", {
    method: "POST",
    body: JSON.stringify(body),
  });
  return { ...response, user: normalizeCurrentUser(response.user) };
}

export function logout() {
  return request<void>("/auth/logout", { method: "POST" });
}

export async function me() {
  return normalizeCurrentUser(await request<WireCurrentUser>("/auth/me"));
}

export async function changePassword(body: { currentPassword: string; newPassword: string }) {
  const response = await request<ChangePasswordResponse>("/auth/change-password", {
    method: "POST",
    body: JSON.stringify({ current_password: body.currentPassword, new_password: body.newPassword }),
  });
  return { ...response, user: normalizeCurrentUser(response.user) };
}
