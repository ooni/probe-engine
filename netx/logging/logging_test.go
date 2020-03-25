package logging_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/netx/logging"
	"github.com/ooni/probe-engine/netx/measurable"
)

func TestIntegration(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	ops := logging.Handler{
		Operations: measurable.Defaults{},
		Logger:     log.Log,
		Prefix:     "<test #1>",
	}
	ctx := measurable.WithOperations(context.Background(), ops)
	req, err := http.NewRequestWithContext(ctx, "GET", "http://facebook.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := measurable.DefaultHTTPClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	ioutil.ReadAll(resp.Body)
	resp.Body.Close()
}
