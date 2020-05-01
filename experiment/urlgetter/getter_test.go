package urlgetter_test

import (
	"context"
	"errors"
	"io"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ooni/probe-engine/experiment/urlgetter"
	"github.com/ooni/probe-engine/internal/mockable"
)

func TestGetterWithCancelledContextVanilla(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	g := urlgetter.Getter{
		Session: &mockable.ExperimentSession{},
		Target:  "https://www.google.com",
	}
	tk, err := g.Get(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatal("not the error we expected")
	}
	if tk.Agent != "redirect" {
		t.Fatal("not the Agent we expected")
	}
	if tk.BootstrapTime != 0 {
		t.Fatal("not the BootstrapTime we expected")
	}
	if tk.Failure == nil || !strings.HasSuffix(*tk.Failure, "context canceled") {
		t.Fatal("not the Failure we expected")
	}
	if len(tk.NetworkEvents) != 3 {
		t.Fatal("not the NetworkEvents we expected")
	}
	if tk.NetworkEvents[0].Operation != "http_transaction_start" {
		t.Fatal("not the NetworkEvents[0].Operation we expected")
	}
	if tk.NetworkEvents[1].Operation != "http_request_metadata" {
		t.Fatal("not the NetworkEvents[1].Operation we expected")
	}
	if tk.NetworkEvents[2].Operation != "http_transaction_done" {
		t.Fatal("not the NetworkEvents[2].Operation we expected")
	}
	if len(tk.Queries) != 0 {
		t.Fatal("not the Queries we expected")
	}
	if len(tk.Requests) != 1 {
		t.Fatal("not the Requests we expected")
	}
	if tk.Requests[0].Request.Method != "GET" {
		t.Fatal("not the Method we expected")
	}
	if tk.Requests[0].Request.URL != "https://www.google.com" {
		t.Fatal("not the URL we expected")
	}
	if tk.SOCKSProxy != "" {
		t.Fatal("not the SOCKSProxy we expected")
	}
	if len(tk.TLSHandshakes) != 0 {
		t.Fatal("not the TLSHandshakes we expected")
	}
	if tk.Tunnel != "" {
		t.Fatal("not the Tunnel we expected")
	}
}

func TestGetterWithCancelledContextNoFollowRedirects(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	g := urlgetter.Getter{
		Config: urlgetter.Config{
			NoFollowRedirects: true,
		},
		Session: &mockable.ExperimentSession{},
		Target:  "https://www.google.com",
	}
	tk, err := g.Get(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatal("not the error we expected")
	}
	if tk.Agent != "agent" {
		t.Fatal("not the Agent we expected")
	}
	if tk.BootstrapTime != 0 {
		t.Fatal("not the BootstrapTime we expected")
	}
	if tk.Failure == nil || !strings.HasSuffix(*tk.Failure, "context canceled") {
		t.Fatal("not the Failure we expected")
	}
	if len(tk.NetworkEvents) != 3 {
		t.Fatal("not the NetworkEvents we expected")
	}
	if tk.NetworkEvents[0].Operation != "http_transaction_start" {
		t.Fatal("not the NetworkEvents[0].Operation we expected")
	}
	if tk.NetworkEvents[1].Operation != "http_request_metadata" {
		t.Fatal("not the NetworkEvents[1].Operation we expected")
	}
	if tk.NetworkEvents[2].Operation != "http_transaction_done" {
		t.Fatal("not the NetworkEvents[2].Operation we expected")
	}
	if len(tk.Queries) != 0 {
		t.Fatal("not the Queries we expected")
	}
	if len(tk.Requests) != 1 {
		t.Fatal("not the Requests we expected")
	}
	if tk.Requests[0].Request.Method != "GET" {
		t.Fatal("not the Method we expected")
	}
	if tk.Requests[0].Request.URL != "https://www.google.com" {
		t.Fatal("not the URL we expected")
	}
	if tk.SOCKSProxy != "" {
		t.Fatal("not the SOCKSProxy we expected")
	}
	if len(tk.TLSHandshakes) != 0 {
		t.Fatal("not the TLSHandshakes we expected")
	}
	if tk.Tunnel != "" {
		t.Fatal("not the Tunnel we expected")
	}
}

