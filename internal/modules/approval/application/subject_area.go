package application

import (
	"context"
	"database/sql"
	"fmt"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	docapp "metaldocs/internal/modules/documents/application"
	"metaldocs/internal/modules/approval/domain"
)

// resolveSubjectAreaCode returns the process-area string to pass to
// authz.Require for a subject (M3 P3.S2b-3a). Document subjects derive the
// real area from the document via docapp.LoadDocumentAreaCode — identical to
// the pre-existing hardcoded call. Template subjects have no process area, so
// they assert area-blind via the 'tenant' sentinel: authz.go's area predicate
// (`$2='tenant' OR upa.area_code=$2`) treats 'tenant' as "skip area scoping",
// which is required here because a derived area would fail-close every
// template approver (templates are not process-area-scoped entities).
func resolveSubjectAreaCode(ctx context.Context, tx *sql.Tx, cdRead controlleddocumentsdomain.CDFieldReader, tenantID string, subject domain.Subject) (string, error) {
	switch subject.Kind {
	case domain.SubjectKindTemplate:
		return "tenant", nil
	case domain.SubjectKindDocument:
		areaCode, _, err := docapp.LoadDocumentAreaCode(ctx, tx, cdRead, tenantID, subject.Key)
		return areaCode, err
	default:
		return "", fmt.Errorf("approval: resolveSubjectAreaCode: unknown subject kind %q", subject.Kind)
	}
}
