package orchestra_test

import (
	"net/http"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/internal/kvstore"
	"github.com/ooni/probe-engine/internal/orchestra"
)

func newclient() *orchestra.Client {
	clnt := orchestra.NewClient(
		http.DefaultClient,
		log.Log,
		"miniooni/0.1.0-dev",
		orchestra.NewStateFile(kvstore.NewMemoryKeyValueStore()),
	)
	clnt.BaseURL = "https://ps-test.ooni.io"
	return clnt
}
