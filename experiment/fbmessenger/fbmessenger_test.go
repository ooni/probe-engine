package fbmessenger_test

import (
	"context"
	"io"
	"testing"

	"github.com/apex/log"
	engine "github.com/ooni/probe-engine"
	"github.com/ooni/probe-engine/experiment/fbmessenger"
	"github.com/ooni/probe-engine/experiment/urlgetter"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/netx/archival"
	"github.com/ooni/probe-engine/netx/errorx"
)

func TestNewExperimentMeasurer(t *testing.T) {
	measurer := fbmessenger.NewExperimentMeasurer(fbmessenger.Config{})
	if measurer.ExperimentName() != "facebook_messenger" {
		t.Fatal("unexpected name")
	}
	if measurer.ExperimentVersion() != "0.1.0" {
		t.Fatal("unexpected version")
	}
}

func TestIntegrationSuccess(t *testing.T) {
	measurer := fbmessenger.NewExperimentMeasurer(fbmessenger.Config{})
	ctx := context.Background()
	// we need a real session because we need to ASN database
	sess := newsession(t)
	measurement := new(model.Measurement)
	callbacks := model.NewPrinterCallbacks(log.Log)
	err := measurer.Run(ctx, sess, measurement, callbacks)
	if err != nil {
		t.Fatal(err)
	}
	tk := measurement.TestKeys.(*fbmessenger.TestKeys)
	if *tk.FacebookBAPIDNSConsistent != true {
		t.Fatal("invalid FacebookBAPIDNSConsistent")
	}
	if *tk.FacebookBAPIReachable != true {
		t.Fatal("invalid FacebookBAPIReachable")
	}
	if *tk.FacebookBGraphDNSConsistent != true {
		t.Fatal("invalid FacebookBGraphDNSConsistent")
	}
	if *tk.FacebookBGraphReachable != true {
		t.Fatal("invalid FacebookBGraphReachable")
	}
	if *tk.FacebookEdgeDNSConsistent != true {
		t.Fatal("invalid FacebookEdgeDNSConsistent")
	}
	if *tk.FacebookEdgeReachable != true {
		t.Fatal("invalid FacebookEdgeReachable")
	}
	if *tk.FacebookExternalCDNDNSConsistent != true {
		t.Fatal("invalid FacebookExternalCDNDNSConsistent")
	}
	if *tk.FacebookExternalCDNReachable != true {
		t.Fatal("invalid FacebookExternalCDNReachable")
	}
	if *tk.FacebookScontentCDNDNSConsistent != true {
		t.Fatal("invalid FacebookScontentCDNDNSConsistent")
	}
	if *tk.FacebookScontentCDNReachable != true {
		t.Fatal("invalid FacebookScontentCDNReachable")
	}
	if *tk.FacebookStarDNSConsistent != true {
		t.Fatal("invalid FacebookStarDNSConsistent")
	}
	if *tk.FacebookStarReachable != true {
		t.Fatal("invalid FacebookStarReachable")
	}
	if *tk.FacebookSTUNDNSConsistent != true {
		t.Fatal("invalid FacebookSTUNDNSConsistent")
	}
	if tk.FacebookSTUNReachable != nil {
		t.Fatal("invalid FacebookSTUNReachable")
	}
	if *tk.FacebookDNSBlocking != false {
		t.Fatal("invalid FacebookDNSBlocking")
	}
	if *tk.FacebookTCPBlocking != false {
		t.Fatal("invalid FacebookTCPBlocking")
	}
}

