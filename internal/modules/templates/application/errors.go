package application

import (
	"errors"
	"fmt"

	"metaldocs/internal/modules/templates/domain"
)

func wrapAppErr(op string, err error) error {
	if err == nil {
		return nil
	}
	if isDomainErr(err) {
		return err
	}
	return fmt.Errorf("%s: %w", op, err)
}

func isDomainErr(err error) bool {
	switch {
	case errors.Is(err, domain.ErrNotFound),
		errors.Is(err, domain.ErrKeyConflict),
		errors.Is(err, domain.ErrSystemTemplateImmutable),
		errors.Is(err, domain.ErrArchived),
		errors.Is(err, domain.ErrInvalidStateTransition),
		errors.Is(err, domain.ErrContentHashMismatch),
		errors.Is(err, domain.ErrStaleBase),
		errors.Is(err, domain.ErrStaleLockVersion),
		errors.Is(err, domain.ErrUploadMissing),
		errors.Is(err, domain.ErrISOSegregationViolation),
		errors.Is(err, domain.ErrForbidden),
		errors.Is(err, domain.ErrForbiddenRole),
		errors.Is(err, domain.ErrInvalidApprovalConfig),
		errors.Is(err, domain.ErrPlaceholderIDEmpty),
		errors.Is(err, domain.ErrDuplicatePlaceholderID),
		errors.Is(err, domain.ErrPlaceholderNameInvalid),
		errors.Is(err, domain.ErrDuplicatePlaceholderName),
		errors.Is(err, domain.ErrInvalidConstraint),
		errors.Is(err, domain.ErrPlaceholderCycle),
		errors.Is(err, domain.ErrUnknownResolver),
		errors.Is(err, domain.ErrPlaceholderNotInCatalog),
		errors.Is(err, domain.ErrPlaceholderNotComputed),
		errors.Is(err, domain.ErrTransactionRequired):
		return true
	default:
		return false
	}
}
