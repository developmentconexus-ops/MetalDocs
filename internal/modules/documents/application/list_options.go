package application

import "metaldocs/internal/modules/documents/infrastructure"

// ListOptions is an alias for infrastructure.ListOptions so handlers depend
// only on application-package types. The canonical struct lives in the
// infrastructure package next to its SQL.
type ListOptions = infrastructure.ListOptions
