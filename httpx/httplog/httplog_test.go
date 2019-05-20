package httplog_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/httpx/httplog"
	"github.com/ooni/probe-engine/httpx/httptracex"
)

func TestGet(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	ctx := context.Background()
	client := &http.Client{
		Transport: &httptracex.Measurer{
			RoundTripper: http.DefaultTransport,
			Handler: &httplog.RoundTripLogger{
				Logger: log.Log,
			},
		},
	}
	request, err := http.NewRequest("GET", "http://facebook.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	response, err := client.Do(request.WithContext(ctx))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	_, err = ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
}
