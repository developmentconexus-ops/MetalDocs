package domain

// VersionRef is a compact value-object reference to a template version:
// internal version counter, regulated revision number (ADR 0013), and
// lifecycle status. ADR 0065 — version pointers travel as one value, so the
// id/number/revision coupling can never drift apart.
type VersionRef struct {
	ID             string
	Number         int
	RevisionNumber int
	Status         VersionStatus
}

// TemplateRead is the read model returned by repository reads: the aggregate
// plus the latest and published version references projected by the list/get
// LEFT JOINs. Published is nil when the template has never been published.
// The write-side aggregate (Template) does not carry these projections.
type TemplateRead struct {
	Template
	Latest    VersionRef
	Published *VersionRef
}
