package engine

import (
	"context"

	"github.com/ooni/probe-engine/model"
)

func (s *Session) SetAssetsDir(assetsDir string) {
	s.assetsDir = assetsDir
}

func (s *Session) GetAvailableProbeServices() []model.Service {
	return s.getAvailableProbeServices()
}

func (s *Session) AppendAvailableProbeService(svc model.Service) {
	s.availableProbeServices = append(s.availableProbeServices, svc)
}

func (s *Session) MaybeLookupBackendsContext(ctx context.Context) (err error) {
	return s.maybeLookupBackends(ctx)
}

func (s *Session) QueryProbeServicesCount() int64 {
	return s.queryProbeServicesCount.Load()
}
