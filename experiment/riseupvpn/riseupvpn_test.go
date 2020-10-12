package riseupvpn_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/experiment/riseupvpn"
	"github.com/ooni/probe-engine/experiment/urlgetter"
	"github.com/ooni/probe-engine/internal/mockable"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/netx/errorx"
	"github.com/ooni/probe-engine/netx/selfcensor"
)

func TestNewExperimentMeasurer(t *testing.T) {
	measurer := riseupvpn.NewExperimentMeasurer(riseupvpn.Config{})
	if measurer.ExperimentName() != "riseupvpn" {
		t.Fatal("unexpected name")
	}
	if measurer.ExperimentVersion() != "0.0.2" {
		t.Fatal("unexpected version")
	}
}

func TestIntegration(t *testing.T) {
	measurer := riseupvpn.NewExperimentMeasurer(riseupvpn.Config{})
	measurement := new(model.Measurement)
	err := measurer.Run(
		context.Background(),
		&mockable.Session{
			MockableLogger: log.Log,
		},
		measurement,
		model.NewPrinterCallbacks(log.Log),
	)
	if err != nil {
		t.Fatal(err)
	}
	tk := measurement.TestKeys.(*riseupvpn.TestKeys)
	if tk.Agent != "" {
		t.Fatal("unexpected Agent: " + tk.Agent)
	}
	if tk.FailedOperation != nil {
		t.Fatal("unexpected FailedOperation")
	}
	if tk.Failure != nil {
		t.Fatal("unexpected Failure")
	}
	if len(tk.NetworkEvents) <= 0 {
		t.Fatal("no NetworkEvents?!")
	}
	if len(tk.Queries) <= 0 {
		t.Fatal("no Queries?!")
	}
	if len(tk.Requests) <= 0 {
		t.Fatal("no Requests?!")
	}
	if len(tk.TCPConnect) <= 0 {
		t.Fatal("no TCPConnect?!")
	}
	if len(tk.TLSHandshakes) <= 0 {
		t.Fatal("no TLSHandshakes?!")
	}
	if tk.RiseupVPNApiFailure != nil {
		t.Fatal("unexpected RiseupVPNApiFailure")
	}
	if tk.RiseupVPNApiStatus != "ok" {
		t.Fatal("unexpected RiseupvpnStatus")
	}
	if tk.RiseupVPNCACertStatus != true {
		t.Fatal("unexpected RiseupvPNCaCertStatus")
	}
	if tk.RiseupVPNFailingGateways != nil {
		t.Fatal("unexpected RiseupVPNFailingGateways value")
	}
}

// TestUpdateWithMixedResults tests if one operation failed
// RiseupVPNApiStatus is considered as blocked
func TestUpdateWithMixedResults(t *testing.T) {
	tk := riseupvpn.NewTestKeys()
	tk.Update(urlgetter.MultiOutput{
		Input: urlgetter.MultiInput{
			Config: urlgetter.Config{Method: "GET"},
			Target: "https://api.black.riseup.net:443/3/config/eip-service.json",
		},
		TestKeys: urlgetter.TestKeys{
			HTTPResponseStatus: 200,
		},
	})
	tk.Update(urlgetter.MultiOutput{
		Input: urlgetter.MultiInput{
			Config: urlgetter.Config{Method: "GET"},
			Target: "https://riseup.net/provider.json",
		},
		TestKeys: urlgetter.TestKeys{
			FailedOperation: (func() *string {
				s := errorx.HTTPRoundTripOperation
				return &s
			})(),
			Failure: (func() *string {
				s := errorx.FailureEOFError
				return &s
			})(),
		},
	})
	tk.Update(urlgetter.MultiOutput{
		Input: urlgetter.MultiInput{
			Config: urlgetter.Config{Method: "GET"},
			Target: "https://api.black.riseup.net:9001/json",
		},
		TestKeys: urlgetter.TestKeys{
			HTTPResponseStatus: 200,
		},
	})
	if tk.RiseupVPNApiStatus != "blocked" {
		t.Fatal("RiseupVPNApiStatus should be blocked")
	}
	if *tk.RiseupVPNApiFailure != errorx.FailureEOFError {
		t.Fatal("invalid RiseupVPNApiFailure")
	}
}

