// Package riseupvpn contains the RiseupVPN network experiment.
//
// See https://github.com/ooni/spec/blob/master/nettests/ts-026-riseupvpn.md
package riseupvpn

import (
	"context"
	"encoding/json"
	"time"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/experiment/urlgetter"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/netx"
	"github.com/ooni/probe-engine/netx/archival"
)

const (
	testName      = "riseupvpn"
	testVersion   = "0.0.2"
	eipServiceURL = "https://api.black.riseup.net:443/3/config/eip-service.json"
	providerURL   = "https://riseup.net/provider.json"
	geoServiceURL = "https://api.black.riseup.net:9001/json"
	tcpConnect    = "tcpconnect://"
)

// EipService main json object of eip-service.json
type EipService struct {
	Gateways []GatewayV3
}

// GatewayV3 json obj Version 3
type GatewayV3 struct {
	Capabilities struct {
		Transport []TransportV3
	}
	Host      string
	IPAddress string `json:"ip_address"`
}

// TransportV3 json obj Version 3
type TransportV3 struct {
	Type      string
	Protocols []string
	Ports     []string
	Options   map[string]string
}

// GatewayConnection describes the connection to a riseupvpn gateway
type GatewayConnection struct {
	IP            string `json:"ip"`
	Port          int    `json:"port"`
	TransportType string `json:"transport_type"`
}

// Config contains the riseupvpn experiment config.
type Config struct {
	urlgetter.Config
}

// TestKeys contains riseupvpn test keys.
type TestKeys struct {
	urlgetter.TestKeys
	ApiFailure      *string             `json:"api_failure"`
	ApiStatus       string              `json:"api_status"`
	CACertStatus    bool                `json:"ca_cert_status"`
	FailingGateways []GatewayConnection `json:"failing_gateways"`
}

// NewTestKeys creates new riseupvpn TestKeys.
func NewTestKeys() *TestKeys {
	return &TestKeys{
		ApiFailure:      nil,
		ApiStatus:       "ok",
		CACertStatus:    true,
		FailingGateways: nil,
	}
}

// UpdateProviderAPITestKeys updates the TestKeys using the given MultiOutput result.
func (tk *TestKeys) UpdateProviderAPITestKeys(v urlgetter.MultiOutput) {
	tk.NetworkEvents = append(tk.NetworkEvents, v.TestKeys.NetworkEvents...)
	tk.Queries = append(tk.Queries, v.TestKeys.Queries...)
	tk.Requests = append(tk.Requests, v.TestKeys.Requests...)
	tk.TCPConnect = append(tk.TCPConnect, v.TestKeys.TCPConnect...)
	tk.TLSHandshakes = append(tk.TLSHandshakes, v.TestKeys.TLSHandshakes...)
	if tk.ApiStatus != "ok" {
		return // we already flipped the state
	}
	if v.TestKeys.Failure != nil {
		tk.ApiStatus = "blocked"
		tk.ApiFailure = v.TestKeys.Failure
		return
	}
}

// AddGatewayConnectTestKeys updates the TestKeys using the given MultiOutput result of gateway connectivity testing.
func (tk *TestKeys) AddGatewayConnectTestKeys(v urlgetter.MultiOutput, transportType string) {
	tk.NetworkEvents = append(tk.NetworkEvents, v.TestKeys.NetworkEvents...)
	tk.TCPConnect = append(tk.TCPConnect, v.TestKeys.TCPConnect...)
	for _, tcpConnect := range v.TestKeys.TCPConnect {
		if !tcpConnect.Status.Success {
			gatewayConnection := newGatewayConnection(tcpConnect, transportType)
			tk.FailingGateways = append(tk.FailingGateways, *gatewayConnection)
		}
	}
	return
}

func newGatewayConnection(tcpConnect archival.TCPConnectEntry, transportType string) *GatewayConnection {
	return &GatewayConnection{
		IP:            tcpConnect.IP,
		Port:          tcpConnect.Port,
		TransportType: transportType,
	}
}

// AddCACertFetchTestKeys Adding generic ctx.Get() testKeys to riseupvpn specific test keys
func (tk *TestKeys) AddCACertFetchTestKeys(testKeys urlgetter.TestKeys) {
	tk.NetworkEvents = append(tk.NetworkEvents, testKeys.NetworkEvents...)
	tk.Queries = append(tk.Queries, testKeys.Queries...)
	tk.Requests = append(tk.Requests, testKeys.Requests...)
	tk.TCPConnect = append(tk.TCPConnect, testKeys.TCPConnect...)
	tk.TLSHandshakes = append(tk.TLSHandshakes, testKeys.TLSHandshakes...)
	if testKeys.Failure != nil {
		tk.ApiStatus = "blocked"
		tk.ApiFailure = tk.Failure
		tk.CACertStatus = false
	}
}

// Measurer performs the measurement
type Measurer struct {
	// Config contains the experiment settings. If empty we
	// will be using default settings.
	Config Config

	// Getter is an optional getter to be used for testing.
	Getter urlgetter.MultiGetter
}

