package application

// ViewResult is the output of ViewService.GetViewURL.
// Defined here (application layer) so that both the service and the delivery
// handler can reference it without creating an import cycle.
type ViewResult struct {
	PDFStatus string // "pending" | "ready" | "failed"
	SignedURL string // populated only when PDFStatus == "ready"
}