func TestIntegrationFailureCaCertFetch(t *testing.T) {
	measurer := riseupvpn.NewExperimentMeasurer(riseupvpn.Config{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	sess := &mockable.Session{MockableLogger: log.Log}
	measurement := new(model.Measurement)
	callbacks := model.NewPrinterCallbacks(log.Log)
	err := measurer.Run(ctx, sess, measurement, callbacks)
	if err != nil {
		t.Fatal(err)
	}
	tk := measurement.TestKeys.(*riseupvpn.TestKeys)
	if tk.RiseupVPNCACertStatus != false {
		t.Fatal("invalid RiseupVPNCACertStatus ")
	}
	if tk.RiseupVPNApiStatus != "blocked" {
		t.Fatal("invalid RiseupVPNApiStatus")
	}

	if tk.RiseupVPNApiFailure != nil {
		t.Fatal("RiseupVPNApiFailure should be null")
	}
	if len(tk.Requests) > 1 {
		t.Fatal("Unexpected requests")
	}

}

func TestIntegrationFailureEipServiceBlocked(t *testing.T) {
	measurer := riseupvpn.NewExperimentMeasurer(riseupvpn.Config{})
	ctx, cancel := context.WithCancel(context.Background())
	selfcensor.Enable(`{"PoisonSystemDNS":{"api.black.riseup.net":["NXDOMAIN"]}}`)

	sess := &mockable.Session{MockableLogger: log.Log}
	measurement := new(model.Measurement)
	callbacks := model.NewPrinterCallbacks(log.Log)
	err := measurer.Run(ctx, sess, measurement, callbacks)
	if err != nil {
		t.Fatal(err)
	}
	tk := measurement.TestKeys.(*riseupvpn.TestKeys)
	if tk.RiseupVPNCACertStatus != true {
		t.Fatal("invalid RiseupVPNCACertStatus ")
	}

	for _, entry := range tk.Requests {
		if entry.Request.URL == "https://api.black.riseup.net:443/3/config/eip-service.json" {
			if entry.Failure == nil {
				t.Fatal("Failure for " + entry.Request.URL + " should not be null")
			}
		}
	}

	if tk.RiseupVPNApiStatus != "blocked" {
		t.Fatal("invalid RiseupVPNApiStatus")
	}

	if tk.RiseupVPNApiFailure == nil {
		t.Fatal("RiseupVPNApiFailure should not be null")
	}

	cancel()
}

func TestIntegrationFailureProviderUrlBlocked(t *testing.T) {
	measurer := riseupvpn.NewExperimentMeasurer(riseupvpn.Config{})
	ctx, cancel := context.WithCancel(context.Background())
	selfcensor.Enable(`{"BlockedEndpoints":{"198.252.153.70:443":"REJECT"}}`)

	sess := &mockable.Session{MockableLogger: log.Log}
	measurement := new(model.Measurement)
	callbacks := model.NewPrinterCallbacks(log.Log)
	err := measurer.Run(ctx, sess, measurement, callbacks)
	if err != nil {
		t.Fatal(err)
	}
	tk := measurement.TestKeys.(*riseupvpn.TestKeys)

	for _, entry := range tk.Requests {
		if entry.Request.URL == "https://riseup.net/provider.json" {
			if entry.Failure == nil {
				t.Fatal("Failure for " + entry.Request.URL + " should not be null")
			}
		}
	}

	if tk.RiseupVPNCACertStatus != true {
		t.Fatal("invalid RiseupVPNCACertStatus ")
	}
	if tk.RiseupVPNApiStatus != "blocked" {
		t.Fatal("invalid RiseupVPNApiStatus")
	}

	if tk.RiseupVPNApiFailure == nil {
		t.Fatal("RiseupVPNApiFailure should not be null")
	}
	cancel()
}

func TestIntegrationFailureGeoIpServiceBlocked(t *testing.T) {
	measurer := riseupvpn.NewExperimentMeasurer(riseupvpn.Config{})
	ctx, cancel := context.WithCancel(context.Background())
	selfcensor.Enable(`{"BlockedEndpoints":{"198.252.153.107:9001":"REJECT"}}`)

	sess := &mockable.Session{MockableLogger: log.Log}
	measurement := new(model.Measurement)
	callbacks := model.NewPrinterCallbacks(log.Log)
	err := measurer.Run(ctx, sess, measurement, callbacks)
	if err != nil {
		t.Fatal(err)
	}
	tk := measurement.TestKeys.(*riseupvpn.TestKeys)
	if tk.RiseupVPNCACertStatus != true {
		t.Fatal("invalid RiseupVPNCACertStatus ")
	}

	for _, entry := range tk.Requests {
		if entry.Request.URL == "https://api.black.riseup.net:9001/json" {
			if entry.Failure == nil {
				t.Fatal("Failure for " + entry.Request.URL + " should not be null")
			}
		}
	}

	if tk.RiseupVPNApiStatus != "blocked" {
		t.Fatal("invalid RiseupVPNApiStatus")
	}

	if tk.RiseupVPNApiFailure == nil {
		t.Fatal("RiseupVPNApiFailure should not be null")
	}

	cancel()
}

func TestIntegrationFailureOpenvpnGateway(t *testing.T) {
	// - fetch client cert and add to certpool
	caFetchClient := &http.Client{
		Timeout: time.Second * 30,
	}

	caCertResponse, err := caFetchClient.Get("https://black.riseup.net/ca.crt")
	if err != nil {
		t.SkipNow()
	}
	defer caCertResponse.Body.Close()

	var bodyString string
	if caCertResponse.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(caCertResponse.Body)
		if err != nil {
			t.SkipNow()
		}
		bodyString = string(bodyBytes)
	}

	certs := x509.NewCertPool()
	certs.AppendCertsFromPEM([]byte(bodyString))

	// - fetch and parse eip-service.json
	client := &http.Client{
		Timeout: time.Second * 30,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: certs,
			},
		},
	}

	eipResponse, err := client.Get("https://api.black.riseup.net/3/config/eip-service.json")
	if err != nil {
		t.SkipNow()
	}
	defer eipResponse.Body.Close()

	if eipResponse.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(eipResponse.Body)
		if err != nil {
			return
		}
		bodyString = string(bodyBytes)
	}

	eipService, err := riseupvpn.DecodeEIP3(bodyString)

	// - self censor random gateway
	gateways := eipService.Gateways
	if gateways == nil || len(gateways) == 0 {
		t.SkipNow()
	}
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	min := 0
	max := len(gateways) - 1
	randomIndex := rnd.Intn(max-min+1) + min

	IP := gateways[randomIndex].IPAddress
	port := gateways[randomIndex].Capabilities.Transport[0].Ports[0]
	selfcensor.Enable(`{"BlockedEndpoints":{"` + IP + `:` + port + `":"REJECT"}}`)

	// - run measurement
	measurer := riseupvpn.NewExperimentMeasurer(riseupvpn.Config{})
	ctx, cancel := context.WithCancel(context.Background())

	sess := &mockable.Session{MockableLogger: log.Log}
	measurement := new(model.Measurement)
	callbacks := model.NewPrinterCallbacks(log.Log)
	err = measurer.Run(ctx, sess, measurement, callbacks)
	if err != nil {
		t.Fatal(err)
	}
	tk := measurement.TestKeys.(*riseupvpn.TestKeys)
	if tk.RiseupVPNCACertStatus != true {
		t.Fatal("invalid RiseupVPNCACertStatus ")
	}

	if tk.RiseupVPNFailingGateways == nil || len(tk.RiseupVPNFailingGateways) != 1 {
		t.Fatal("unexpected amount of failing gateways")
	}

	entry := tk.RiseupVPNFailingGateways[0]
	if entry.IP != IP || fmt.Sprint(entry.Port) != port {
		t.Fatal("unexpected failed gateway configuration")
	}

	if tk.RiseupVPNApiStatus == "blocked" {
		t.Fatal("invalid RiseupVPNApiStatus")
	}

	if tk.RiseupVPNApiFailure != nil {
		t.Fatal("RiseupVPNApiFailure should be null")
	}

	cancel()
}

