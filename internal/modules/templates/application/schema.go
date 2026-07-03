package application

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"time"

	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	renderdomain "metaldocs/internal/modules/render/domain"
	"metaldocs/internal/modules/templates/domain"
)

// UpdateSchemasCmd carries the replacement metadata and placeholder schemas
// for a draft template version, guarded by optimistic concurrency via
// ExpectedLockVersion.
type UpdateSchemasCmd struct {
	TenantID, ActorUserID, TemplateID string
	VersionNumber                     int
	MetadataSchema                    domain.MetadataSchema
	PlaceholderSchema                 []domain.Placeholder
	ExpectedLockVersion               int
}

// UpdateSchemas validates and replaces a draft version's metadata schema and
// placeholder schema. The version must still be a draft. Placeholders are
// validated via ValidatePlaceholders (IDs, names, resolver keys against the
// known resolver registry, select-type option requirements, and visibility
// dependency cycles). The update is applied via a compare-and-swap against
// ExpectedLockVersion (stale lock is surfaced as a domain error) and an
// AuditSaved event ("kind": "schema") is appended, atomically in one
// transaction.
func (s *Service) UpdateSchemas(ctx context.Context, cmd UpdateSchemasCmd) (*domain.TemplateVersion, error) {
	if _, err := s.GetTemplate(ctx, cmd.TenantID, cmd.TemplateID); err != nil {
		return nil, wrapAppErr("templates update schemas: get template", err)
	}
	version, err := s.GetVersion(ctx, cmd.TenantID, cmd.TemplateID, cmd.VersionNumber)
	if err != nil {
		return nil, wrapAppErr("templates update schemas: get version", err)
	}
	if version.Status != domain.VersionStatusDraft {
		return nil, domain.ErrInvalidStateTransition
	}

	var nativeNames map[string]int
	if s.resolvers != nil {
		nativeNames = s.resolvers.Known()
	}
	if err := ValidatePlaceholders(cmd.PlaceholderSchema, nativeNames); err != nil {
		return nil, err
	}
	metadata, err := domain.NewMetadataSchema(
		cmd.MetadataSchema.DocCodePattern,
		cmd.MetadataSchema.RetentionDays,
		cloneStringSlice(cmd.MetadataSchema.DistributionDefault),
		cloneStringSlice(cmd.MetadataSchema.RequiredMetadata),
	)
	if err != nil {
		return nil, err
	}
	if nativeNames != nil {
		for _, p := range cmd.PlaceholderSchema {
			if p.ResolverKey == nil {
				continue
			}
			if _, ok := nativeNames[*p.ResolverKey]; !ok {
				return nil, fmt.Errorf("placeholder[%s] resolver_key %q: %w", p.ID, *p.ResolverKey, domain.ErrUnknownResolver)
			}
		}
	}
	for _, p := range cmd.PlaceholderSchema {
		if p.Type == domain.PHSelect && len(p.Options) == 0 {
			return nil, fmt.Errorf("select_placeholder_requires_options: %s", p.ID)
		}
		if p.Type != domain.PHSelect && len(p.Options) > 0 {
			return nil, fmt.Errorf("options_allowed_only_for_select: %s", p.ID)
		}
	}

	version.MetadataSchema = metadata
	version.PlaceholderSchema = clonePlaceholders(cmd.PlaceholderSchema)

	audit, err := newAuditEvent(cmd.TenantID, cmd.TemplateID, cmd.ActorUserID, &version.ID, domain.AuditSaved, map[string]any{"kind": "schema"}, s.clock.Now())
	if err != nil {
		return nil, err
	}

	if err := s.runner.Do(ctx, func(tx *sql.Tx) error {
		if err := authz.SeedTxIdentity(ctx, tx, cmd.TenantID, cmd.ActorUserID); err != nil {
			return fmt.Errorf("templates update schemas: setAuthzGUC: %w", err)
		}
		if err := authz.Require(ctx, tx, string(iamdomain.CapTemplateEdit), "tenant"); err != nil {
			return fmt.Errorf("templates update schemas: authz: %w", err)
		}
		if err := s.repo.UpdateVersionSchemaCASTx(ctx, tx, cmd.TenantID, version, cmd.ExpectedLockVersion); err != nil {
			return wrapAppErr("templates update schemas: update version", err)
		}
		if err := s.repo.AppendAuditTx(ctx, tx, audit); err != nil {
			return wrapAppErr("templates update schemas: append audit", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return version, nil
}

var placeholderNameRe = regexp.MustCompile(`^[a-z][a-z0-9_]{0,49}$`)

// computedCatalogSet returns the set of valid computed-placeholder names,
// derived from render/domain.ComputedCatalog() — single source of truth per
// ADR 0050. ComputedCatalog is a pure function returning a fresh static slice,
// so it is called inline at each use-site (no caching needed); adding a token
// there automatically unlocks it here.
func computedCatalogSet() map[string]struct{} {
	tokens := renderdomain.ComputedCatalog()
	set := make(map[string]struct{}, len(tokens))
	for _, t := range tokens {
		set[t.Key] = struct{}{}
	}
	return set
}

// ValidatePlaceholders validates a placeholder schema in full: non-empty
// unique IDs, unique names matching the placeholder-name pattern, dictionary
// placeholders that don't collide with native/computed names and carry no
// resolver, computed placeholders that ARE in the computed catalog (ADR
// 0050) and carry a matching resolver_key, well-formed regex/min-max/date
// constraints, computed placeholders requiring a resolver_key, valid
// visibility operators, and (via DetectVisibilityCycle) an acyclic
// visibility dependency graph. nativeNames is the set of known native/
// computed resolver keys used to reject reserved dictionary names.
func ValidatePlaceholders(phs []domain.Placeholder, nativeNames map[string]int) error {
	seen := make(map[string]struct{}, len(phs))
	seenNames := make(map[string]struct{}, len(phs))
	for i, p := range phs {
		if p.ID == "" {
			return fmt.Errorf("placeholder[%d]: %w", i, domain.ErrPlaceholderIDEmpty)
		}
		if _, exists := seen[p.ID]; exists {
			return fmt.Errorf("duplicate_placeholder_id: %s: %w", p.ID, domain.ErrDuplicatePlaceholderID)
		}
		seen[p.ID] = struct{}{}
		if p.Name != "" {
			if !placeholderNameRe.MatchString(p.Name) {
				return fmt.Errorf("placeholder[%s] name invalid: %w", p.ID, domain.ErrPlaceholderNameInvalid)
			}
			if _, exists := seenNames[p.Name]; exists {
				return fmt.Errorf("duplicate_placeholder_name: %s: %w", p.Name, domain.ErrDuplicatePlaceholderName)
			}
			seenNames[p.Name] = struct{}{}

			if p.Type == domain.PHDictionary {
				// SP-2 D5: a dictionary reference name must NOT be a native/computed
				// token name. Uses nativeNames (caller-supplied from resolver registry),
				// which covers all ComputedCatalog() keys including approval_date.
				if _, isNative := nativeNames[p.Name]; isNative {
					return fmt.Errorf("placeholder[%s] name %q: %w", p.ID, p.Name, domain.ErrPlaceholderReservedName)
				}
				// A dictionary placeholder is a pure reference: no resolver, not computed.
				if p.ResolverKey != nil || p.Computed {
					return fmt.Errorf("placeholder[%s] name %q: %w", p.ID, p.Name, domain.ErrPlaceholderDictionaryInvalid)
				}
			} else {
				// A named placeholder must be in the computed catalog (derived from
				// render/domain.ComputedCatalog(), single source per ADR 0050) and be
				// computed with a matching resolver_key.
				if _, ok := computedCatalogSet()[p.Name]; !ok {
					return fmt.Errorf("placeholder[%s] name %q: %w", p.ID, p.Name, domain.ErrPlaceholderNotInCatalog)
				}
				if p.Type != domain.PHComputed || p.ResolverKey == nil || *p.ResolverKey != p.Name {
					return fmt.Errorf("placeholder[%s] %q: %w", p.ID, p.Name, domain.ErrPlaceholderNotComputed)
				}
			}
		}
		if p.Regex != nil {
			if _, err := regexp.Compile(*p.Regex); err != nil {
				return fmt.Errorf("placeholder[%s] regex: %w", p.ID, domain.ErrInvalidConstraint)
			}
		}
		if p.MinNumber != nil && p.MaxNumber != nil && *p.MinNumber > *p.MaxNumber {
			return fmt.Errorf("placeholder[%s] min_number greater than max_number: %w", p.ID, domain.ErrInvalidConstraint)
		}
		if p.MinDate != nil {
			if _, err := time.Parse("2006-01-02", *p.MinDate); err != nil {
				return fmt.Errorf("placeholder[%s] min_date invalid: %w", p.ID, domain.ErrInvalidConstraint)
			}
		}
		if p.MaxDate != nil {
			if _, err := time.Parse("2006-01-02", *p.MaxDate); err != nil {
				return fmt.Errorf("placeholder[%s] max_date invalid: %w", p.ID, domain.ErrInvalidConstraint)
			}
		}
		if p.MinDate != nil && p.MaxDate != nil && *p.MinDate > *p.MaxDate {
			return fmt.Errorf("placeholder[%s] min_date greater than max_date: %w", p.ID, domain.ErrInvalidConstraint)
		}
		if p.MaxLength != nil && *p.MaxLength <= 0 {
			return fmt.Errorf("placeholder[%s] max_length must be positive: %w", p.ID, domain.ErrInvalidConstraint)
		}
		if p.Computed && (p.ResolverKey == nil || *p.ResolverKey == "") {
			return fmt.Errorf("placeholder[%s] computed requires resolver_key: %w", p.ID, domain.ErrInvalidConstraint)
		}
		if p.VisibleIf != nil && !p.VisibleIf.Op.Valid() {
			return fmt.Errorf("placeholder[%s] visibility op %q: %w", p.ID, p.VisibleIf.Op, domain.ErrInvalidConstraint)
		}
	}
	if err := DetectVisibilityCycle(phs); err != nil {
		return err
	}
	return nil
}
