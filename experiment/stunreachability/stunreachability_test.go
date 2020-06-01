package stunreachability_test

import (
	"context"
	"errors"
	"net"
	"strings"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/experiment/handler"
	"github.com/ooni/probe-engine/experiment/stunreachability"
	"github.com/ooni/probe-engine/internal/mockable"
	"github.com/ooni/probe-engine/model"
	"github.com/pion/stun"
)

func TestMeasurerExperimentNameVersion(t *testing.T) {
	measurer := stunreachability.NewExperimentMeasurer(stunreachability.Config{})
	if measurer.ExperimentName() != "stun_reachability" {
		t.Fatal("unexpected ExperimentName")
	}
	if measurer.ExperimentVersion() != "0.0.1" {
		t.Fatal("unexpected ExperimentVersion")
	}
}

func TestIntegrationRun(t *testing.T) {
	measurer := stunreachability.NewExperimentMeasurer(stunreachability.Config{})
	measurement := new(model.Measurement)
	err := measurer.Run(
		context.Background(),
		&mockable.ExperimentSession{},
		measurement,
		handler.NewPrinterCallbacks(log.Log),
	)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // immediately fail everything
	measurer := stunreachability.NewExperimentMeasurer(stunreachability.Config{})
	measurement := new(model.Measurement)
	err := measurer.Run(
		ctx,
		&mockable.ExperimentSession{},
		measurement,
		handler.NewPrinterCallbacks(log.Log),
	)
	if !strings.HasSuffix(err.Error(), "operation was canceled") {
		t.Fatal("not the error we expected")
	}
}

func TestNewClientFailure(t *testing.T) {
	config := &stunreachability.Config{}
	expected := errors.New("mocked error")
	config.SetNewClient(
		func(conn stun.Connection, options ...stun.ClientOption) (*stun.Client, error) {
			return nil, expected
		})
	measurer := stunreachability.NewExperimentMeasurer(*config)
	measurement := new(model.Measurement)
	err := measurer.Run(
		context.Background(),
		&mockable.ExperimentSession{},
		measurement,
		handler.NewPrinterCallbacks(log.Log),
	)
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
}

func TestStartFailure(t *testing.T) {
	config := &stunreachability.Config{}
	expected := errors.New("mocked error")
	config.SetDialContext(
		func(ctx context.Context, network, address string) (net.Conn, error) {
			conn := &stunreachability.FakeConn{WriteError: expected}
			return conn, nil
		})
	measurer := stunreachability.NewExperimentMeasurer(*config)
	measurement := new(model.Measurement)
	err := measurer.Run(
		context.Background(),
		&mockable.ExperimentSession{},
		measurement,
		handler.NewPrinterCallbacks(log.Log),
	)
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
}

func TestReadFailure(t *testing.T) {
	config := &stunreachability.Config{}
	expected := errors.New("mocked error")
	config.SetDialContext(
		func(ctx context.Context, network, address string) (net.Conn, error) {
			conn := &stunreachability.FakeConn{ReadError: expected}
			return conn, nil
		})
	measurer := stunreachability.NewExperimentMeasurer(*config)
	measurement := new(model.Measurement)
	err := measurer.Run(
		context.Background(),
		&mockable.ExperimentSession{},
		measurement,
		handler.NewPrinterCallbacks(log.Log),
	)
	if !errors.Is(err, stun.ErrTransactionTimeOut) {
		t.Fatal("not the error we expected")
	}
}
