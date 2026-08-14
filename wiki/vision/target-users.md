# Target Users

> **Last verified:** 2026-08-14
> **Status:** Active persona intent; exact permission bundles remain under Cohesive Platform Redesign.

## Primary personas

### 1. Tenant Owner / Product Administrator

Owns MetalDocs configuration for one company/tenant: people/groups/access assignments, document-type/policy configuration and tenant-level administration. This is **not** a platform-global superadmin and never bypasses domain invariants.

### 2. Area Manager

Operational manager for a business Area. Oversees working information and approval activity in that Area, can perform explicit administrative workflow operations where allowed, and manages document lifecycle operations such as obsolete/supersede when policy permits. It is not RBAC administration.

### 3. Author

Subject-matter expert who creates/edits eligible working revisions, collaborates/comments and submits governed information for approval. `created_by` is evidence, not exclusive edit ownership.

### 4. Approver

A qualified participant who receives concrete approval/review work through an Approval Step. Being an approver in an Area does not grant blanket access to every draft in that Area; authority to act requires both base qualification and participation in that case.

### 5. Viewer / Reader

Consumes released/effective official information in permitted scopes. Future distribution policies may additionally require read/acknowledgement evidence for specific released revisions.

## Organization model

Users may receive access directly or through flat Groups such as `Vendedores`, with RoleAssignments scoped to the Tenant or an Area such as `COMERCIAL`, `QUALIDADE`, `LOGISTICA`, `FINANCEIRO` or `RH`.

## Target organizations

Organizations that require disciplined operational-information control, including quality/safety/security-regulated or audit-sensitive environments. Product design should serve normal commercial organizations without importing pharmaceutical-level ceremony into every workflow; stronger reauthentication/signature/evidence rules are applied where the configured governance actually requires them.

## See also

- [product-vision.md](product-vision.md)
- [../architecture/cohesive-platform-redesign.md](../architecture/cohesive-platform-redesign.md)
