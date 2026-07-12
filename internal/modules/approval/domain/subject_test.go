package domain

import "testing"

func TestSubjectValidateDocumentHappy(t *testing.T) {
	s := NewDocumentSubject("SOP")
	if err := s.Validate(); err != nil {
		t.Fatalf("valid document subject rejected: %v", err)
	}
	if s.Kind != SubjectKindDocument {
		t.Fatalf("expected Kind=document, got %q", s.Kind)
	}
	if s.Key != "SOP" {
		t.Fatalf("expected Key=SOP, got %q", s.Key)
	}
}

func TestSubjectValidateTemplateHappy(t *testing.T) {
	s := Subject{Kind: SubjectKindTemplate, Key: "tmpl-1"}
	if err := s.Validate(); err != nil {
		t.Fatalf("valid template subject rejected: %v", err)
	}
}

func TestSubjectValidateInvalidKind(t *testing.T) {
	s := Subject{Kind: "bogus", Key: "x"}
	if err := s.Validate(); err == nil {
		t.Error("invalid kind should fail validation")
	}
}

func TestSubjectValidateEmptyKey(t *testing.T) {
	s := Subject{Kind: SubjectKindDocument, Key: ""}
	if err := s.Validate(); err == nil {
		t.Error("empty key should fail validation")
	}
}

func TestSubjectEqual(t *testing.T) {
	a := NewDocumentSubject("doc-1")
	b := NewDocumentSubject("doc-1")
	c := NewDocumentSubject("doc-2")
	if !a.Equal(b) {
		t.Error("expected equal subjects to be Equal")
	}
	if a.Equal(c) {
		t.Error("expected different-key subjects to not be Equal")
	}
}

func TestSubjectString(t *testing.T) {
	s := NewDocumentSubject("doc-1")
	if got, want := s.String(), "document:doc-1"; got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
}

// TestRouteCarriesDocumentSubject asserts that a document Route constructed
// through the normal domain path carries Subject == {document, ProfileCode}
// — the projection invariant P2.S2 establishes so every existing ProfileCode
// consumer keeps working unchanged.
func TestRouteCarriesDocumentSubject(t *testing.T) {
	r := Route{
		ID: "r1", TenantID: "t1", ProfileCode: "SOP", Version: 1,
		Subject: NewDocumentSubject("SOP"),
		Stages: []Stage{
			{Order: 1, Name: "QA", RequiredRole: "approver", RequiredCapability: "document.signoff", AreaCode: "qa", Quorum: QuorumAny1Of, OnEligibilityDrift: DriftReduceQuorum},
		},
	}
	if r.Subject.Kind != SubjectKindDocument {
		t.Fatalf("expected Subject.Kind=document, got %q", r.Subject.Kind)
	}
	if r.Subject.Key != r.ProfileCode {
		t.Fatalf("expected Subject.Key == ProfileCode, got Subject.Key=%q ProfileCode=%q", r.Subject.Key, r.ProfileCode)
	}
}

// TestInstanceCarriesDocumentSubject is the Instance-side sibling of
// TestRouteCarriesDocumentSubject.
func TestInstanceCarriesDocumentSubject(t *testing.T) {
	inst := Instance{
		ID: "i1", TenantID: "t1", DocumentID: "doc-1",
		Subject: NewDocumentSubject("doc-1"),
	}
	if inst.Subject.Kind != SubjectKindDocument {
		t.Fatalf("expected Subject.Kind=document, got %q", inst.Subject.Kind)
	}
	if inst.Subject.Key != inst.DocumentID {
		t.Fatalf("expected Subject.Key == DocumentID, got Subject.Key=%q DocumentID=%q", inst.Subject.Key, inst.DocumentID)
	}
}
