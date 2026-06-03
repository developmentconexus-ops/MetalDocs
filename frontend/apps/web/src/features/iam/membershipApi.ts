import { apiFetch } from "../../lib/api";

export interface AreaMembership {
  userId: string;
  tenantId: string;
  areaCode: string;
  role: string;
  effectiveFrom: string;
  effectiveTo: string | null;
  grantedBy: string | null;
}

export interface GrantMembershipRequest {
  userId: string;
  areaCode: string;
  role: 'viewer' | 'editor' | 'author' | 'approver' | 'signer' | 'area_admin' | 'qms_admin';
}

interface MembershipListResponse {
  items: AreaMembership[] | null;
}

const BASE = "/api/v1/iam/area-memberships";

export async function fetchMemberships(userId: string): Promise<AreaMembership[]> {
  const res = await apiFetch<MembershipListResponse>(`${BASE}?userId=${encodeURIComponent(userId)}`);
  return res.items ?? [];
}

export async function grantMembership(req: GrantMembershipRequest): Promise<void> {
  await apiFetch<void>(BASE, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
}

export async function revokeMembership(userId: string, areaCode: string): Promise<void> {
  await apiFetch<void>(
    `${BASE}?userId=${encodeURIComponent(userId)}&areaCode=${encodeURIComponent(areaCode)}`,
    { method: "DELETE" },
  );
}
