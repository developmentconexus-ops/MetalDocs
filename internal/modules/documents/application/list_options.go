package application

import "metaldocs/internal/modules/documents/repository"

// ListOptions is an alias for repository.ListOptions so handlers depend
// only on application-package types. The canonical struct lives in the
// repository package next to its SQL.
type ListOptions = repository.ListOptions
