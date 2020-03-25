package bytecounting_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/ooni/probe-engine/internal/bytecounting"
	"github.com/ooni/probe-engine/internal/measurable"
)

func TestIntegration(t *testing.T) {
	counter := bytecounting.NewCounter()
	ctx := bytecounting.WithCounter(context.Background(), counter)
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
	t.Logf("%d %d", counter.BytesRecv.Load(), counter.BytesSent.Load())
}
