package domain

import "errors"

var ErrISOSegregationViolation = errors.New("templates: iso_segregation_violation")
var ErrForbidden = errors.New("templates: forbidden")
var ErrForbiddenRole = errors.New("templates: forbidden_role")
var ErrUploadMissing = errors.New("templates: upload_missing")
var ErrInvalidApprovalConfig = errors.New("templates: invalid_approval_config")
var ErrPlaceholderIDEmpty = errors.New("placeholder id empty")
var ErrDuplicatePlaceholderID = errors.New("duplicate placeholder id")
var ErrPlaceholderNameInvalid = errors.New("placeholder name invalid")
var ErrDuplicatePlaceholderName = errors.New("duplicate placeholder name")
var ErrInvalidConstraint = errors.New("invalid constraint")
var ErrPlaceholderCycle = errors.New("placeholder visibility cycle")
var ErrUnknownResolver = errors.New("unknown resolver")
var ErrPlaceholderNotInCatalog = errors.New("placeholder name not in fixed catalog")
var ErrPlaceholderNotComputed = errors.New("catalog placeholders must be computed with resolver_key")
