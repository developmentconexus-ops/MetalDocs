package main

import (
	"context"
	"errors"

	tokensdomain "metaldocs/internal/modules/tokens/domain"
)

// dictionaryValueReaderAdapter adapts the tokens module's published DictionaryReader
// to the documents-local DictionaryValueReader port. Mapping tokensdomain.ErrNotFound
// -> found=false keeps the documents module free of any tokens import (SP-2 §11,
// invariant #6). Lives at the composition root.
type dictionaryValueReaderAdapter struct {
	reader tokensdomain.DictionaryReader
}

func (a dictionaryValueReaderAdapter) Lookup(ctx context.Context, tenantID, name string) (string, bool, error) {
	e, err := a.reader.GetByName(ctx, tenantID, name)
	if err != nil {
		if errors.Is(err, tokensdomain.ErrNotFound) {
			return "", false, nil
		}
		return "", false, err
	}
	return e.Value, true, nil
}
