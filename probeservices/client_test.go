package probeservices_test

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/apex/log"
	"github.com/google/go-cmp/cmp"
	"github.com/ooni/probe-engine/internal/mockable"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/probeservices"
)

func TestNewClientHTTPS(t *testing.T) {
	client, err := probeservices.NewClient(
		&mockable.ExperimentSession{}, model.Service{
			Address: "https://x.org",
			Type:    "https",
		})
	if err != nil {
		t.Fatal(err)
	}
	if client.BaseURL != "https://x.org" {
		t.Fatal("not the URL we expected")
	}
}

func TestNewClientUnsupportedEndpoint(t *testing.T) {
	client, err := probeservices.NewClient(
		&mockable.ExperimentSession{}, model.Service{
			Address: "https://x.org",
			Type:    "onion",
		})
	if !errors.Is(err, probeservices.ErrUnsupportedEndpoint) {
		t.Fatal("not the error we expected")
	}
	if client != nil {
		t.Fatal("expected nil client here")
	}
}

func TestNewClientCloudfrontInvalidURL(t *testing.T) {
	client, err := probeservices.NewClient(
		&mockable.ExperimentSession{}, model.Service{
			Address: "\t\t\t",
			Type:    "cloudfront",
		})
	if err == nil || !strings.HasSuffix(err.Error(), "invalid control character in URL") {
		t.Fatal("not the error we expected")
	}
	if client != nil {
		t.Fatal("expected nil client here")
	}
}

func TestNewClientCloudfrontInvalidURLScheme(t *testing.T) {
	client, err := probeservices.NewClient(
		&mockable.ExperimentSession{}, model.Service{
			Address: "http://x.org",
			Type:    "cloudfront",
		})
	if !errors.Is(err, probeservices.ErrUnsupportedCloudFrontAddress) {
		t.Fatal("not the error we expected")
	}
	if client != nil {
		t.Fatal("expected nil client here")
	}
}

func TestNewClientCloudfrontInvalidURLWithPort(t *testing.T) {
	client, err := probeservices.NewClient(
		&mockable.ExperimentSession{}, model.Service{
			Address: "https://x.org:54321",
			Type:    "cloudfront",
		})
	if !errors.Is(err, probeservices.ErrUnsupportedCloudFrontAddress) {
		t.Fatal("not the error we expected")
	}
	if client != nil {
		t.Fatal("expected nil client here")
	}
}

func TestNewClientCloudfrontInvalidFront(t *testing.T) {
	client, err := probeservices.NewClient(
		&mockable.ExperimentSession{}, model.Service{
			Address: "https://x.org",
			Type:    "cloudfront",
			Front:   "\t\t\t",
		})
	if err == nil || !strings.HasSuffix(err.Error(), `invalid URL escape "%09"`) {
		t.Fatal("not the error we expected")
	}
	if client != nil {
		t.Fatal("expected nil client here")
	}
}

func TestNewClientCloudfrontGood(t *testing.T) {
	client, err := probeservices.NewClient(
		&mockable.ExperimentSession{}, model.Service{
			Address: "https://x.org",
			Type:    "cloudfront",
			Front:   "google.com",
		})
	if err != nil {
		t.Fatal(err)
	}
	if client.BaseURL != "https://google.com" {
		t.Fatal("not the BaseURL we expected")
	}
	if client.Host != "x.org" {
		t.Fatal("not the Host we expected")
	}
}

func TestIntegrationCloudfront(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	client, err := probeservices.NewClient(
		&mockable.ExperimentSession{}, model.Service{
			Address: "https://meek.azureedge.net",
			Type:    "cloudfront",
			Front:   "ajax.aspnetcdn.com",
		})
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest("GET", client.BaseURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Host = client.Host
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatal("unexpected status code")
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "I’m just a happy little web server.\n" {
		t.Fatal("unexpected response body")
	}
}

func TestDefaultProbeServicesWorkAsIntended(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	for _, e := range probeservices.Default() {
		client, err := probeservices.NewClient(&mockable.ExperimentSession{
			MockableHTTPClient: http.DefaultClient,
			MockableLogger:     log.Log,
		}, e)
		if err != nil {
			t.Fatal(err)
		}
		testhelpers, err := client.GetTestHelpers(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(testhelpers) < 1 {
			t.Fatal("no test helpers?!")
		}
	}
}

func TestSortEndpoints(t *testing.T) {
	in := []model.Service{{
		Type:    "onion",
		Address: "httpo://jehhrikjjqrlpufu.onion",
	}, {
		Front:   "dkyhjv0wpi2dk.cloudfront.net",
		Type:    "cloudfront",
		Address: "https://dkyhjv0wpi2dk.cloudfront.net",
	}, {
		Type:    "https",
		Address: "https://ams-ps2.ooni.nu:443",
	}}
	expect := []model.Service{{
		Type:    "https",
		Address: "https://ams-ps2.ooni.nu:443",
	}, {
		Front:   "dkyhjv0wpi2dk.cloudfront.net",
		Type:    "cloudfront",
		Address: "https://dkyhjv0wpi2dk.cloudfront.net",
	}, {
		Type:    "onion",
		Address: "httpo://jehhrikjjqrlpufu.onion",
	}}
	out := probeservices.SortEndpoints(in)
	diff := cmp.Diff(out, expect)
	if diff != "" {
		t.Fatal(diff)
	}
}
