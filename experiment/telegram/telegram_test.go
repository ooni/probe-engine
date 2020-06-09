package telegram_test

import (
	"context"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/experiment/handler"
	"github.com/ooni/probe-engine/experiment/telegram"
	"github.com/ooni/probe-engine/experiment/urlgetter"
	"github.com/ooni/probe-engine/internal/mockable"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/netx/modelx"
)

func TestNewExperimentMeasurer(t *testing.T) {
	measurer := telegram.NewExperimentMeasurer(telegram.Config{})
	if measurer.ExperimentName() != "telegram" {
		t.Fatal("unexpected name")
	}
	if measurer.ExperimentVersion() != "0.1.0" {
		t.Fatal("unexpected version")
	}
}

func TestIntegration(t *testing.T) {
	measurer := telegram.NewExperimentMeasurer(telegram.Config{})
	measurement := new(model.Measurement)
	err := measurer.Run(
		context.Background(),
		&mockable.ExperimentSession{
			MockableLogger: log.Log,
		},
		measurement,
		handler.NewPrinterCallbacks(log.Log),
	)
	if err != nil {
		t.Fatal(err)
	}
	testkeys := measurement.TestKeys.(*telegram.TestKeys)
	if testkeys.Agent != "redirect" {
		t.Fatal("unexpected Agent")
	}
	if testkeys.FailedOperation != nil {
		t.Fatal("unexpected FailedOperation")
	}
	if testkeys.Failure != nil {
		t.Fatal("unexpected Failure")
	}
	if len(testkeys.NetworkEvents) <= 0 {
		t.Fatal("no NetworkEvents?!")
	}
	if len(testkeys.Queries) <= 0 {
		t.Fatal("no Queries?!")
	}
	if len(testkeys.Requests) <= 0 {
		t.Fatal("no Requests?!")
	}
	if len(testkeys.TCPConnect) <= 0 {
		t.Fatal("no TCPConnect?!")
	}
	if len(testkeys.TLSHandshakes) <= 0 {
		t.Fatal("no TLSHandshakes?!")
	}
	if testkeys.TelegramHTTPBlocking != false {
		t.Fatal("unexpected TelegramHTTPBlocking")
	}
	if testkeys.TelegramTCPBlocking != false {
		t.Fatal("unexpected TelegramTCPBlocking")
	}
	if testkeys.TelegramWebFailure != nil {
		t.Fatal("unexpected TelegramWebFailure")
	}
	if testkeys.TelegramWebStatus != "ok" {
		t.Fatal("unexpected TelegramWebStatus")
	}
}

func TestUpdateWithNoAccessPointsBlocking(t *testing.T) {
	testkeys := telegram.NewTestKeys()
	testkeys.Update(urlgetter.MultiOutput{
		Input: urlgetter.MultiInput{
			Config: urlgetter.Config{Method: "POST"},
			Target: "http://149.154.175.50/",
		},
		TestKeys: urlgetter.TestKeys{
			Failure: (func() *string {
				s := modelx.FailureEOFError
				return &s
			})(),
		},
	})
	testkeys.Update(urlgetter.MultiOutput{
		Input: urlgetter.MultiInput{
			Config: urlgetter.Config{Method: "POST"},
			Target: "http://149.154.175.50:443/",
		},
		TestKeys: urlgetter.TestKeys{
			Failure: nil, // this should be enough to declare success
		},
	})
	if testkeys.TelegramHTTPBlocking == true {
		t.Fatal("there should be no TelegramHTTPBlocking")
	}
	if testkeys.TelegramTCPBlocking == true {
		t.Fatal("there should be no TelegramTCPBlocking")
	}
}

