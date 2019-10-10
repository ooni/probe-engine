// +build !cgo

package telegram

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/apex/log"
	modelx "github.com/ooni/netx/model"
	"github.com/ooni/probe-engine/session"
)

const (
	softwareName    = "ooniprobe-example"
	softwareVersion = "0.0.1"
)

func TestIntegration(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	ctx := context.Background()

	sess := session.New(
		log.Log, softwareName, softwareVersion, "../../testdata", nil, nil,
	)
	if err := sess.MaybeLookupBackends(ctx); err != nil {
		t.Fatal(err)
	}
	if err := sess.MaybeLookupLocation(ctx); err != nil {
		t.Fatal(err)
	}

	experiment := NewExperiment(sess, Config{})
	if err := experiment.OpenReport(ctx); err != nil {
		t.Fatal(err)
	}
	defer experiment.CloseReport(ctx)

	measurement, err := experiment.Measure(ctx, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := experiment.SubmitMeasurement(ctx, &measurement); err != nil {
		t.Fatal(err)
	}
}

func TestMeasurerRequest(t *testing.T) {
	m := &measurer{
		client: &http.Client{},
		logger: log.Log,
	}
	resp, err := m.request(context.Background(), "\t", "/", "1.2.3.4", "444")
	if err == nil {
		t.Fatal("expected an error here")
	}
	if resp != nil {
		t.Fatal("expected a nil response here")
	}
}

func TestMeasurerMeasureDCRequestError(t *testing.T) {
	m := &measurer{
		client: &http.Client{
			Transport: &failImmediatelyRoundTripper{},
		},
		logger: log.Log,
	}
	m.measureDC(context.Background())
	if m.tk.TelegramHTTPBlocking != true {
		t.Fatal("expected to see telegram HTTP blocking")
	}
}

type failImmediatelyRoundTripper struct{}

func (firt *failImmediatelyRoundTripper) RoundTrip(
	*http.Request,
) (*http.Response, error) {
	return nil, errors.New("mocked error")
}

func TestMeasurerMeasureDCReadBodyError(t *testing.T) {
	m := &measurer{
		client: &http.Client{
			Transport: &failReadBodyRoundTripper{},
		},
		logger: log.Log,
	}
	m.measureDC(context.Background())
	if m.tk.TelegramHTTPBlocking != true {
		t.Fatal("expected to see telegram HTTP blocking")
	}
}

type failReadBodyRoundTripper struct{}

func (firt *failReadBodyRoundTripper) RoundTrip(
	*http.Request,
) (*http.Response, error) {
	return &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(&readFailerBody{}),
	}, nil
}

type readFailerBody struct{}

func (rfb *readFailerBody) Read(b []byte) (int, error) {
	return 0, errors.New("mocked error")
}

func TestMeasurerMeasureWebError(t *testing.T) {
	m := &measurer{
		client: &http.Client{},
		logger: log.Log,
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	m.measureWeb(ctx)
	if m.tk.TelegramWebStatus != "failure" {
		t.Fatal("expected to see failure in telegram web status")
	}
	if m.tk.TelegramWebFailure == nil {
		t.Fatal("expected to see non nil telegram web failure")
	}
}

func TestMeasurerMeasureWebWithSchemeBadStatus(t *testing.T) {
	m := &measurer{
		client: &http.Client{
			Transport: &failResponse{},
		},
		logger: log.Log,
	}
	err := m.measureWebWithScheme(context.Background(), "http")
	if err == nil {
		t.Fatal("expected to see an error here")
	}
}

type failResponse struct{}

func (fr *failResponse) RoundTrip(
	*http.Request,
) (*http.Response, error) {
	return &http.Response{
		StatusCode: 500,
		Body:       ioutil.NopCloser(bytes.NewReader(nil)),
	}, nil
}

func TestMeasurerMeasureWebWithReadBodyError(t *testing.T) {
	m := &measurer{
		client: &http.Client{
			Transport: &failReadBodyRoundTripper{},
		},
		logger: log.Log,
	}
	err := m.measureWebWithScheme(context.Background(), "http")
	if err == nil {
		t.Fatal("expected to see an error here")
	}
}

func TestMeasurerMeasureWebWithNoTitleError(t *testing.T) {
	m := &measurer{
		client: &http.Client{
			Transport: &failNoTitle{},
		},
		logger: log.Log,
	}
	err := m.measureWebWithScheme(context.Background(), "http")
	if err == nil {
		t.Fatal("expected to see an error here")
	}
}

type failNoTitle struct{}

func (fnt *failNoTitle) RoundTrip(
	*http.Request,
) (*http.Response, error) {
	return &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(strings.NewReader("<html></html>")),
	}, nil
}

func TestSetTelegramTCPBlockingFalseOnEmptySlice(t *testing.T) {
	m := &measurer{}
	m.setTelegramTCPBlocking(nil)
	if m.tk.TelegramTCPBlocking != true {
		t.Fatal("expected to see a false value here")
	}
}

func TestSetTelegramTCPBlockingFalseOnErrors(t *testing.T) {
	m := &measurer{}
	m.setTelegramTCPBlocking([][]modelx.Measurement{
		[]modelx.Measurement{
			modelx.Measurement{
				Connect: &modelx.ConnectEvent{
					Error: errors.New("mocked error"),
				},
			},
		},
		[]modelx.Measurement{
			modelx.Measurement{
				Connect: &modelx.ConnectEvent{
					Error: errors.New("mocked error"),
				},
			},
		},
	})
	if m.tk.TelegramTCPBlocking != true {
		t.Fatal("expected to see a false value here")
	}
}

func TestSetTelegramTCPBlockingTrueOnSuccess(t *testing.T) {
	m := &measurer{}
	m.setTelegramTCPBlocking([][]modelx.Measurement{
		[]modelx.Measurement{
			modelx.Measurement{
				Connect: &modelx.ConnectEvent{
					Error: nil,
				},
			},
		},
		[]modelx.Measurement{
			modelx.Measurement{
				Connect: &modelx.ConnectEvent{
					Error: errors.New("mocked error"),
				},
			},
		},
	})
	if m.tk.TelegramTCPBlocking != false {
		t.Fatal("expected to see a true value here")
	}
}
