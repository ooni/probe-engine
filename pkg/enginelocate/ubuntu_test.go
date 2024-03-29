package enginelocate

import (
	"context"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/pkg/model"
	"github.com/ooni/probe-engine/pkg/netxlite"
)

func TestUbuntuParseError(t *testing.T) {
	netx := &netxlite.Netx{}
	ip, err := ubuntuIPLookup(
		context.Background(),
		&http.Client{Transport: FakeTransport{
			Resp: &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader("<")),
			},
		}},
		log.Log,
		model.HTTPHeaderUserAgent,
		netx.NewStdlibResolver(model.DiscardLogger),
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

	netx := &netxlite.Netx{}
	ip, err := ubuntuIPLookup(
		context.Background(),
		http.DefaultClient,
		log.Log,
		model.HTTPHeaderUserAgent,
		netx.NewStdlibResolver(model.DiscardLogger),
	)
	if err != nil {
		t.Fatal(err)
	}
	if net.ParseIP(ip) == nil {
		t.Fatalf("not an IP address: '%s'", ip)
	}
}
