package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	v2domain "metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/modules/documents/infrastructure"
	templatesdomain "metaldocs/internal/modules/templates/domain"
	"metaldocs/internal/platform/db"
)

// SchemaReader loads the placeholder schema a document's fill-in values are
// validated against.
type SchemaReader interface {
	// LoadPlaceholderSchema loads the placeholder schema snapshot for a document.
	// docID is a documents.id (the snapshot lives on the documents row); the param
	// was historically misnamed revisionID even though the SQL filters WHERE id=$2
	// (F-D5, truth-in-naming — sibling TemplateVersionSchemaReader.LoadFillInSchema
	// uses docID identically).
	LoadPlaceholderSchema(ctx context.Context, tenantID, docID string) ([]templatesdomain.Placeholder, error)
}

// FillInWriter persists placeholder value writes and reads back a
// placeholder's current value source, used to enforce the author-editable
// governance rule (SP-2 D11).
type FillInWriter interface {
	UpsertValue(ctx context.Context, v infrastructure.PlaceholderValue, q ...infrastructure.DBTX) error
	UpsertAuthorValue(ctx context.Context, v infrastructure.PlaceholderValue, q ...infrastructure.DBTX) (int64, error)
	CurrentSource(ctx context.Context, tenantID, revisionID, placeholderID string) (string, bool, error)
}

// FillInService validates and writes author-supplied placeholder values for a
// document revision, enforcing document.edit authz and the author-editable
// governance rule that rejects writes to computed/dictionary-sourced rows.
type FillInService struct {
	runner        db.TxRunner
	schemas       SchemaReader
	writer        FillInWriter
	reader        FillInReader
	schemaFromTpl *TemplateVersionSchemaReader
	iam           IAMUserOptionsReader
	cdRead        controlleddocumentsdomain.CDFieldReader
}

// NewFillInService wires the service with a TxRunner for authz enforcement.
// Production callers MUST use this constructor — it enforces the document.edit
// capability (ADR 0022 Phase 10 merged the redundant doc.edit_draft cap into it).
// cdRead is the controlleddocuments read-port used by the area-grade authz check
// (M2/F2.1); a nil reader fail-closes the CD area term to "".
func NewFillInService(runner db.TxRunner, s SchemaReader, w FillInWriter) *FillInService {
	return &FillInService{runner: runner, schemas: s, writer: w}
}

// WithCDFieldReader attaches the controlleddocuments read-port used to resolve a
// document's controlled-document area in the document.edit authz check (M2/F2.1).
func (s *FillInService) WithCDFieldReader(r controlleddocumentsdomain.CDFieldReader) *FillInService {
	s.cdRead = r
	return s
}

// WithIAMReader attaches an IAMUserOptionsReader for validating user-typed placeholders.
func (s *FillInService) WithIAMReader(r IAMUserOptionsReader) *FillInService {
	s.iam = r
	return s
}

// SnapshotSchemaReader implements SchemaReader by reading the placeholder
// schema frozen onto a document's placeholder_schema_snapshot column.
type SnapshotSchemaReader struct {
	db *sql.DB
}

// NewSnapshotSchemaReader constructs a SnapshotSchemaReader over db.
func NewSnapshotSchemaReader(db *sql.DB) *SnapshotSchemaReader {
	return &SnapshotSchemaReader{db: db}
}

// LoadPlaceholderSchema implements SchemaReader by reading and parsing the
// document's placeholder_schema_snapshot column.
func (r *SnapshotSchemaReader) LoadPlaceholderSchema(ctx context.Context, tenantID, docID string) ([]templatesdomain.Placeholder, error) {
	var raw []byte
	if err := r.db.QueryRowContext(ctx, `
		SELECT placeholder_schema_snapshot
		  FROM documents
		 WHERE tenant_id=$1::uuid AND id=$2::uuid`, tenantID, docID).
		Scan(&raw); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, v2domain.ErrNotFound
		}
		return nil, err
	}
	return parsePlaceholderSchema(raw)
}

// parsePlaceholderSchema handles both storage formats:
//   - raw array:          [{"id":"..."}]           (eigenpal native, stored on template versions)
//   - wrapped object:     {"placeholders":[...]}   (legacy snapshot format)
func parsePlaceholderSchema(raw []byte) ([]templatesdomain.Placeholder, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	// Try raw array first (eigenpal format).
	var phs []templatesdomain.Placeholder
	if err := json.Unmarshal(raw, &phs); err == nil {
		return phs, nil
	}
	// Fallback: wrapped object format.
	var wrapped struct {
		Placeholders []templatesdomain.Placeholder `json:"placeholders"`
	}
	if err := json.Unmarshal(raw, &wrapped); err != nil {
		return nil, err
	}
	return wrapped.Placeholders, nil
}

