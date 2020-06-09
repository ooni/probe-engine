package urlgetter_test

import (
	"context"
	"errors"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/experiment/handler"
	"github.com/ooni/probe-engine/experiment/urlgetter"
	"github.com/ooni/probe-engine/internal/mockable"
)

func TestMultiIntegration(t *testing.T) {
	multi := urlgetter.Multi{Session: &mockable.ExperimentSession{}}
	inputs := []urlgetter.MultiInput{{
		Config: urlgetter.Config{Method: "HEAD", NoFollowRedirects: true},
		Target: "https://www.google.com",
	}, {
		Config: urlgetter.Config{Method: "HEAD", NoFollowRedirects: true},
		Target: "https://www.facebook.com",
	}, {
		Config: urlgetter.Config{Method: "HEAD", NoFollowRedirects: true},
		Target: "https://www.kernel.org",
	}, {
		Config: urlgetter.Config{Method: "HEAD", NoFollowRedirects: true},
		Target: "https://www.instagram.com",
	}}
	outputs := multi.Collect(context.Background(), inputs, "integration-test",
		handler.NewPrinterCallbacks(log.Log))
	var count int
	for result := range outputs {
		count++
		switch result.Input.Target {
		case "https://www.google.com":
		case "https://www.facebook.com":
		case "https://www.kernel.org":
		case "https://www.instagram.com":
		default:
			t.Fatal("unexpected Input.Target")
		}
		if result.Input.Config.Method != "HEAD" {
			t.Fatal("unexpected Input.Config.Method")
		}
		if result.Err != nil {
			t.Fatal(result.Err)
		}
		if result.TestKeys.Agent != "agent" {
			t.Fatal("invalid TestKeys.Agent")
		}
		if len(result.TestKeys.Queries) != 2 {
			t.Fatal("invalid number of Queries")
		}
		if len(result.TestKeys.Requests) != 1 {
			t.Fatal("invalid number of Requests")
		}
		if len(result.TestKeys.TCPConnect) != 1 {
			t.Fatal("invalid number of TCPConnects")
		}
		if len(result.TestKeys.TLSHandshakes) != 1 {
			t.Fatal("invalid number of TLSHandshakes")
		}
	}
	if count != 4 {
		t.Fatal("invalid number of outputs")
	}
}

func TestMultiContextCanceled(t *testing.T) {
	multi := urlgetter.Multi{Session: &mockable.ExperimentSession{}}
	inputs := []urlgetter.MultiInput{{
		Config: urlgetter.Config{Method: "HEAD", NoFollowRedirects: true},
		Target: "https://www.google.com",
	}, {
		Config: urlgetter.Config{Method: "HEAD", NoFollowRedirects: true},
		Target: "https://www.facebook.com",
	}, {
		Config: urlgetter.Config{Method: "HEAD", NoFollowRedirects: true},
		Target: "https://www.kernel.org",
	}, {
		Config: urlgetter.Config{Method: "HEAD", NoFollowRedirects: true},
		Target: "https://www.instagram.com",
	}}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	outputs := multi.Collect(ctx, inputs, "integration-test",
		handler.NewPrinterCallbacks(log.Log))
	var count int
	for result := range outputs {
		count++
		switch result.Input.Target {
		case "https://www.google.com":
		case "https://www.facebook.com":
		case "https://www.kernel.org":
		case "https://www.instagram.com":
		default:
			t.Fatal("unexpected Input.Target")
		}
		if result.Input.Config.Method != "HEAD" {
			t.Fatal("unexpected Input.Config.Method")
		}
		if !errors.Is(result.Err, context.Canceled) {
			t.Fatal("unexpected error")
		}
		if result.TestKeys.Agent != "agent" {
			t.Fatal("invalid TestKeys.Agent")
		}
		if len(result.TestKeys.Queries) != 0 {
			t.Fatal("invalid number of Queries")
		}
		if len(result.TestKeys.Requests) != 1 {
			t.Fatal("invalid number of Requests")
		}
		if len(result.TestKeys.TCPConnect) != 0 {
			t.Fatal("invalid number of TCPConnects")
		}
		if len(result.TestKeys.TLSHandshakes) != 0 {
			t.Fatal("invalid number of TLSHandshakes")
		}
	}
	if count != 4 {
		t.Fatal("invalid number of outputs")
	}
}
