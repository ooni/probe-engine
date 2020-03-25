package eventsaving_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/ooni/probe-engine/internal/eventsaving"
	"github.com/ooni/probe-engine/internal/measurable"
)

func TestIntegration(t *testing.T) {
	saver := new(eventsaving.Saver)
	ctx := eventsaving.WithSaver(context.Background(), saver)
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
	t.Logf("%+v", saver.ReadEvents())
}