// SetPlaceholderValue validates raw against placeholderID's schema and writes
// it as an author value on revisionID, after enforcing document.edit authz and
// rejecting writes to governed (computed/dictionary-sourced) rows.
func (s *FillInService) SetPlaceholderValue(ctx context.Context, tenantID, actorID, revisionID, placeholderID, raw string) error {
	if err := requireDocEditDraft(ctx, s.runner, s.cdRead, tenantID, actorID, revisionID); err != nil {
		return err
	}
	schema, err := s.schemas.LoadPlaceholderSchema(ctx, tenantID, revisionID)
	if err != nil {
		return err
	}
	if len(schema) == 0 && s.schemaFromTpl != nil {
		schema, err = s.schemaFromTpl.LoadFillInSchema(ctx, tenantID, revisionID)
		if err != nil {
			return err
		}
	}

	p, ok := findPlaceholder(schema, placeholderID)
	if !ok {
		return fmt.Errorf("%w: unknown placeholder %s", v2domain.ErrValidationFailed, placeholderID)
	}
	if err := validateValue(ctx, tenantID, p, raw, s.iam); err != nil {
		return err
	}

	// SP-2 D11: an author may write only rows whose current source is
	// author-editable ('user'/'default'). Governed rows (computed/dictionary) are
	// rejected — friendly app check first, DB WHERE-guard backstops.
	curSource, exists, err := s.writer.CurrentSource(ctx, tenantID, revisionID, placeholderID)
	if err != nil {
		return err
	}
	if exists && (curSource == "computed" || curSource == "dictionary") {
		return fmt.Errorf("%w: placeholder %s is governed (%s)", v2domain.ErrPlaceholderNotAuthorEditable, placeholderID, curSource)
	}

	value := raw
	affected, err := s.writer.UpsertAuthorValue(ctx, infrastructure.PlaceholderValue{
		TenantID:      tenantID,
		RevisionID:    revisionID,
		PlaceholderID: placeholderID,
		ValueText:     &value,
		Source:        "user",
	})
	if err != nil {
		return err
	}
	if affected == 0 {
		// DB guard rejected (governed row, or a race past the app check).
		return fmt.Errorf("%w: placeholder %s is governed", v2domain.ErrPlaceholderNotAuthorEditable, placeholderID)
	}
	return nil
}

func findPlaceholder(phs []templatesdomain.Placeholder, id string) (templatesdomain.Placeholder, bool) {
	for _, p := range phs {
		if p.ID == id {
			return p, true
		}
	}
	return templatesdomain.Placeholder{}, false
}

func validateValue(ctx context.Context, tenantID string, p templatesdomain.Placeholder, raw string, iam IAMUserOptionsReader) error {
	if p.Required && raw == "" {
		return fmt.Errorf("%w: %s required", v2domain.ErrValidationFailed, p.ID)
	}
	if p.MaxLength != nil && len(raw) > *p.MaxLength {
		return fmt.Errorf("%w: %s max_length exceeded", v2domain.ErrValidationFailed, p.ID)
	}
	if p.Regex != nil {
		re, err := regexp.Compile(*p.Regex)
		if err != nil {
			return err
		}
		if !re.MatchString(raw) {
			return fmt.Errorf("%w: %s regex mismatch", v2domain.ErrValidationFailed, p.ID)
		}
	}

	switch p.Type {
	case templatesdomain.PHNumber:
		n, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return fmt.Errorf("%w: %s not a number", v2domain.ErrValidationFailed, p.ID)
		}
		if p.MinNumber != nil && n < *p.MinNumber {
			return fmt.Errorf("%w: %s < min_number", v2domain.ErrValidationFailed, p.ID)
		}
		if p.MaxNumber != nil && n > *p.MaxNumber {
			return fmt.Errorf("%w: %s > max_number", v2domain.ErrValidationFailed, p.ID)
		}
	case templatesdomain.PHDate:
		if _, err := time.Parse("2006-01-02", raw); err != nil {
			return fmt.Errorf("%w: %s not YYYY-MM-DD", v2domain.ErrValidationFailed, p.ID)
		}
		if p.MinDate != nil && raw < *p.MinDate {
			return fmt.Errorf("%w: %s < min_date", v2domain.ErrValidationFailed, p.ID)
		}
		if p.MaxDate != nil && raw > *p.MaxDate {
			return fmt.Errorf("%w: %s > max_date", v2domain.ErrValidationFailed, p.ID)
		}
	case templatesdomain.PHSelect:
		for _, opt := range p.Options {
			if opt == raw {
				return nil
			}
		}
		return fmt.Errorf("%w: %s not in options", v2domain.ErrValidationFailed, p.ID)
	case templatesdomain.PHUser:
		if iam == nil {
			// Fail-closed: an unconfigured IAM reader must never silently accept any
			// value. Production wiring MUST call WithIAMReader (module.go); a missing
			// reader is a misconfiguration, not a permissive default (mirrors the
			// nil-area → deny posture of requireDocEditDraft in fillin_authz.go).
			return fmt.Errorf("%w: IAM reader not wired, cannot validate PHUser placeholder %q", v2domain.ErrValidationFailed, p.ID)
		}
		opts, err := iam.ListUserOptions(ctx, tenantID)
		if err != nil {
			return err
		}
		for _, o := range opts {
			if o.UserID == raw {
				return nil
			}
		}
		return fmt.Errorf("%w: %s unknown user %s", v2domain.ErrValidationFailed, p.ID, raw)
	case templatesdomain.PHText, templatesdomain.PHPicture, templatesdomain.PHComputed, templatesdomain.PHDictionary:
		// No additional type-specific validation: required/max_length/regex are
		// already enforced above. PHComputed/PHDictionary values can never reach
		// an actual write as author content — SetPlaceholderValue's CurrentSource
		// governance check rejects governed rows before the DB write runs.
	}

	return nil
}

