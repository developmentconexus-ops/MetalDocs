package application

import "database/sql"

type Service struct {
	repo      Repository
	presign   Presigner
	clock     Clock
	uuid      UUIDGen
	resolvers ResolverRegistryReader
	db        *sql.DB
}

func New(repo Repository, presign Presigner, clock Clock, uuid UUIDGen, resolvers ...ResolverRegistryReader) *Service {
	var registry ResolverRegistryReader
	if len(resolvers) > 0 {
		registry = resolvers[0]
	}
	return &Service{repo: repo, presign: presign, clock: clock, uuid: uuid, resolvers: registry}
}

func (s *Service) WithDB(db *sql.DB) *Service {
	s.db = db
	return s
}