// ExperimentName implements ExperimentMeasurer.ExperimentName
func (m Measurer) ExperimentName() string {
	return testName
}

// ExperimentVersion implements ExperimentMeasurer.ExperimentVersion
func (m Measurer) ExperimentVersion() string {
	return testVersion
}

// Run implements ExperimentMeasurer.Run
func (m Measurer) Run(ctx context.Context, sess model.ExperimentSession,
	measurement *model.Measurement, callbacks model.ExperimentCallbacks) error {
	ctx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()
	testkeys := NewTestKeys()
	measurement.TestKeys = testkeys
	urlgetter.RegisterExtensions(measurement)

	caTarget := "https://black.riseup.net/ca.crt"
	caGetter := urlgetter.Getter{
		Config:  m.Config.Config,
		Session: sess,
		Target:  caTarget,
	}
	log.Info("Getting CA cerificate; please be patient...")
	tk, err := caGetter.Get(ctx)

	if err != nil {
		log.Error("Getting CA cerificate failed. Aborting test.")
		testkeys.AddCACertFetchTestKeys(tk)
		measurement.TestKeys = testkeys
		return nil
	}

	if ok := netx.CertPool.AppendCertsFromPEM([]byte(tk.HTTPResponseBody)); !ok {
		testkeys.CACertStatus = false
		testkeys.ApiStatus = "blocked"
		errorValue := "invalid_ca"
		testkeys.ApiFailure = &errorValue
		return nil
	}

	inputs := []urlgetter.MultiInput{

		// Here we need to provide the method explicitly. See
		// https://github.com/ooni/probe-engine/issues/827.
		{Target: providerURL, Config: urlgetter.Config{
			Method:          "GET",
			FailOnHTTPError: true,
		}},
		{Target: eipServiceURL, Config: urlgetter.Config{
			Method:          "GET",
			FailOnHTTPError: true,
		}},
		{Target: geoServiceURL, Config: urlgetter.Config{
			Method:          "GET",
			FailOnHTTPError: true,
		}},
	}
	multi := urlgetter.Multi{Begin: time.Now(), Getter: m.Getter, Session: sess}

	for entry := range multi.Collect(ctx, inputs, "riseupvpn", callbacks) {
		testkeys.UpdateProviderAPITestKeys(entry)
	}

	// test gateways now
	gateways := parseGateways(testkeys)
	openvpnEndpoints := generateMultiInputs(gateways, "openvpn")
	obfs4Endpoints := generateMultiInputs(gateways, "obfs4")

	// measure openvpn in parallel
	multi = urlgetter.Multi{Begin: measurement.MeasurementStartTimeSaved, Getter: m.Getter, Session: sess}
	for entry := range multi.Collect(ctx, openvpnEndpoints, "riseupvpn", callbacks) {
		testkeys.AddGatewayConnectTestKeys(entry, "openvpn")
	}

	// measure obfs4 in parallel
	multi = urlgetter.Multi{Begin: measurement.MeasurementStartTimeSaved, Getter: m.Getter, Session: sess}
	for entry := range multi.Collect(ctx, obfs4Endpoints, "riseupvpn", callbacks) {
		testkeys.AddGatewayConnectTestKeys(entry, "obfs4")
	}

	return nil
}

func generateMultiInputs(gateways []GatewayV3, transportType string) []urlgetter.MultiInput {
	var gatewayInputs []urlgetter.MultiInput
	for _, gateway := range gateways {
		for _, transport := range gateway.Capabilities.Transport {
			if transport.Type != transportType {
				continue
			}
			supportsTCP := false
			for _, protocol := range transport.Protocols {
				if protocol == "tcp" {
					supportsTCP = true
				}
			}
			if !supportsTCP {
				continue
			}
			for _, port := range transport.Ports {
				tcpConnection := tcpConnect + gateway.IPAddress + ":" + port
				gatewayInputs = append(gatewayInputs, urlgetter.MultiInput{Target: tcpConnection})
			}
		}
	}
	return gatewayInputs
}

func parseGateways(testKeys *TestKeys) []GatewayV3 {
	for _, requestEntry := range testKeys.Requests {
		if requestEntry.Request.URL == eipServiceURL && requestEntry.Failure == nil {
			eipService, err := DecodeEIP3(requestEntry.Response.Body.Value)
			if err == nil {
				return eipService.Gateways
			}
		}
	}

	return nil
}

// DecodeEIP3 decodes eip-service.json version 3
func DecodeEIP3(body string) (*EipService, error) {
	var eip EipService
	err := json.Unmarshal([]byte(body), &eip)
	if err != nil {
		return nil, err
	}
	return &eip, nil
}

// NewExperimentMeasurer creates a new ExperimentMeasurer.
func NewExperimentMeasurer(config Config) model.ExperimentMeasurer {
	return Measurer{Config: config}
}
