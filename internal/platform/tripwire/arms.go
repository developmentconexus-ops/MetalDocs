// Package tripwire holds the Go source of truth for the
// enforce_capability_asserted() DB trigger's required-capability arms
// (M2 F2.1, docs/superpowers/milestones/global-maximum-remediation/
// milestone-2-authz-enforcement-generation/validation-contract.md §1.2 —
// binding). TripwireArms is the single ordered map from a gated (table, op)
// to the set of capabilities the trigger accepts; RenderMigration (render.go)
// regenerates the SQL migration byte-for-byte from this map so the Go
// registry, the generated SQL, and the live DB function can never silently
// drift apart (enforced by the TRIPWIRE-ARM-PARITY api-lint rule, Stage 2a).
//
// One-way import: this package imports internal/modules/iam/domain for the
// capability consts and IsValidCapability; iam/domain and its dependents
// never import this package. No cycle.
package tripwire

import (
	iamdomain "metaldocs/internal/modules/iam/domain"
)

// Op identifies the trigger operation an arm gates. OpAny means the arm
// applies regardless of TG_OP (the CASE branch tests table name only).
type Op string

const (
	OpInsert Op = "INSERT"
	OpUpdate Op = "UPDATE"
	OpAny    Op = "*"
)

// Arm is one CASE branch of enforce_capability_asserted(): the gated table,
// the operation it applies to (or OpAny), and the ordered set of
// capabilities the trigger accepts (any one present in metaldocs.asserted_caps
// satisfies the branch — match-one, not match-all).
type Arm struct {
	Table string
	Op    Op
	Caps  []iamdomain.Capability
}

// TripwireArms is the binding source of truth for every gated (table, op)
// pair enforced by public.enforce_capability_asserted(). Order is stable and
// mirrors the CASE-branch order in db/migrations/0270_*.sql exactly. The latest
// rendered migration is db/migrations/0275_*.sql; entry #6 (documents/UPDATE)
// has grown additively across two milestones: 0271 (M2) added document.obsolete
// + membership.manage per M2 validation-contract.md §1.1/§1.2; 0275 (M6 F6.2)
// added document.review per M6 validation-contract.md §3 — the intended
// registry-driven growth path, not drift.
//
// Content MUST equal M2 validation-contract.md §1.2 (18 gated entries) as
// extended by M6 validation-contract.md §3 (documents/UPDATE gains
// document.review). Any other deviation is HS-7 (see those binding clauses).
var TripwireArms = []Arm{
	{ // 1
		Table: "approval_instances",
		Op:    OpInsert,
		Caps:  []iamdomain.Capability{iamdomain.CapDocumentSubmit},
	},
	{ // 2
		Table: "approval_signoffs",
		Op:    OpInsert,
		Caps:  []iamdomain.Capability{iamdomain.CapDocumentSignoff},
	},
	{ // 3
		Table: "iam_user_roles",
		Op:    OpAny,
		Caps:  []iamdomain.Capability{iamdomain.CapUserManage},
	},
	{ // 4
		Table: "user_process_areas",
		Op:    OpAny,
		Caps:  []iamdomain.Capability{iamdomain.CapMembershipManage},
	},
	{ // 5
		Table: "documents",
		Op:    OpInsert,
		Caps:  []iamdomain.Capability{iamdomain.CapDocumentCreate},
	},
	{ // 6 — the documents/UPDATE arm. Widened (additive) across two milestones.
		// M2 (0271): covers two function-local latent P0001 incidents —
		// ForceReleaseSession / ForceReleaseSessionTx
		// (documents/repository/repository.go:798,828) assert only
		// CapMembershipManage; MarkObsolete's obsolete_service.go:88->93 asserts
		// only CapDocumentObsolete. Neither co-asserts document.edit, so the
		// pre-0271 arm ({document.edit}) fail-closed unconditionally on both
		// paths. See M2 validation-contract.md §1.1.
		// M6 (0275): gains CapDocumentReview — the F6.2 mark-reviewed workflow
		// asserts only document.review then UPDATEs documents (last_reviewed_at +
		// review_due_at); without the arm every mark-reviewed UPDATE is
		// fail-closed P0001, the same defect class as 0269/0270/0271. Additive
		// registry-driven growth per M6 validation-contract.md §3 (touchpoint 6).
		Table: "documents",
		Op:    OpUpdate,
		Caps: []iamdomain.Capability{
			iamdomain.CapDocumentEdit,
			iamdomain.CapDocumentObsolete,
			iamdomain.CapMembershipManage,
			iamdomain.CapDocumentReview,
		},
	},
	{ // 7
		Table: "controlled_documents",
		Op:    OpInsert,
		Caps:  []iamdomain.Capability{iamdomain.CapControlledDocumentCreate},
	},
	{ // 8
		Table: "controlled_documents",
		Op:    OpUpdate,
		Caps: []iamdomain.Capability{
			iamdomain.CapControlledDocumentObsolete,
			iamdomain.CapControlledDocumentSupersede,
		},
	},
	{ // 9
		Table: "cd_sequence_counters",
		Op:    OpAny,
		Caps:  []iamdomain.Capability{iamdomain.CapControlledDocumentCreate},
	},
	{ // 10
		Table: "document_profiles",
		Op:    OpAny,
		Caps:  []iamdomain.Capability{iamdomain.CapTaxonomyManage},
	},
	{ // 11
		Table: "document_process_areas",
		Op:    OpAny,
		Caps:  []iamdomain.Capability{iamdomain.CapTaxonomyManage},
	},
	{ // 12
		Table: "document_families",
		Op:    OpAny,
		Caps:  []iamdomain.Capability{iamdomain.CapTaxonomyManage},
	},
	{ // 13 — 'template.submit' is a deliberately retained harmless superset
		// (no writer asserts submit while writing this table); pruning is a
		// tightening deferred to M9 arm-hygiene. See validation-contract.md §1.3.
		Table: "templates_template",
		Op:    OpAny,
		Caps: []iamdomain.Capability{
			iamdomain.CapTemplateCreate,
			iamdomain.CapTemplateEdit,
			iamdomain.CapTemplateSubmit,
			iamdomain.CapTemplateApprove,
			iamdomain.CapTemplatePublish,
			iamdomain.CapTemplateArchive,
		},
	},
	{ // 14
		Table: "templates_template_version",
		Op:    OpAny,
		Caps: []iamdomain.Capability{
			iamdomain.CapTemplateCreate,
			iamdomain.CapTemplateEdit,
			iamdomain.CapTemplateSubmit,
			iamdomain.CapTemplateReview,
			iamdomain.CapTemplateApprove,
			iamdomain.CapTemplatePublish,
		},
	},
	{ // 15
		Table: "iam_users",
		Op:    OpAny,
		Caps:  []iamdomain.Capability{iamdomain.CapUserManage},
	},
	{ // 16
		Table: "iam_groups",
		Op:    OpAny,
		Caps:  []iamdomain.Capability{iamdomain.CapUserManage},
	},
	{ // 17
		Table: "iam_group_members",
		Op:    OpAny,
		Caps:  []iamdomain.Capability{iamdomain.CapUserManage},
	},
	{ // 18 — tenant scoped via group_id (v_tenant_id := NULL in the trigger);
		// iam_group_roles has no tenant_id column.
		Table: "iam_group_roles",
		Op:    OpAny,
		Caps:  []iamdomain.Capability{iamdomain.CapUserManage},
	},
}
