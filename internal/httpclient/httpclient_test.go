package httpclient_test

import (
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/internal/httpclient"
)

func TestIntegration(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	client := httpclient.New(log.Log)
	resp, err := client.Get("http://facebook.com")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
}
