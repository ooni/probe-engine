// +build !cgo

package telegram

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/netx/modelx"
	"github.com/ooni/netx/x/porcelain"
	"github.com/ooni/probe-engine/experiment/handler"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/session"
)

const (
	softwareName    = "ooniprobe-example"
	softwareVersion = "0.0.1"
)

func TestUnitNewExperiment(t *testing.T) {
	sess := session.New(
		log.Log, softwareName, softwareVersion,
		"../../testdata", nil, nil, "../../testdata",
	)
	experiment := NewExperiment(sess, Config{})
	if experiment == nil {
		t.Fatal("nil experiment returned")
	}
}

func TestUnitMeasureWithCancelledContext(t *testing.T) {
	m := newMeasurer(Config{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := m.measure(
		ctx,
		&session.Session{
			Logger: log.Log,
		},
		new(model.Measurement),
		handler.NewPrinterCallbacks(log.Log),
	)
	if err == nil {
		t.Fatal("expected an error here")
	}
	if err.Error() != "passed nil results" {
		t.Fatal("unexpected error")
	}
}

func TestIntegrationMeasure(t *testing.T) {
	m := newMeasurer(Config{})
	err := m.measure(
		context.Background(),
		&session.Session{
			Logger: log.Log,
		},
		new(model.Measurement),
		handler.NewPrinterCallbacks(log.Log),
	)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnitProcessoneNil(t *testing.T) {
	tk := newTestKeys()
	err := tk.processone(nil)
	if err == nil {
		t.Fatal("expected an error here")
	}
	if err.Error() != "passed nil data to processone" {
		t.Fatal("not the error we expected")
	}
}

func TestUnitProcessallWithNoAccessPointsBlocking(t *testing.T) {
	tk := newTestKeys()
	err := tk.processall(map[string]*urlMeasurements{
		"http://149.154.175.50/": &urlMeasurements{
			method: "POST",
			results: &porcelain.HTTPDoResults{
				Error: errors.New("mocked error"),
			},
		},
		"http://149.154.175.50:443/": &urlMeasurements{
			method: "POST",
			results: &porcelain.HTTPDoResults{
				Error: nil, // this should be enough to declare success
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if tk.TelegramHTTPBlocking == true {
		t.Fatal("there should be no TelegramHTTPBlocking")
	}
	if tk.TelegramTCPBlocking == true {
		t.Fatal("there should be no TelegramTCPBlocking")
	}
}

func TestUnitProcessallWithTelegramHTTPBlocking(t *testing.T) {
	tk := newTestKeys()
	err := tk.processall(map[string]*urlMeasurements{
		"http://149.154.175.50/": &urlMeasurements{
			method: "POST",
			results: &porcelain.HTTPDoResults{
				Error: errors.New("mocked error"),
			},
		},
		"http://149.154.175.50:443/": &urlMeasurements{
			method: "POST",
			results: &porcelain.HTTPDoResults{
				Error: errors.New("mocked error"),
				TestKeys: porcelain.Results{
					Connects: []*modelx.ConnectEvent{
						&modelx.ConnectEvent{
							Error: nil, // enough  to declare we can TCP connect
						},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if tk.TelegramHTTPBlocking == false {
		t.Fatal("there should be TelegramHTTPBlocking")
	}
	if tk.TelegramTCPBlocking == true {
		t.Fatal("there should be no TelegramTCPBlocking")
	}
}

func TestUnitProcessallWithMixedResults(t *testing.T) {
	tk := newTestKeys()
	err := tk.processall(map[string]*urlMeasurements{
		"http://web.telegram.org/": &urlMeasurements{
			method: "GET",
			results: &porcelain.HTTPDoResults{
				Error: errors.New("mocked error"),
			},
		},
		"https://web.telegram.org/": &urlMeasurements{
			method: "GET",
			results: &porcelain.HTTPDoResults{
				Error:      nil,
				StatusCode: 200,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if tk.TelegramWebStatus != "blocked" {
		t.Fatal("TelegramWebStatus should be blocked")
	}
	// To better understand https://github.com/ooni/probe-engine/issues/142
	t.Logf("%+v", *tk.TelegramWebFailure)
	if *tk.TelegramWebFailure != "mocked error" {
		t.Fatal("invalid TelegramWebFailure")
	}
}

func TestUnitProcessallWithBadRequest(t *testing.T) {
	tk := newTestKeys()
	err := tk.processall(map[string]*urlMeasurements{
		"http://web.telegram.org/": &urlMeasurements{
			method: "GET",
			results: &porcelain.HTTPDoResults{
				StatusCode: 400,
			},
		},
		"https://web.telegram.org/": &urlMeasurements{
			method: "GET",
			results: &porcelain.HTTPDoResults{
				Error: nil,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if tk.TelegramWebStatus != "blocked" {
		t.Fatal("TelegramWebStatus should be blocked")
	}
	if *tk.TelegramWebFailure != "http_request_failed" {
		t.Fatal("invalid TelegramWebFailure")
	}
}

func TestUnitProcessallWithMissingTitle(t *testing.T) {
	tk := newTestKeys()
	err := tk.processall(map[string]*urlMeasurements{
		"http://web.telegram.org/": &urlMeasurements{
			method: "GET",
			results: &porcelain.HTTPDoResults{
				StatusCode: 200,
				BodySnap:   []byte("<HTML><title>Telegram Web</title></HTML>"),
			},
		},
		"https://web.telegram.org/": &urlMeasurements{
			method: "GET",
			results: &porcelain.HTTPDoResults{
				StatusCode: 200,
				BodySnap:   []byte("<HTML><TITLE>Antani Web</TITLE></HTML>"),
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if tk.TelegramWebStatus != "blocked" {
		t.Fatal("TelegramWebStatus should be blocked")
	}
	if *tk.TelegramWebFailure != "telegram_missing_title_error" {
		t.Fatal("invalid TelegramWebFailure")
	}
}

func TestUnitProcessallWithAllGood(t *testing.T) {
	tk := newTestKeys()
	err := tk.processall(map[string]*urlMeasurements{
		"http://web.telegram.org/": &urlMeasurements{
			method: "GET",
			results: &porcelain.HTTPDoResults{
				StatusCode: 200,
				BodySnap:   []byte("<HTML><title>Telegram Web</title></HTML>"),
			},
		},
		"https://web.telegram.org/": &urlMeasurements{
			method: "GET",
			results: &porcelain.HTTPDoResults{
				StatusCode: 200,
				BodySnap:   []byte("<HTML><title>Telegram Web</title></HTML>"),
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if tk.TelegramWebStatus != "ok" {
		t.Fatal("TelegramWebStatus should be ok")
	}
	if tk.TelegramWebFailure != nil {
		t.Fatal("invalid TelegramWebFailure")
	}
}

func TestUnitErrString(t *testing.T) {
	if errString(nil) != "success" {
		t.Fatal("unexpected value with nil error")
	}
	if errString(io.EOF) != "EOF" {
		t.Fatal("unexpected value with real error")
	}
}
