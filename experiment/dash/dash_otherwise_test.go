// +build nomk

package dash

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/apex/log"
	"github.com/montanaflynn/stats"
	neubotModel "github.com/neubot/dash/model"
	"github.com/ooni/probe-engine/experiment/handler"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/session"
)

func TestUnitNewExperimentMeasurer(t *testing.T) {
	measurer := NewExperimentMeasurer(Config{})
	if measurer.ExperimentName() != "dash" {
		t.Fatal("unexpected name")
	}
	if measurer.ExperimentVersion() != "0.8.0" {
		t.Fatal("unexpected version")
	}
}

func TestUnitMeasureWithCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cause failure
	m := &measurer{}
	err := m.Run(
		ctx,
		&session.Session{
			Logger: log.Log,
		},
		&model.Measurement{},
		handler.NewPrinterCallbacks(log.Log),
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatal("unexpected error value")
	}
}

func TestUnitRunnerLoopClientStartDownloadError(t *testing.T) {
	expect := errors.New("mocked error")
	c := &mockableClient{StartDownloadError: expect}
	runner := newRunner(
		log.Log, c, handler.NewPrinterCallbacks(log.Log), json.Marshal,
	)
	err := runner.loop(context.Background())
	if !errors.Is(err, expect) {
		t.Fatal("not the expected error")
	}
}

func TestUnitRunnerLoopClientJSONMarshalErrorInLoop(t *testing.T) {
	expect := errors.New("mocked error")
	c := &mockableClient{
		ClientResults: make([]neubotModel.ClientResults, 10),
	}
	runner := newRunner(
		log.Log, c, handler.NewPrinterCallbacks(log.Log),
		func(v interface{}) ([]byte, error) {
			return nil, expect
		},
	)
	err := runner.loop(context.Background())
	if !errors.Is(err, expect) {
		t.Fatal("not the expected error")
	}
}

func TestUnitRunnerLoopClientError(t *testing.T) {
	expect := errors.New("mocked error")
	c := &mockableClient{ErrorResult: expect}
	runner := newRunner(
		log.Log, c, handler.NewPrinterCallbacks(log.Log), json.Marshal,
	)
	err := runner.loop(context.Background())
	if !errors.Is(err, expect) {
		t.Fatal("not the expected error")
	}
}

func TestUnitRunnerLoopJSONMarshalErrorAfterLoop(t *testing.T) {
	expect := errors.New("mocked error")
	c := &mockableClient{}
	runner := newRunner(
		log.Log, c, handler.NewPrinterCallbacks(log.Log),
		func(v interface{}) ([]byte, error) {
			return nil, expect
		},
	)
	err := runner.loop(context.Background())
	if !errors.Is(err, expect) {
		t.Fatal("not the expected error")
	}
}

func TestUnitRunnerLoopGood(t *testing.T) {
	c := &mockableClient{
		ClientResults: make([]neubotModel.ClientResults, 10),
	}
	runner := newRunner(
		log.Log, c, handler.NewPrinterCallbacks(log.Log), json.Marshal,
	)
	err := runner.loop(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}

type mockableClient struct {
	StartDownloadError error
	ErrorResult        error
	ClientResults      []neubotModel.ClientResults
}

func (c *mockableClient) StartDownload(
	ctx context.Context,
) (<-chan neubotModel.ClientResults, error) {
	ch := make(chan neubotModel.ClientResults)
	go func() {
		defer close(ch)
		for _, cr := range c.ClientResults {
			ch <- cr
		}
	}()
	return ch, c.StartDownloadError
}

func (c *mockableClient) Error() error {
	return c.ErrorResult
}

func (c *mockableClient) ServerResults() []neubotModel.ServerResults {
	return nil
}

func TestUnitTestKeysAnalyzeWithNoData(t *testing.T) {
	tk := &TestKeys{}
	err := tk.analyze()
	if !errors.Is(err, stats.EmptyInputErr) {
		t.Fatal("expected an error here")
	}
}

func TestUnitTestKeysAnalyzeMedian(t *testing.T) {
	tk := &TestKeys{
		ReceiverData: []neubotModel.ClientResults{
			neubotModel.ClientResults{
				Rate: 1,
			},
			neubotModel.ClientResults{
				Rate: 2,
			},
			neubotModel.ClientResults{
				Rate: 3,
			},
		},
	}
	err := tk.analyze()
	if err != nil {
		t.Fatal(err)
	}
	if tk.Simple.MedianBitrate != 2 {
		t.Fatal("unexpected median value")
	}
}

func TestUnitTestKeysAnalyzeMinPlayoutDelay(t *testing.T) {
	tk := &TestKeys{
		ReceiverData: []neubotModel.ClientResults{
			neubotModel.ClientResults{
				ElapsedTarget: 2,
				Elapsed:       1.4,
			},
			neubotModel.ClientResults{
				ElapsedTarget: 2,
				Elapsed:       3.0,
			},
			neubotModel.ClientResults{
				ElapsedTarget: 2,
				Elapsed:       1.8,
			},
		},
	}
	err := tk.analyze()
	if err != nil {
		t.Fatal(err)
	}
	if tk.Simple.MinPlayoutDelay < 0.99 || tk.Simple.MinPlayoutDelay > 1.01 {
		t.Fatal("unexpected min-playout-delay value")
	}
}

func TestUnitTestKeysPrintSummaryWithNoData(t *testing.T) {
	// The main concern here is that we don't crash when we're
	// provided empty input from the caller
	tk := &TestKeys{}
	tk.printSummary(log.Log)
}

func TestUnitRunnerDoWithLoopSuccess(t *testing.T) {
	c := &mockableClient{
		ClientResults: make([]neubotModel.ClientResults, 10),
	}
	runner := newRunner(
		log.Log, c, handler.NewPrinterCallbacks(log.Log), json.Marshal,
	)
	err := runner.do(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnitRunnerDoWithNoData(t *testing.T) {
	c := &mockableClient{}
	runner := newRunner(
		log.Log, c, handler.NewPrinterCallbacks(log.Log), json.Marshal,
	)
	err := runner.do(context.Background())
	if !errors.Is(err, stats.EmptyInputErr) {
		t.Fatal(err)
	}
}
