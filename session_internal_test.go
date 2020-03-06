package engine

import (
	"context"

	"github.com/ooni/probe-engine/model"
)

func (s *Session) SetAssetsDir(assetsDir string) {
	s.assetsDir = assetsDir
}

func (s *Session) FetchResourcesIdempotent(ctx context.Context) error {
	return s.fetchResourcesIdempotent(ctx)
}

func (s *Session) GetAvailableBouncers() []model.Service {
	return s.getAvailableBouncers()
}

func (s *Session) AppendAvailableBouncer(bouncer model.Service) {
	s.availableBouncers = append(s.availableBouncers, bouncer)
}

func (s *Session) MaybeLookupBackendsContext(ctx context.Context) (err error) {
	return s.maybeLookupBackends(ctx)
}

func (s *Session) MaybeLookupTestHelpersContext(ctx context.Context) error {
	return s.maybeLookupTestHelpers(ctx)
}

func (s *Session) QueryBouncerCount() int64 {
	return s.queryBouncerCount
}
