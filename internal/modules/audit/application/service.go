package application

import (
	"context"
	"errors"
	"strings"

	"metaldocs/internal/modules/audit/domain"
)

var ErrTenantRequired = errors.New("audit: tenant id is required")
var ErrReaderRequired = errors.New("audit: reader is required")

type Service struct {
	reader domain.Reader
}

func NewService(reader domain.Reader) *Service {
	if reader == nil {
		panic(ErrReaderRequired.Error())
	}
	return &Service{reader: reader}
}

func (s *Service) ListEvents(ctx context.Context, query domain.ListEventsQuery) ([]domain.Event, error) {
	if s == nil || s.reader == nil {
		return nil, ErrReaderRequired
	}

	normalized := domain.ListEventsQuery{
		ResourceType: strings.TrimSpace(query.ResourceType),
		ResourceID:   strings.TrimSpace(query.ResourceID),
		TenantID:     strings.TrimSpace(query.TenantID),
		Limit:        query.Limit,
	}
	if normalized.TenantID == "" {
		return nil, ErrTenantRequired
	}
	if normalized.Limit <= 0 {
		normalized.Limit = 50
	}
	if normalized.Limit > 200 {
		normalized.Limit = 200
	}

	return s.reader.ListEvents(ctx, normalized)
}