func TestUpdateWithNilFailedOperation(t *testing.T) {
	testkeys := telegram.NewTestKeys()
	testkeys.Update(urlgetter.MultiOutput{
		Input: urlgetter.MultiInput{
			Config: urlgetter.Config{Method: "POST"},
			Target: "http://149.154.175.50/",
		},
		TestKeys: urlgetter.TestKeys{
			Failure: (func() *string {
				s := modelx.FailureEOFError
				return &s
			})(),
		},
	})
	testkeys.Update(urlgetter.MultiOutput{
		Input: urlgetter.MultiInput{
			Config: urlgetter.Config{Method: "POST"},
			Target: "http://149.154.175.50:443/",
		},
		TestKeys: urlgetter.TestKeys{
			Failure: (func() *string {
				s := modelx.FailureEOFError
				return &s
			})(),
		},
	})
	if testkeys.TelegramHTTPBlocking == false {
		t.Fatal("there should be TelegramHTTPBlocking")
	}
	if testkeys.TelegramTCPBlocking == true {
		t.Fatal("there should be no TelegramTCPBlocking")
	}
}

func TestUpdateWithNonConnectFailedOperation(t *testing.T) {
	testkeys := telegram.NewTestKeys()
	testkeys.Update(urlgetter.MultiOutput{
		Input: urlgetter.MultiInput{
			Config: urlgetter.Config{Method: "POST"},
			Target: "http://149.154.175.50/",
		},
		TestKeys: urlgetter.TestKeys{
			FailedOperation: (func() *string {
				s := modelx.ConnectOperation
				return &s
			})(),
			Failure: (func() *string {
				s := modelx.FailureEOFError
				return &s
			})(),
		},
	})
	testkeys.Update(urlgetter.MultiOutput{
		Input: urlgetter.MultiInput{
			Config: urlgetter.Config{Method: "POST"},
			Target: "http://149.154.175.50:443/",
		},
		TestKeys: urlgetter.TestKeys{
			FailedOperation: (func() *string {
				s := modelx.HTTPRoundTripOperation
				return &s
			})(),
			Failure: (func() *string {
				s := modelx.FailureEOFError
				return &s
			})(),
		},
	})
	if testkeys.TelegramHTTPBlocking == false {
		t.Fatal("there should be TelegramHTTPBlocking")
	}
	if testkeys.TelegramTCPBlocking == true {
		t.Fatal("there should be no TelegramTCPBlocking")
	}
}

func TestUpdateWithAllConnectsFailed(t *testing.T) {
	testkeys := telegram.NewTestKeys()
	testkeys.Update(urlgetter.MultiOutput{
		Input: urlgetter.MultiInput{
			Config: urlgetter.Config{Method: "POST"},
			Target: "http://149.154.175.50/",
		},
		TestKeys: urlgetter.TestKeys{
			FailedOperation: (func() *string {
				s := modelx.ConnectOperation
				return &s
			})(),
			Failure: (func() *string {
				s := modelx.FailureEOFError
				return &s
			})(),
		},
	})
	testkeys.Update(urlgetter.MultiOutput{
		Input: urlgetter.MultiInput{
			Config: urlgetter.Config{Method: "POST"},
			Target: "http://149.154.175.50:443/",
		},
		TestKeys: urlgetter.TestKeys{
			FailedOperation: (func() *string {
				s := modelx.ConnectOperation
				return &s
			})(),
			Failure: (func() *string {
				s := modelx.FailureEOFError
				return &s
			})(),
		},
	})
	if testkeys.TelegramHTTPBlocking == false {
		t.Fatal("there should be TelegramHTTPBlocking")
	}
	if testkeys.TelegramTCPBlocking == false {
		t.Fatal("there should be TelegramTCPBlocking")
	}
}

func TestUpdateWebWithMixedResults(t *testing.T) {
	testkeys := telegram.NewTestKeys()
	testkeys.Update(urlgetter.MultiOutput{
		Input: urlgetter.MultiInput{
			Config: urlgetter.Config{Method: "GET"},
			Target: "http://web.telegram.org/",
		},
		TestKeys: urlgetter.TestKeys{
			FailedOperation: (func() *string {
				s := modelx.HTTPRoundTripOperation
				return &s
			})(),
			Failure: (func() *string {
				s := modelx.FailureEOFError
				return &s
			})(),
		},
	})
	testkeys.Update(urlgetter.MultiOutput{
		Input: urlgetter.MultiInput{
			Config: urlgetter.Config{Method: "GET"},
			Target: "https://web.telegram.org/",
		},
		TestKeys: urlgetter.TestKeys{
			HTTPResponseBody:   `<title>Telegram Web</title>`,
			HTTPResponseStatus: 200,
		},
	})
	if testkeys.TelegramWebStatus != "blocked" {
		t.Fatal("TelegramWebStatus should be blocked")
	}
	if *testkeys.TelegramWebFailure != modelx.FailureEOFError {
		t.Fatal("invalid TelegramWebFailure")
	}
}

