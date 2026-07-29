package domain

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	docsdomain "metaldocs/internal/modules/documents/domain"
)

func tp(t time.Time) *time.Time { return &t }

var evalNow = time.Date(2026, 7, 28, 12, 0, 0, 0, time.UTC)

// TestEvaluateRelease_DecisionTable walks every branch of the ADR 0085 release
// predicate. The table IS the decision table — a new conjunct without a row
// here is an untested branch.
func TestEvaluateRelease_DecisionTable(t *testing.T) {
	// ready is the fully-satisfied baseline every case perturbs by one field.
	ready := ReleaseInputs{
		IsFreezeHead:   true,
		ApprovalFact:   true,
		ArtifactFact:   true,
		DocumentStatus: "approved",
	}
	mutate := func(f func(*ReleaseInputs)) ReleaseInputs {
		in := ready
		f(&in)
		return in
	}

	cases := []struct {
		name       string
		in         ReleaseInputs
		wantAction ReleaseAction
		wantHold   ReleaseHoldReason
	}{
		{
			name:       "fully ready releases",
			in:         ready,
			wantAction: ReleaseActionRelease,
		},
		{
			name:       "already released is a noop",
			in:         mutate(func(i *ReleaseInputs) { i.Released = true }),
			wantAction: ReleaseActionNoop,
		},
		{
			// Superseded generation: NOT a hold — recording one would park a
			// permanently-stuck row in the readiness projection.
			name:       "not the freeze head is a noop, never a hold",
			in:         mutate(func(i *ReleaseInputs) { i.IsFreezeHead = false }),
			wantAction: ReleaseActionNoop,
		},
		{
			name:       "draft document is a noop",
			in:         mutate(func(i *ReleaseInputs) { i.DocumentStatus = "draft" }),
			wantAction: ReleaseActionNoop,
		},
		{
			name:       "already published document is a noop",
			in:         mutate(func(i *ReleaseInputs) { i.DocumentStatus = "published" }),
			wantAction: ReleaseActionNoop,
		},
		{
			name:       "scheduled document is releasable",
			in:         mutate(func(i *ReleaseInputs) { i.DocumentStatus = "scheduled" }),
			wantAction: ReleaseActionRelease,
		},
		{
			name:       "missing approval fact holds awaiting_approval_fact",
			in:         mutate(func(i *ReleaseInputs) { i.ApprovalFact = false }),
			wantAction: ReleaseActionHold,
			wantHold:   HoldAwaitingApprovalFact,
		},
		{
			name:       "missing artifact fact holds materializing",
			in:         mutate(func(i *ReleaseInputs) { i.ArtifactFact = false }),
			wantAction: ReleaseActionHold,
			wantHold:   HoldMaterializing,
		},
		{
			name:       "elapsed effective_to holds plan_invalid",
			in:         mutate(func(i *ReleaseInputs) { i.EffectiveTo = tp(evalNow.Add(-time.Hour)) }),
			wantAction: ReleaseActionHold,
			wantHold:   HoldPlanInvalid,
		},
		{
			// Boundary: effective_to == now is NOT after now, so the window is
			// already closed. ck_documents_effective_window is strict.
			name:       "effective_to exactly now holds plan_invalid",
			in:         mutate(func(i *ReleaseInputs) { i.EffectiveTo = tp(evalNow) }),
			wantAction: ReleaseActionHold,
			wantHold:   HoldPlanInvalid,
		},
		{
			name:       "future effective_to still releases",
			in:         mutate(func(i *ReleaseInputs) { i.EffectiveTo = tp(evalNow.Add(time.Hour)) }),
			wantAction: ReleaseActionRelease,
		},
		{
			name:       "future planned date schedules with awaiting_effective_date",
			in:         mutate(func(i *ReleaseInputs) { i.PlannedEffectiveFrom = tp(evalNow.Add(24 * time.Hour)) }),
			wantAction: ReleaseActionSchedule,
			wantHold:   HoldAwaitingEffectiveDate,
		},
		{
			// Boundary: planned == now is NOT after now, so the gate is open.
			name:       "planned date exactly now releases",
			in:         mutate(func(i *ReleaseInputs) { i.PlannedEffectiveFrom = tp(evalNow) }),
			wantAction: ReleaseActionRelease,
		},
		{
			name:       "elapsed planned date releases",
			in:         mutate(func(i *ReleaseInputs) { i.PlannedEffectiveFrom = tp(evalNow.Add(-24 * time.Hour)) }),
			wantAction: ReleaseActionRelease,
		},
		{
			// Ordering: plan validity is checked BEFORE the effective-date
			// gate, so an unsatisfiable plan never parks on a timer.
			name: "invalid plan wins over a future planned date",
			in: mutate(func(i *ReleaseInputs) {
				i.PlannedEffectiveFrom = tp(evalNow.Add(24 * time.Hour))
				i.EffectiveTo = tp(evalNow.Add(-time.Hour))
			}),
			wantAction: ReleaseActionHold,
			wantHold:   HoldPlanInvalid,
		},
		{
			// Ordering: the head check precedes the fact checks, so a stale
			// generation never reports "materializing".
			name: "stale generation with no artifact fact is still a noop",
			in: mutate(func(i *ReleaseInputs) {
				i.IsFreezeHead = false
				i.ArtifactFact = false
			}),
			wantAction: ReleaseActionNoop,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := EvaluateRelease(tc.in, evalNow)
			if got.Action != tc.wantAction {
				t.Fatalf("Action = %q, want %q", got.Action, tc.wantAction)
			}
			if got.Hold != tc.wantHold {
				t.Fatalf("Hold = %q, want %q", got.Hold, tc.wantHold)
			}
		})
	}
}

