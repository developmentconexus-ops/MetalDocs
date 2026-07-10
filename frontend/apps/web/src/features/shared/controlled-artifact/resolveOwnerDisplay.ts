/**
 * Resolve the display name for a controlled artifact's owner (its `created_by`).
 * When the creator is the currently authenticated user, prefer their friendly
 * displayName; otherwise echo the raw creator id. Returns null when there is no
 * creator — callers that want an em-dash coalesce the null themselves.
 *
 * Deduplicated from useDocumentArtifact / useTemplateArtifact.
 */
export function resolveOwnerDisplay(
  createdBy: string | null | undefined,
  user: { userId?: string | null; displayName?: string | null } | null | undefined,
): string | null {
  if (createdBy == null) {
    return null;
  }
  if (user?.userId && createdBy === user.userId) {
    return user.displayName ?? createdBy;
  }
  return createdBy;
}
