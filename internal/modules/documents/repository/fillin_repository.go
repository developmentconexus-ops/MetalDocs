package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
)

// FillInRepository manages document_placeholder_values rows.
type FillInRepository struct {
	db     *sql.DB
	schema string
}

type PlaceholderValue struct {
	TenantID        string
	RevisionID      string
	PlaceholderID   string
	ValueText       *string
	ValueTyped      map[string]any
	Source          string
	ComputedFrom    *string
	ResolverVersion *int
	InputsHash      []byte
}

// NewFillInRepository creates a FillInRepository using bare table names.
func NewFillInRepository(db *sql.DB) *FillInRepository {
	return &FillInRepository{db: db}
}

// NewFillInRepositoryWithSchema creates a FillInRepository that qualifies
// table names with the given schema. Used by integration tests.
func NewFillInRepositoryWithSchema(db *sql.DB, schema string) *FillInRepository {
	return &FillInRepository{db: db, schema: schema}
}

func (r *FillInRepository) table(name string) string {
	if r.schema == "" {
		return name
	}
	return fmt.Sprintf("%q.%q", r.schema, name)
}

func (r *FillInRepository) UpsertValue(ctx context.Context, v PlaceholderValue, q ...DBTX) error {
	var valueTyped any
	if v.ValueTyped != nil {
		b, err := json.Marshal(v.ValueTyped)
		if err != nil {
			return err
		}
		valueTyped = b
	}

	exec := DBTX(r.db)
	if len(q) > 0 && q[0] != nil {
		exec = q[0]
	}

	_, err := exec.ExecContext(ctx, fmt.Sprintf(`
		INSERT INTO %s
		    (tenant_id, revision_id, placeholder_id, value_text, value_typed,
		     source, computed_from, resolver_version, inputs_hash, validated_at, created_at, updated_at)
		VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW(), NOW())
		ON CONFLICT (tenant_id, revision_id, placeholder_id) DO UPDATE SET
			value_text       = EXCLUDED.value_text,
			value_typed      = EXCLUDED.value_typed,
			source           = EXCLUDED.source,
			computed_from    = EXCLUDED.computed_from,
			resolver_version = EXCLUDED.resolver_version,
			inputs_hash      = EXCLUDED.inputs_hash,
			validated_at     = NOW(),
			updated_at       = NOW()`, r.table("document_placeholder_values")),
		v.TenantID, v.RevisionID, v.PlaceholderID, v.ValueText, valueTyped,
		v.Source, v.ComputedFrom, v.ResolverVersion, v.InputsHash,
	)
	return err
}

func (r *FillInRepository) ListValues(ctx context.Context, tenantID, revisionID string) ([]PlaceholderValue, error) {
	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT placeholder_id, value_text, value_typed, source, computed_from, resolver_version, inputs_hash
		  FROM %s
		 WHERE tenant_id=$1::uuid AND revision_id=$2::uuid`, r.table("document_placeholder_values")),
		tenantID, revisionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []PlaceholderValue
	for rows.Next() {
		var v PlaceholderValue
		var valueTyped []byte
		if err := rows.Scan(
			&v.PlaceholderID,
			&v.ValueText,
			&valueTyped,
			&v.Source,
			&v.ComputedFrom,
			&v.ResolverVersion,
			&v.InputsHash,
		); err != nil {
			return nil, err
		}
		if len(valueTyped) > 0 {
			v.ValueTyped = map[string]any{}
			if err := json.Unmarshal(valueTyped, &v.ValueTyped); err != nil {
				return nil, err
			}
		}
		v.TenantID = tenantID
		v.RevisionID = revisionID
		out = append(out, v)
	}

	return out, rows.Err()
}

// UpsertAuthorValue writes an author (source='user') fill-in value but ONLY for
// rows whose CURRENT source is author-editable ('user' or 'default'). The
// DO UPDATE ... WHERE guard leaves an existing governed row (computed/dictionary)
// untouched and reports 0 rows affected — the DB enforcement of SP-2 D11. Returns
// rows affected so the service can reject a blocked write.
func (r *FillInRepository) UpsertAuthorValue(ctx context.Context, v PlaceholderValue, q ...DBTX) (int64, error) {
	exec := DBTX(r.db)
	if len(q) > 0 && q[0] != nil {
		exec = q[0]
	}
	res, err := exec.ExecContext(ctx, fmt.Sprintf(`
		INSERT INTO %s
		    (tenant_id, revision_id, placeholder_id, value_text, value_typed,
		     source, computed_from, resolver_version, inputs_hash, validated_at, created_at, updated_at)
		VALUES ($1::uuid, $2::uuid, $3, $4, NULL, 'user', NULL, NULL, NULL, NOW(), NOW(), NOW())
		ON CONFLICT (tenant_id, revision_id, placeholder_id) DO UPDATE SET
			value_text       = EXCLUDED.value_text,
			value_typed      = NULL,
			source           = 'user',
			computed_from    = NULL,
			resolver_version = NULL,
			inputs_hash      = NULL,
			validated_at     = NOW(),
			updated_at       = NOW()
		-- NOTE: the DO UPDATE WHERE must reference the conflict target by its BARE
		-- relation name (Postgres requirement); do NOT schema-qualify it here even
		-- though the INSERT target is qualified — qualifying it is a parse error.
		WHERE document_placeholder_values.source IN ('user','default')`,
		r.table("document_placeholder_values")),
		v.TenantID, v.RevisionID, v.PlaceholderID, v.ValueText,
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// CurrentSource returns the persisted source discriminator for a placeholder
// value row and whether the row exists. Backs the D11 app-layer friendly
// rejection before the guarded write is attempted.
func (r *FillInRepository) CurrentSource(ctx context.Context, tenantID, revisionID, placeholderID string) (string, bool, error) {
	var source string
	err := r.db.QueryRowContext(ctx, fmt.Sprintf(`
		SELECT source FROM %s
		 WHERE tenant_id=$1::uuid AND revision_id=$2::uuid AND placeholder_id=$3`,
		r.table("document_placeholder_values")),
		tenantID, revisionID, placeholderID,
	).Scan(&source)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return source, true, nil
}