func TestIntegrationWithCancelledContext(t *testing.T) {
	measurer := fbmessenger.NewExperimentMeasurer(fbmessenger.Config{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // so we fail immediately
	sess := newsession(t)
	measurement := new(model.Measurement)
	callbacks := model.NewPrinterCallbacks(log.Log)
	err := measurer.Run(ctx, sess, measurement, callbacks)
	if err != nil {
		t.Fatal(err)
	}
	tk := measurement.TestKeys.(*fbmessenger.TestKeys)
	if *tk.FacebookBAPIDNSConsistent != false {
		t.Fatal("invalid FacebookBAPIDNSConsistent")
	}
	if tk.FacebookBAPIReachable != nil {
		t.Fatal("invalid FacebookBAPIReachable")
	}
	if *tk.FacebookBGraphDNSConsistent != false {
		t.Fatal("invalid FacebookBGraphDNSConsistent")
	}
	if tk.FacebookBGraphReachable != nil {
		t.Fatal("invalid FacebookBGraphReachable")
	}
	if *tk.FacebookEdgeDNSConsistent != false {
		t.Fatal("invalid FacebookEdgeDNSConsistent")
	}
	if tk.FacebookEdgeReachable != nil {
		t.Fatal("invalid FacebookEdgeReachable")
	}
	if *tk.FacebookExternalCDNDNSConsistent != false {
		t.Fatal("invalid FacebookExternalCDNDNSConsistent")
	}
	if tk.FacebookExternalCDNReachable != nil {
		t.Fatal("invalid FacebookExternalCDNReachable")
	}
	if *tk.FacebookScontentCDNDNSConsistent != false {
		t.Fatal("invalid FacebookScontentCDNDNSConsistent")
	}
	if tk.FacebookScontentCDNReachable != nil {
		t.Fatal("invalid FacebookScontentCDNReachable")
	}
	if *tk.FacebookStarDNSConsistent != false {
		t.Fatal("invalid FacebookStarDNSConsistent")
	}
	if tk.FacebookStarReachable != nil {
		t.Fatal("invalid FacebookStarReachable")
	}
	if *tk.FacebookSTUNDNSConsistent != false {
		t.Fatal("invalid FacebookSTUNDNSConsistent")
	}
	if tk.FacebookSTUNReachable != nil {
		t.Fatal("invalid FacebookSTUNReachable")
	}
	if *tk.FacebookDNSBlocking != true {
		t.Fatal("invalid FacebookDNSBlocking")
	}
	// no TCP blocking because we didn't ever reach TCP connect
	if *tk.FacebookTCPBlocking != false {
		t.Fatal("invalid FacebookTCPBlocking")
	}
}

func TestComputeEndpointStatsTCPBlocking(t *testing.T) {
	failure := io.EOF.Error()
	operation := errorx.ConnectOperation
	tk := fbmessenger.TestKeys{}
	tk.Update(urlgetter.MultiOutput{
		Input: urlgetter.MultiInput{Target: fbmessenger.ServiceEdge},
		TestKeys: urlgetter.TestKeys{
			Failure:         &failure,
			FailedOperation: &operation,
			Queries: []archival.DNSQueryEntry{{
				Answers: []archival.DNSAnswerEntry{{
					ASN: fbmessenger.FacebookASN,
				}},
			}},
		},
	})
	if *tk.FacebookEdgeDNSConsistent != true {
		t.Fatal("invalid FacebookEdgeDNSConsistent")
	}
	if *tk.FacebookEdgeReachable != false {
		t.Fatal("invalid FacebookEdgeReachable")
	}
	if tk.FacebookDNSBlocking != nil { // meaning: not determined yet
		t.Fatal("invalid FacebookDNSBlocking")
	}
	if *tk.FacebookTCPBlocking != true {
		t.Fatal("invalid FacebookTCPBlocking")
	}
}

func TestComputeEndpointStatsDNSIsLying(t *testing.T) {
	failure := io.EOF.Error()
	operation := errorx.ConnectOperation
	tk := fbmessenger.TestKeys{}
	tk.Update(urlgetter.MultiOutput{
		Input: urlgetter.MultiInput{Target: fbmessenger.ServiceEdge},
		TestKeys: urlgetter.TestKeys{
			Failure:         &failure,
			FailedOperation: &operation,
			Queries: []archival.DNSQueryEntry{{
				Answers: []archival.DNSAnswerEntry{{
					ASN: 0,
				}},
			}},
		},
	})
	if *tk.FacebookEdgeDNSConsistent != false {
		t.Fatal("invalid FacebookEdgeDNSConsistent")
	}
	if tk.FacebookEdgeReachable != nil {
		t.Fatal("invalid FacebookEdgeReachable")
	}
	if *tk.FacebookDNSBlocking != true {
		t.Fatal("invalid FacebookDNSBlocking")
	}
	if tk.FacebookTCPBlocking != nil { // meaning: not determined yet
		t.Fatal("invalid FacebookTCPBlocking")
	}
}

func newsession(t *testing.T) model.ExperimentSession {
	sess, err := engine.NewSession(engine.SessionConfig{
		AssetsDir: "../../testdata",
		AvailableProbeServices: []model.Service{{
			Address: "https://ams-pg.ooni.org",
			Type:    "https",
		}},
		Logger: log.Log,
		PrivacySettings: model.PrivacySettings{
			IncludeASN:     true,
			IncludeCountry: true,
			IncludeIP:      false,
		},
		SoftwareName:    "ooniprobe-engine",
		SoftwareVersion: "0.0.1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := sess.MaybeLookupLocation(); err != nil {
		t.Fatal(err)
	}
	return sess
}
