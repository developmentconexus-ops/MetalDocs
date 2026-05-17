# V2 Operation ID Inventory

Generated operation IDs still carrying release-facing `V2` before Phase 6 cleanup.

## Rename in Phase 6

| Module | Old operationId | New operationId | Route behavior |
|---|---|---|---|
| IAM | `listAreaMembershipsV2` | `listAreaMemberships` | unchanged |
| IAM | `grantAreaMembershipV2` | `grantAreaMembership` | unchanged |
| IAM | `revokeAreaMembershipV2` | `revokeAreaMembership` | unchanged |
| taxonomy | `listTaxonomyProfilesV2` | `listTaxonomyProfiles` | unchanged |
| taxonomy | `createTaxonomyProfileV2` | `createTaxonomyProfile` | unchanged |
| taxonomy | `getTaxonomyProfileV2` | `getTaxonomyProfile` | unchanged |
| taxonomy | `updateTaxonomyProfileV2` | `updateTaxonomyProfile` | unchanged |
| taxonomy | `archiveTaxonomyProfileV2` | `archiveTaxonomyProfile` | unchanged |
| taxonomy | `setTaxonomyProfileDefaultTemplateV2` | `setTaxonomyProfileDefaultTemplate` | unchanged |
| taxonomy | `listTaxonomyAreasV2` | `listTaxonomyAreas` | unchanged |
| taxonomy | `createTaxonomyAreaV2` | `createTaxonomyArea` | unchanged |
| taxonomy | `getTaxonomyAreaV2` | `getTaxonomyArea` | unchanged |
| taxonomy | `updateTaxonomyAreaV2` | `updateTaxonomyArea` | unchanged |
| taxonomy | `archiveTaxonomyAreaV2` | `archiveTaxonomyArea` | unchanged |
| taxonomy | `listTaxonomyFamiliesV2` | `listTaxonomyFamilies` | unchanged |
| taxonomy | `createTaxonomyFamilyV2` | `createTaxonomyFamily` | unchanged |
| taxonomy | `getTaxonomyFamilyV2` | `getTaxonomyFamily` | unchanged |
| taxonomy | `updateTaxonomyFamilyV2` | `updateTaxonomyFamily` | unchanged |
| taxonomy | `deactivateTaxonomyFamilyV2` | `deactivateTaxonomyFamily` | unchanged |
| approval | `recordApprovalStageSignoffV2` | `recordApprovalStageSignoff` | unchanged |
| approval | `cancelApprovalInstanceV2` | `cancelApprovalInstance` | unchanged |
| approval | `getApprovalInstanceV2` | `getApprovalInstance` | unchanged |
| approval | `listApprovalInboxV2` | `listApprovalInbox` | unchanged |
| approval | `createApprovalRouteV2` | `createApprovalRoute` | unchanged |
| approval | `listApprovalRoutesV2` | `listApprovalRoutes` | unchanged |
| approval | `updateApprovalRouteV2` | `updateApprovalRoute` | unchanged |
| approval | `deactivateApprovalRouteV2` | `deactivateApprovalRoute` | unchanged |

## Deferred

`docgen-v2` service and package names are not renamed in Phase 6. They require Phase 7 compatibility classification because they may be deployment, environment, or service-bound names rather than generated API operation names.