// FillInReader reads current fill-in values from the DB.
type FillInReader interface {
	ListValues(ctx context.Context, tenantID, docID string) ([]infrastructure.PlaceholderValue, error)
}

// TemplateVersionSchemaReader reads fill-in schema from the template version
// referenced by the document. Works for draft documents without snapshots.
//
// M2/F2.4: documents resolves the version id from its OWN documents table, then
// reads templates_template_version.placeholder_schema through the templates-owned
// TemplateVersionPort (ADR-0030/ADR-0039 D3(b)) — no cross-module JOIN into the
// templates base table.
type TemplateVersionSchemaReader struct {
	db          *sql.DB
	tplVersions templatesdomain.TemplateVersionPort
}

// NewTemplateVersionSchemaReader constructs a TemplateVersionSchemaReader
// backed by db and the templates-owned TemplateVersionPort.
func NewTemplateVersionSchemaReader(db *sql.DB, tplVersions templatesdomain.TemplateVersionPort) *TemplateVersionSchemaReader {
	return &TemplateVersionSchemaReader{db: db, tplVersions: tplVersions}
}

// LoadFillInSchema resolves docID's template version on the documents table,
// then reads that version's placeholder schema through the templates-owned
// TemplateVersionPort.
func (r *TemplateVersionSchemaReader) LoadFillInSchema(ctx context.Context, tenantID, docID string) ([]templatesdomain.Placeholder, error) {
	// Resolve the document's template version on the documents-owned table.
	var versionID string
	err := r.db.QueryRowContext(ctx,
		`SELECT template_version_id FROM documents WHERE id = $1::uuid AND tenant_id = $2::uuid`,
		docID, tenantID,
	).Scan(&versionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	// Read the placeholder schema through the templates owner port.
	pRaw, err := r.tplVersions.PlaceholderSchema(ctx, tenantID, versionID)
	if err != nil {
		return nil, err
	}
	var placeholders []templatesdomain.Placeholder
	if len(pRaw) > 0 {
		if err := json.Unmarshal(pRaw, &placeholders); err != nil {
			return nil, err
		}
	}
	return placeholders, nil
}

// WithReader attaches a FillInReader for GET list operations.
func (s *FillInService) WithReader(r FillInReader) *FillInService {
	s.reader = r
	return s
}

// WithTemplateSchemaReader attaches a reader that loads schema from the template version.
func (s *FillInService) WithTemplateSchemaReader(r *TemplateVersionSchemaReader) *FillInService {
	s.schemaFromTpl = r
	return s
}

// GetPlaceholderValues returns docID's current fill-in values via the wired
// FillInReader.
func (s *FillInService) GetPlaceholderValues(ctx context.Context, tenantID, docID string) ([]infrastructure.PlaceholderValue, error) {
	if s.reader == nil {
		return nil, errors.New("fill-in reader not configured")
	}
	return s.reader.ListValues(ctx, tenantID, docID)
}

// GetFillInSchema returns docID's fill-in placeholder schema via the wired
// template-version schema reader.
func (s *FillInService) GetFillInSchema(ctx context.Context, tenantID, docID string) ([]templatesdomain.Placeholder, error) {
	if s.schemaFromTpl == nil {
		return nil, errors.New("fill-in schema reader not configured")
	}
	return s.schemaFromTpl.LoadFillInSchema(ctx, tenantID, docID)
}
