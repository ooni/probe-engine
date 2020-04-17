package dash

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/apex/log"
	"github.com/montanaflynn/stats"
	"github.com/ooni/probe-engine/experiment/handler"
	"github.com/ooni/probe-engine/internal/mockable"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/netx/trace"
)

func TestUnitRunnerLoopLocateFailure(t *testing.T) {
	expected := errors.New("mocked error")
	r := runner{
		callbacks: handler.NewPrinterCallbacks(log.Log),
		httpClient: &http.Client{
			Transport: FakeHTTPTransport{
				err: expected,
			},
		},
		saver: new(trace.Saver),
		sess: &mockable.ExperimentSession{
			MockableLogger: log.Log,
		},
		tk: new(TestKeys),
	}
	err := r.loop(context.Background(), 1)
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
}

func TestUnitRunnerLoopNegotiateFailure(t *testing.T) {
	expected := errors.New("mocked error")
	r := runner{
		callbacks: handler.NewPrinterCallbacks(log.Log),
		httpClient: &http.Client{
			Transport: &FakeHTTPTransportStack{
				all: []FakeHTTPTransport{
					FakeHTTPTransport{
						resp: &http.Response{
							Body: ioutil.NopCloser(strings.NewReader(
								`{"fqdn": "ams01.measurementlab.net"}`)),
							StatusCode: 200,
						},
					},
					FakeHTTPTransport{err: expected},
				},
			},
		},
		saver: new(trace.Saver),
		sess: &mockable.ExperimentSession{
			MockableLogger: log.Log,
		},
		tk: new(TestKeys),
	}
	err := r.loop(context.Background(), 1)
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
}

func TestUnitRunnerLoopMeasureFailure(t *testing.T) {
	expected := errors.New("mocked error")
	r := runner{
		callbacks: handler.NewPrinterCallbacks(log.Log),
		httpClient: &http.Client{
			Transport: &FakeHTTPTransportStack{
				all: []FakeHTTPTransport{
					FakeHTTPTransport{
						resp: &http.Response{
							Body: ioutil.NopCloser(strings.NewReader(
								`{"fqdn": "ams01.measurementlab.net"}`)),
							StatusCode: 200,
						},
					},
					FakeHTTPTransport{
						resp: &http.Response{
							Body: ioutil.NopCloser(strings.NewReader(
								`{"authorization": "xx", "unchoked": 1}`)),
							StatusCode: 200,
						},
					},
					FakeHTTPTransport{err: expected},
				},
			},
		},
		saver: new(trace.Saver),
		sess: &mockable.ExperimentSession{
			MockableLogger: log.Log,
		},
		tk: new(TestKeys),
	}
	err := r.loop(context.Background(), 1)
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
}

func TestUnitRunnerLoopCollectFailure(t *testing.T) {
	expected := errors.New("mocked error")
	saver := new(trace.Saver)
	saver.Write(trace.Event{Name: "connect", Duration: 150 * time.Millisecond})
	r := runner{
		callbacks: handler.NewPrinterCallbacks(log.Log),
		httpClient: &http.Client{
			Transport: &FakeHTTPTransportStack{
				all: []FakeHTTPTransport{
					FakeHTTPTransport{
						resp: &http.Response{
							Body: ioutil.NopCloser(strings.NewReader(
								`{"fqdn": "ams01.measurementlab.net"}`)),
							StatusCode: 200,
						},
					},
					FakeHTTPTransport{
						resp: &http.Response{
							Body: ioutil.NopCloser(strings.NewReader(
								`{"authorization": "xx", "unchoked": 1}`)),
							StatusCode: 200,
						},
					},
					FakeHTTPTransport{
						resp: &http.Response{
							Body:       ioutil.NopCloser(strings.NewReader(`1234567`)),
							StatusCode: 200,
						},
					},
					FakeHTTPTransport{err: expected},
				},
			},
		},
		saver: saver,
		sess: &mockable.ExperimentSession{
			MockableLogger: log.Log,
		},
		tk: new(TestKeys),
	}
	err := r.loop(context.Background(), 1)
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
}

func TestUnitRunnerLoopSuccess(t *testing.T) {
	saver := new(trace.Saver)
	saver.Write(trace.Event{Name: "connect", Duration: 150 * time.Millisecond})
	r := runner{
		callbacks: handler.NewPrinterCallbacks(log.Log),
		httpClient: &http.Client{
			Transport: &FakeHTTPTransportStack{
				all: []FakeHTTPTransport{
					FakeHTTPTransport{
						resp: &http.Response{
							Body: ioutil.NopCloser(strings.NewReader(
								`{"fqdn": "ams01.measurementlab.net"}`)),
							StatusCode: 200,
						},
					},
					FakeHTTPTransport{
						resp: &http.Response{
							Body: ioutil.NopCloser(strings.NewReader(
								`{"authorization": "xx", "unchoked": 1}`)),
							StatusCode: 200,
						},
					},
					FakeHTTPTransport{
						resp: &http.Response{
							Body:       ioutil.NopCloser(strings.NewReader(`1234567`)),
							StatusCode: 200,
						},
					},
					FakeHTTPTransport{
						resp: &http.Response{
							Body:       ioutil.NopCloser(strings.NewReader(`[]`)),
							StatusCode: 200,
						},
					},
				},
			},
		},
		saver: saver,
		sess: &mockable.ExperimentSession{
			MockableLogger: log.Log,
		},
		tk: new(TestKeys),
	}
	err := r.loop(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
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
