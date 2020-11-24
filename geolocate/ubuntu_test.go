package geolocate_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/geolocate"
	"github.com/ooni/probe-engine/internal/httpheader"
	"github.com/ooni/probe-engine/model"
)

func TestUbuntuParseError(t *testing.T) {
	ip, err := geolocate.UbuntuIPLookup(
		context.Background(),
		&http.Client{
			Transport: geolocate.FakeTransport{
				Resp: &http.Response{
					StatusCode: 200,
					Body:       ioutil.NopCloser(strings.NewReader("<")),
				},
			},
		},
		log.Log,
		httpheader.UserAgent(),
	)
	if err == nil || !strings.HasPrefix(err.Error(), "XML syntax error") {
		t.Fatalf("not the error we expected: %+v", err)
	}
	if ip != model.DefaultProbeIP {
		t.Fatalf("not the expected IP address: %s", ip)
	}
}
