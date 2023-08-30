package netemx

import (
	"net/http"

	"github.com/ooni/netem"
	"github.com/ooni/probe-engine/pkg/testingx"
)

// GeoIPHandlerFactoryUbuntu is a [QAEnvHTTPHandlerFactory] for [testingx.GeoIPHandlerUbuntu].
type GeoIPHandlerFactoryUbuntu struct {
	ProbeIP string
}

var _ QAEnvHTTPHandlerFactory = &GeoIPHandlerFactoryUbuntu{}

// NewHandler implements QAEnvHTTPHandlerFactory.
func (f *GeoIPHandlerFactoryUbuntu) NewHandler(unet netem.UnderlyingNetwork) http.Handler {
	return &testingx.GeoIPHandlerUbuntu{ProbeIP: f.ProbeIP}
}
