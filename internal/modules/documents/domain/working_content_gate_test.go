package domain

import "testing"

func TestCanWriteWorkingContent(t *testing.T) {
	cases := []struct {
		name               string
		status             string
		isOwner            bool
		isEligibleApprover bool
		want               bool
	}{
		{"draft owner writes", string(DocStatusDraft), true, false, true},
		{"draft non-owner denied", string(DocStatusDraft), false, false, false},
		{"under_review eligible approver writes", string(DocStatusUnderReview), false, true, true},
		{"under_review author denied", string(DocStatusUnderReview), true, false, false},
		{"approved nobody writes", string(DocStatusApproved), true, true, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := CanWriteWorkingContent(c.status, c.isOwner, c.isEligibleApprover); got != c.want {
				t.Fatalf("CanWriteWorkingContent(%q,%v,%v)=%v want %v", c.status, c.isOwner, c.isEligibleApprover, got, c.want)
			}
		})
	}
}