func TestGetterWithCancelledContextCannotStartTunnel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	g := urlgetter.Getter{
		Session: &mockable.ExperimentSession{
			MockableMaybeStartTunnelErr: io.EOF,
		},
		Target: "https://www.google.com",
	}
	tk, err := g.Get(ctx)
	if !errors.Is(err, io.EOF) {
		t.Fatal("not the error we expected")
	}
	if tk.Agent != "redirect" {
		t.Fatal("not the Agent we expected")
	}
	if tk.BootstrapTime != 0 {
		t.Fatal("not the BootstrapTime we expected")
	}
	if tk.Failure == nil || !strings.HasSuffix(*tk.Failure, "EOF") {
		t.Fatal("not the Failure we expected")
	}
	if len(tk.NetworkEvents) != 0 {
		t.Fatal("not the NetworkEvents we expected")
	}
	if len(tk.Queries) != 0 {
		t.Fatal("not the Queries we expected")
	}
	if len(tk.Requests) != 0 {
		t.Fatal("not the Requests we expected")
	}
	if tk.SOCKSProxy != "" {
		t.Fatal("not the SOCKSProxy we expected")
	}
	if len(tk.TLSHandshakes) != 0 {
		t.Fatal("not the TLSHandshakes we expected")
	}
	if tk.Tunnel != "" {
		t.Fatal("not the Tunnel we expected")
	}
}

func TestGetterWithCancelledContextWithTunnel(t *testing.T) {
	tunnelURL, _ := url.Parse("socks5://127.0.0.1:9050")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	g := urlgetter.Getter{
		Config: urlgetter.Config{
			Tunnel: "psiphon",
		},
		Session: &mockable.ExperimentSession{
			MockableProxyURL:            tunnelURL,
			MockableTunnelBootstrapTime: 10 * time.Second,
		},
		Target: "https://www.google.com",
	}
	tk, err := g.Get(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatal("not the error we expected")
	}
	if tk.Agent != "redirect" {
		t.Fatal("not the Agent we expected")
	}
	if tk.BootstrapTime != 10.0 {
		t.Fatal("not the BootstrapTime we expected")
	}
	if tk.Failure == nil || !strings.HasSuffix(*tk.Failure, "context canceled") {
		t.Fatal("not the Failure we expected")
	}
	if len(tk.NetworkEvents) != 3 {
		t.Fatal("not the NetworkEvents we expected")
	}
	if tk.NetworkEvents[0].Operation != "http_transaction_start" {
		t.Fatal("not the NetworkEvents[0].Operation we expected")
	}
	if tk.NetworkEvents[1].Operation != "http_request_metadata" {
		t.Fatal("not the NetworkEvents[1].Operation we expected")
	}
	if tk.NetworkEvents[2].Operation != "http_transaction_done" {
		t.Fatal("not the NetworkEvents[2].Operation we expected")
	}
	if len(tk.Queries) != 0 {
		t.Fatal("not the Queries we expected")
	}
	if len(tk.Requests) != 1 {
		t.Fatal("not the Requests we expected")
	}
	if tk.Requests[0].Request.Method != "GET" {
		t.Fatal("not the Method we expected")
	}
	if tk.Requests[0].Request.URL != "https://www.google.com" {
		t.Fatal("not the URL we expected")
	}
	if tk.SOCKSProxy != "127.0.0.1:9050" {
		t.Fatal("not the SOCKSProxy we expected")
	}
	if len(tk.TLSHandshakes) != 0 {
		t.Fatal("not the TLSHandshakes we expected")
	}
	if tk.Tunnel != "psiphon" {
		t.Fatal("not the Tunnel we expected")
	}
}

func TestGetterWithCancelledContextUnknownResolverURL(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	g := urlgetter.Getter{
		Config: urlgetter.Config{
			ResolverURL: "antani://8.8.8.8:53",
		},
		Session: &mockable.ExperimentSession{},
		Target:  "https://www.google.com",
	}
	tk, err := g.Get(ctx)
	if err == nil || err.Error() != "unsupported resolver scheme" {
		t.Fatal("not the error we expected")
	}
	if tk.Agent != "redirect" {
		t.Fatal("not the Agent we expected")
	}
	if tk.BootstrapTime != 0 {
		t.Fatal("not the BootstrapTime we expected")
	}
	if tk.Failure == nil || *tk.Failure != "unsupported resolver scheme" {
		t.Fatal("not the Failure we expected")
	}
	if len(tk.NetworkEvents) != 0 {
		t.Fatal("not the NetworkEvents we expected")
	}
	if len(tk.Queries) != 0 {
		t.Fatal("not the Queries we expected")
	}
	if len(tk.Requests) != 0 {
		t.Fatal("not the Requests we expected")
	}
	if tk.SOCKSProxy != "" {
		t.Fatal("not the SOCKSProxy we expected")
	}
	if len(tk.TLSHandshakes) != 0 {
		t.Fatal("not the TLSHandshakes we expected")
	}
	if tk.Tunnel != "" {
		t.Fatal("not the Tunnel we expected")
	}
}