func TestIntegrationFailureObfs4Gateway(t *testing.T) {
	// - fetch client cert and add to certpool
	caFetchClient := &http.Client{
		Timeout: time.Second * 30,
	}

	caCertResponse, err := caFetchClient.Get("https://black.riseup.net/ca.crt")
	if err != nil {
		t.SkipNow()
	}
	defer caCertResponse.Body.Close()

	var bodyString string
	if caCertResponse.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(caCertResponse.Body)
		if err != nil {
			t.SkipNow()
		}
		bodyString = string(bodyBytes)
	}

	certs := x509.NewCertPool()
	certs.AppendCertsFromPEM([]byte(bodyString))

	// - fetch and parse eip-service.json
	client := &http.Client{
		Timeout: time.Second * 30,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: certs,
			},
		},
	}

	eipResponse, err := client.Get("https://api.black.riseup.net/3/config/eip-service.json")
	if err != nil {
		t.SkipNow()
	}
	defer eipResponse.Body.Close()

	if eipResponse.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(eipResponse.Body)
		if err != nil {
			return
		}
		bodyString = string(bodyBytes)
	}

	eipService, err := riseupvpn.DecodeEIP3(bodyString)

	// - self censor random gateway
	gateways := eipService.Gateways
	if gateways == nil || len(gateways) == 0 {
		t.SkipNow()
	}

	var selfcensoredGateways []string
	for _, gateway := range gateways {
		for _, transport := range gateway.Capabilities.Transport {
			if transport.Type == "obfs4" {
				selfcensoredGateways = append(selfcensoredGateways, `{"BlockedEndpoints":{"`+gateway.IPAddress+`:`+transport.Ports[0]+`":"REJECT"}}`)
			}
		}
	}

	if len(selfcensoredGateways) == 0 {
		t.SkipNow()
	}

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	min := 0
	max := len(selfcensoredGateways) - 1
	randomIndex := rnd.Intn(max-min+1) + min

	selfcensor.Enable(selfcensoredGateways[randomIndex])

	// - run measurement
	measurer := riseupvpn.NewExperimentMeasurer(riseupvpn.Config{})
	ctx, cancel := context.WithCancel(context.Background())

	sess := &mockable.Session{MockableLogger: log.Log}
	measurement := new(model.Measurement)
	callbacks := model.NewPrinterCallbacks(log.Log)
	err = measurer.Run(ctx, sess, measurement, callbacks)
	if err != nil {
		t.Fatal(err)
	}
	tk := measurement.TestKeys.(*riseupvpn.TestKeys)
	if tk.RiseupVPNCACertStatus != true {
		t.Fatal("invalid RiseupVPNCACertStatus ")
	}

	if tk.RiseupVPNFailingGateways == nil || len(tk.RiseupVPNFailingGateways) != 1 {
		t.Fatal("unexpected amount of failing gateways")
	}

	entry := tk.RiseupVPNFailingGateways[0]
	if !strings.Contains(selfcensoredGateways[randomIndex], entry.IP) || !strings.Contains(selfcensoredGateways[randomIndex], strconv.Itoa(entry.Port)) {
		t.Fatal("unexpected failed gateway configuration")
	}

	if tk.RiseupVPNApiStatus == "blocked" {
		t.Fatal("invalid RiseupVPNApiStatus")
	}

	if tk.RiseupVPNApiFailure != nil {
		t.Fatal("RiseupVPNApiFailure should be null")
	}

	cancel()
}
