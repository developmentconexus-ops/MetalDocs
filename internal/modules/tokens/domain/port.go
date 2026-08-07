package domain

import (
	"context"

	"metaldocs/internal/platform/db"
)

// Repository is the storage port the application service consumes. All methods
// operate on the db.Tx the service supplies (the service owns the tx boundary
// via platform/db.TxRunner). The repo touches ONLY token_dictionary_entries.
type Repository interface {
	Create(ctx context.Context, tx db.Tx, e *Entry) (*Entry, error)
	Update(ctx context.Context, tx db.Tx, e *Entry) (*Entry, error)
	Delete(ctx context.Context, tx db.Tx, tenantID, id string) error
	GetByID(ctx context.Context, tx db.Tx, tenantID, id string) (*Entry, error)
	GetByName(ctx context.Context, tx db.Tx, tenantID, name string) (*Entry, error)
	List(ctx context.Context, tx db.Tx, tenantID string) ([]Entry, error)
}

// DictionaryReader is the provider port this module PUBLISHES for SP-2. The
// documents/controlleddocuments creation path reads dictionary values through it
// off-tx to pin them at document creation. Implemented by application.Service.
type DictionaryReader interface {
	GetByName(ctx context.Context, tenantID, name string) (*Entry, error)
	List(ctx context.Context, tenantID string) ([]Entry, error)
}
