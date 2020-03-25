package eventsaving_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/internal/runtimex"
	"github.com/ooni/probe-engine/netx/eventsaving"
	"github.com/ooni/probe-engine/netx/logging"
	"github.com/ooni/probe-engine/netx/measurable"
)

func TestIntegration(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	logger := logging.Handler{
		Operations: measurable.Defaults{},
		Logger:     log.Log,
		Prefix:     "<httptest>",
	}
	ctx := measurable.WithOperations(context.Background(), logger)
	saver := perform(ctx, "http://www.google.com")
	t.Logf("%+v", saver.ReadEvents())
	saver = perform(ctx, "http://facebook.com")
	t.Logf("%+v", saver.ReadEvents())
}

func perform(ctx context.Context, url string) *eventsaving.Saver {
	ops := measurable.ContextOperations(ctx)
	if ops == nil {
		ops = measurable.Defaults{}
	}
	saver := &eventsaving.Saver{Operations: ops}
	ctx = measurable.WithOperations(context.Background(), saver)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	runtimex.PanicOnError(err, "http.NewRequestWithContext failed")
	resp, err := measurable.DefaultHTTPClient.Do(req)
	runtimex.PanicOnError(err, "measurable.DefaultHTTPClient.Do failed")
	ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	return saver
}
