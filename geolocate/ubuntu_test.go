package geolocate

import (
	"context"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/internal/httpheader"
	"github.com/ooni/probe-engine/model"
)

func TestUbuntuParseError(t *testing.T) {
	ip, err := ubuntuIPLookup(
		context.Background(),
		&http.Client{Transport: FakeTransport{
			Resp: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader("<")),
			},
		}},
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

func TestIPLookupWorksUsingUbuntu(t *testing.T) {
	if testing.Short() {
		t.Skip("skip test in short mode")
	}
	ip, err := ubuntuIPLookup(
		context.Background(),
		http.DefaultClient,
		log.Log,
		httpheader.UserAgent(),
	)
	if err != nil {
		t.Fatal(err)
	}
	if net.ParseIP(ip) == nil {
		t.Fatalf("not an IP address: '%s'", ip)
	}
}
