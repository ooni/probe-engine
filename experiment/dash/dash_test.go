package dash

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/apex/log"
	"github.com/montanaflynn/stats"
	"github.com/ooni/probe-engine/experiment/handler"
	"github.com/ooni/probe-engine/internal/mockable"
	"github.com/ooni/probe-engine/model"
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
		&mockable.ExperimentSession{
			MockableHTTPClient: http.DefaultClient,
			MockableLogger:     log.Log,
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
		log.Log, c, handler.NewPrinterCallbacks(log.Log),
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
		log.Log, c, handler.NewPrinterCallbacks(log.Log),
	)
	err := runner.loop(context.Background())
	if !errors.Is(err, expect) {
		t.Fatal("not the expected error")
	}
}

func TestUnitRunnerLoopGood(t *testing.T) {
	c := &mockableClient{
		ClientResults: make([]clientResults, 10),
	}
	runner := newRunner(
		log.Log, c, handler.NewPrinterCallbacks(log.Log),
	)
	err := runner.loop(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}

type mockableClient struct {
	StartDownloadError error
	ErrorResult        error
	ClientResults      []clientResults
}

func (c *mockableClient) StartDownload(
	ctx context.Context,
) (<-chan clientResults, error) {
	ch := make(chan clientResults)
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

func (c *mockableClient) ServerResults() []serverResults {
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
		ReceiverData: []clientResults{
			{
				Rate: 1,
			},
			{
				Rate: 2,
			},
			{
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
		ReceiverData: []clientResults{
			{
				ElapsedTarget: 2,
				Elapsed:       1.4,
			},
			{
				ElapsedTarget: 2,
				Elapsed:       3.0,
			},
			{
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

func TestUnitRunnerDoWithLoopSuccess(t *testing.T) {
	c := &mockableClient{
		ClientResults: make([]clientResults, 10),
	}
	runner := newRunner(
		log.Log, c, handler.NewPrinterCallbacks(log.Log),
	)
	err := runner.do(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnitRunnerDoWithNoData(t *testing.T) {
	c := &mockableClient{}
	runner := newRunner(
		log.Log, c, handler.NewPrinterCallbacks(log.Log),
	)
	err := runner.do(context.Background())
	if !errors.Is(err, stats.EmptyInputErr) {
		t.Fatal(err)
	}
}
