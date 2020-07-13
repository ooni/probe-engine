package webconnectivity_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/ooni/probe-engine/experiment/urlgetter"
	"github.com/ooni/probe-engine/experiment/webconnectivity"
	"github.com/ooni/probe-engine/internal/mockable"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/netx/archival"
)

func TestNewControlRequest(t *testing.T) {
	type args struct {
		target model.MeasurementTarget
		result urlgetter.TestKeys
	}
	tests := []struct {
		name    string
		args    args
		wantOut webconnectivity.ControlRequest
	}{{
		name: "with empty input",
		wantOut: webconnectivity.ControlRequest{
			HTTPRequestHeaders: make(map[string][]string),
			TCPConnect:         []string{},
		},
	}, {
		name: "common case",
		args: args{
			target: "http://www.example.com/",
			result: urlgetter.TestKeys{
				Requests: []archival.RequestEntry{{
					Request: archival.HTTPRequest{
						Headers: map[string]archival.MaybeBinaryValue{
							"User-Agent":      {Value: "antani/1.0"},
							"Accept":          {Value: "*/*"},
							"Accept-Language": {Value: "en_UK"},
							"Host":            {Value: "www.example.com"}, // extra!
						},
					},
				}},
				TCPConnect: []archival.TCPConnectEntry{{
					IP:   "10.0.0.1",
					Port: 80,
				}, {
					IP:   "10.0.0.2",
					Port: 443,
				}, {
					IP:   "10.0.0.1", // duplicate: do we reduce it?
					Port: 80,
				}},
			},
		},
		wantOut: webconnectivity.ControlRequest{
			HTTPRequest: "http://www.example.com/",
			HTTPRequestHeaders: map[string][]string{
				"User-Agent":      {"antani/1.0"},
				"Accept":          {"*/*"},
				"Accept-Language": {"en_UK"},
			},
			TCPConnect: []string{
				"10.0.0.1:80", "10.0.0.2:443",
			},
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOut := webconnectivity.NewControlRequest(tt.args.target, tt.args.result)
			if diff := cmp.Diff(tt.wantOut, gotOut); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestFillASNsEmpty(t *testing.T) {
	dns := new(webconnectivity.ControlDNSResult)
	dns.FillASNs(new(mockable.ExperimentSession))
	if diff := cmp.Diff(dns.ASNs, []int64{}); diff != "" {
		t.Fatal(diff)
	}
}

func TestFillASNsNoDatabase(t *testing.T) {
	dns := new(webconnectivity.ControlDNSResult)
	dns.Addrs = []string{"8.8.8.8", "1.1.1.1"}
	dns.FillASNs(new(mockable.ExperimentSession))
	if diff := cmp.Diff(dns.ASNs, []int64{0, 0}); diff != "" {
		t.Fatal(diff)
	}
}

func TestFillASNsSuccess(t *testing.T) {
	sess := newsession(t, false)
	dns := new(webconnectivity.ControlDNSResult)
	dns.Addrs = []string{"8.8.8.8", "1.1.1.1"}
	dns.FillASNs(sess)
	if diff := cmp.Diff(dns.ASNs, []int64{15169, 13335}); diff != "" {
		t.Fatal(diff)
	}
}