func TestUpdateWithBadRequest(t *testing.T) {
	testkeys := telegram.NewTestKeys()
	testkeys.Update(urlgetter.MultiOutput{
		Input: urlgetter.MultiInput{
			Config: urlgetter.Config{Method: "GET"},
			Target: "http://web.telegram.org/",
		},
		TestKeys: urlgetter.TestKeys{
			HTTPResponseStatus: 400,
		},
	})
	testkeys.Update(urlgetter.MultiOutput{
		Input: urlgetter.MultiInput{
			Config: urlgetter.Config{Method: "GET"},
			Target: "https://web.telegram.org/",
		},
		TestKeys: urlgetter.TestKeys{
			HTTPResponseBody:   `<title>Telegram Web</title>`,
			HTTPResponseStatus: 200,
		},
	})
	if testkeys.TelegramWebStatus != "blocked" {
		t.Fatal("TelegramWebStatus should be blocked")
	}
	if *testkeys.TelegramWebFailure != "http_request_failed" {
		t.Fatal("invalid TelegramWebFailure")
	}
}

func TestUpdateWithMissingTitle(t *testing.T) {
	testkeys := telegram.NewTestKeys()
	testkeys.Update(urlgetter.MultiOutput{
		Input: urlgetter.MultiInput{
			Config: urlgetter.Config{Method: "GET"},
			Target: "http://web.telegram.org/",
		},
		TestKeys: urlgetter.TestKeys{
			HTTPResponseStatus: 200,
			HTTPResponseBody:   "<HTML><title>Telegram Web</title></HTML>",
		},
	})
	testkeys.Update(urlgetter.MultiOutput{
		Input: urlgetter.MultiInput{
			Config: urlgetter.Config{Method: "GET"},
			Target: "http://web.telegram.org/",
		},
		TestKeys: urlgetter.TestKeys{
			HTTPResponseStatus: 200,
			HTTPResponseBody:   "<HTML><title>Antani Web</title></HTML>",
		},
	})
	if testkeys.TelegramWebStatus != "blocked" {
		t.Fatal("TelegramWebStatus should be blocked")
	}
	if *testkeys.TelegramWebFailure != "telegram_missing_title_error" {
		t.Fatal("invalid TelegramWebFailure")
	}
}

func TestUpdateWithAllGood(t *testing.T) {
	testkeys := telegram.NewTestKeys()
	testkeys.Update(urlgetter.MultiOutput{
		Input: urlgetter.MultiInput{
			Config: urlgetter.Config{Method: "GET"},
			Target: "http://web.telegram.org/",
		},
		TestKeys: urlgetter.TestKeys{
			HTTPResponseStatus: 200,
			HTTPResponseBody:   "<HTML><title>Telegram Web</title></HTML>",
		},
	})
	testkeys.Update(urlgetter.MultiOutput{
		Input: urlgetter.MultiInput{
			Config: urlgetter.Config{Method: "GET"},
			Target: "http://web.telegram.org/",
		},
		TestKeys: urlgetter.TestKeys{
			HTTPResponseStatus: 200,
			HTTPResponseBody:   "<HTML><title>Telegram Web</title></HTML>",
		},
	})
	if testkeys.TelegramWebStatus != "ok" {
		t.Fatal("TelegramWebStatus should be ok")
	}
	if testkeys.TelegramWebFailure != nil {
		t.Fatal("invalid TelegramWebFailure")
	}
}
