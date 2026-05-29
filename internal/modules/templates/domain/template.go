package domain

import (
	"errors"
	"time"
)

type Template struct {
	ID                 string
	TenantID           string
	DocTypeCode        string
	Key                string
	Name               string
	Description        string
	LatestVersion          int
	PublishedVersionID     *string
	PublishedVersionNumber *int
	CreatedBy              string
	SystemOwned        bool
	CreatedAt          time.Time
	ArchivedAt         *time.Time
}

func (t *Template) IsArchived() bool { return t.ArchivedAt != nil }

var (
	ErrNotFound                = errors.New("templates: not_found")
	ErrKeyConflict             = errors.New("templates: key_conflict")
	ErrArchived                = errors.New("templates: archived")
	ErrSystemTemplateImmutable = errors.New("templates: system_template_immutable")
)