func TestResolveActualReviewDueAt(t *testing.T) {
	actual := time.Date(2026, 7, 28, 12, 0, 0, 0, time.UTC)

	cases := []struct {
		name     string
		planned  *time.Time
		interval int
		want     *time.Time
	}{
		{
			name:     "planned date after actual is kept verbatim",
			planned:  tp(actual.Add(90 * 24 * time.Hour)),
			interval: 365,
			want:     tp(actual.Add(90 * 24 * time.Hour)),
		},
		{
			name:     "planned date exactly at actual is kept",
			planned:  tp(actual),
			interval: 365,
			want:     tp(actual),
		},
		{
			// A plan whose review date already elapsed by the time the release
			// actually happened would violate ck_documents_review_due_sane.
			name:     "stale planned date is recomputed from actual",
			planned:  tp(actual.Add(-24 * time.Hour)),
			interval: 30,
			want:     tp(actual.AddDate(0, 0, 30)),
		},
		{
			name:     "no plan derives from the profile interval",
			planned:  nil,
			interval: 365,
			want:     tp(actual.AddDate(0, 0, 365)),
		},
		{
			name:     "no plan and no interval means no review cycle",
			planned:  nil,
			interval: 0,
			want:     nil,
		},
		{
			// No usable interval to recompute from: nil beats a constraint
			// violation. The document simply has no review cycle.
			name:     "stale plan with no interval is dropped",
			planned:  tp(actual.Add(-24 * time.Hour)),
			interval: 0,
			want:     nil,
		},
		{
			name:     "negative interval is treated as absent",
			planned:  nil,
			interval: -5,
			want:     nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ResolveActualReviewDueAt(tc.planned, actual, tc.interval)
			switch {
			case tc.want == nil && got != nil:
				t.Fatalf("got %v, want nil", *got)
			case tc.want != nil && got == nil:
				t.Fatalf("got nil, want %v", *tc.want)
			case tc.want != nil && !got.Equal(*tc.want):
				t.Fatalf("got %v, want %v", *got, *tc.want)
			}
		})
	}
}

// TestResolveActualReviewDueAt_NeverPrecedesEffectiveFrom is the invariant the
// individual cases above exist to protect: whatever comes back must satisfy
// ck_documents_review_due_sane (review_due_at >= effective_from).
func TestResolveActualReviewDueAt_NeverPrecedesEffectiveFrom(t *testing.T) {
	actual := time.Date(2026, 7, 28, 12, 0, 0, 0, time.UTC)
	offsets := []time.Duration{-720 * time.Hour, -time.Hour, 0, time.Hour, 720 * time.Hour}
	intervals := []int{-1, 0, 1, 30, 365}
	for _, off := range offsets {
		for _, iv := range intervals {
			planned := tp(actual.Add(off))
			for _, p := range []*time.Time{nil, planned} {
				got := ResolveActualReviewDueAt(p, actual, iv)
				if got != nil && got.Before(actual) {
					t.Fatalf("planned=%v interval=%d produced %v, before effective_from %v", p, iv, *got, actual)
				}
			}
		}
	}
}

// TestReleaseSourceStatus_ParityWithDocumentsDomain pins the two locally
// declared source statuses to the documents module's canonical values. The
// approval kernel deliberately does not import them at runtime (the predicate
// stays cross-module-dependency-free); this test is what keeps that copy
// honest.
func TestReleaseSourceStatus_ParityWithDocumentsDomain(t *testing.T) {
	if string(DocStatusApprovedForRelease) != string(docsdomain.DocStatusApproved) {
		t.Fatalf("approved parity: %q vs %q", DocStatusApprovedForRelease, docsdomain.DocStatusApproved)
	}
	if string(DocStatusScheduledForRelease) != string(docsdomain.DocStatusScheduled) {
		t.Fatalf("scheduled parity: %q vs %q", DocStatusScheduledForRelease, docsdomain.DocStatusScheduled)
	}
}

// TestReleaseHoldReasons_ParityWithMigrationCheck parses the
// ck_release_generations_hold_reason CHECK out of migration 0310 and asserts it
// carries exactly the Go vocabulary. A reason added on one side only would let
// the coordinator write a value the DB rejects (or leave a DB value no code can
// produce) — a class of drift the "hand-synced enumerations" meta-defect has
// bitten this repo with before.
func TestReleaseHoldReasons_ParityWithMigrationCheck(t *testing.T) {
	path := filepath.Join("..", "..", "..", "..", "db", "migrations", "0310_release_coordinator.sql")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	block := regexp.MustCompile(`(?s)ck_release_generations_hold_reason\s+CHECK \(hold_reason IS NULL OR hold_reason IN \((.*?)\)\)`).
		FindStringSubmatch(string(raw))
	if block == nil {
		t.Fatal("ck_release_generations_hold_reason CHECK not found in migration 0310")
	}
	var fromSQL []string
	for _, lit := range regexp.MustCompile(`'([a-z_]+)'`).FindAllStringSubmatch(block[1], -1) {
		fromSQL = append(fromSQL, lit[1])
	}
	var fromGo []string
	for _, r := range ReleaseHoldReasons() {
		fromGo = append(fromGo, string(r))
	}
	if strings.Join(fromSQL, ",") != strings.Join(fromGo, ",") {
		t.Fatalf("hold reason vocabulary drift:\n  migration: %v\n  go:        %v", fromSQL, fromGo)
	}
}
