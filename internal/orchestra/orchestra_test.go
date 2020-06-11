package orchestra_test

import (
	"net/http"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/atomicx"
	"github.com/ooni/probe-engine/internal/httpx"
	"github.com/ooni/probe-engine/internal/kvstore"
	"github.com/ooni/probe-engine/internal/orchestra"
)

func newclient() *orchestra.Client {
	client := &orchestra.Client{
		Client: httpx.Client{
			BaseURL:    "https://ps-test.ooni.io/",
			HTTPClient: http.DefaultClient,
			Logger:     log.Log,
			UserAgent:  "miniooni/0.1.0",
		},
		LoginCalls:    atomicx.NewInt64(),
		RegisterCalls: atomicx.NewInt64(),
		StateFile:     orchestra.NewStateFile(kvstore.NewMemoryKeyValueStore()),
	}
	return client
}
