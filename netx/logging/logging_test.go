package logging_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/internal/logging"
	"github.com/ooni/probe-engine/internal/measurable"
)

func TestIntegration(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	defer log.SetLevel(log.InfoLevel)
	ctx := logging.WithLogger(context.Background(), log.Log)
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
