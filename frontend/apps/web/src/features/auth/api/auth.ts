import type { CurrentUser, UserRole } from "../../../lib/types";
import type { components, operations } from "../../../lib/api-types";
import { isUserRole } from "../../../lib/iam/role-vocabulary";
import { request } from "../../../lib/api/client";

// FE-20: response types derived from the generated operations instead of a
// hand-rolled `WireCurrentUser` alias. Verified 1:1 against
// internal/modules/auth/delivery/http/handler.go (handleLogin/handleMe/
// handleChangePassword write authdomain.CurrentUser directly — a flat,
// non-enveloped body — matching components["schemas"]["CurrentUser"]).
//
// The former "known contract gap" note here is gone (F-QA4-2): CurrentUser.roles
// $refs the UserRole schema, which carries all 8 canonical roles, so the
// generated union is exactly as wide as the runtime contract. `normalizeRoles`
// stays as a defensive wire filter, but its allowlist is now DERIVED from the
// generated enum rather than hand-listed.
type WireCurrentUser = components["schemas"]["CurrentUser"];
type LoginResponse = operations["login"]["responses"][200]["content"]["application/json"];
type ChangePasswordResponse = operations["changePassword"]["responses"][200]["content"]["application/json"];

function normalizeRoles(value: unknown): UserRole[] {
  if (!Array.isArray(value)) {
    return [];
  }
  return value.filter(isUserRole);
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
    // normalizeRoles narrows via the generated-enum guard, so an unexpected
    // wire value is dropped rather than trusted.
    roles: normalizeRoles(value?.roles),
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
