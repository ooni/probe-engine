package urlgetter

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"github.com/ooni/probe-engine/internal/httpheader"
	"github.com/ooni/probe-engine/internal/runtimex"
	"github.com/ooni/probe-engine/netx/httptransport"
)

// The Runner job is to run a single measurement
type Runner struct {
	Config     Config
	HTTPConfig httptransport.Config
	Target     string
}

// Run runs a measurement and returns the measurement result
func (r Runner) Run(ctx context.Context) error {
	targetURL, err := url.Parse(r.Target)
	if err != nil {
		return fmt.Errorf("urlgetter: invalid target URL: %w", err)
	}
	switch targetURL.Scheme {
	case "http", "https":
		return r.httpGet(ctx, r.Target)
	case "dnslookup":
		return r.dnsLookup(ctx, targetURL.Hostname())
	case "tlshandshake":
		return r.tlsHandshake(ctx, targetURL.Host)
	case "tcpconnect":
		return r.tcpConnect(ctx, targetURL.Host)
	default:
		return errors.New("unknown targetURL scheme")
	}
}

func (r Runner) httpGet(ctx context.Context, url string) error {
	req, err := http.NewRequest("GET", url, nil)
	runtimex.PanicOnError(err, "http.NewRequest failed")
	req = req.WithContext(ctx)
	req.Header.Set("Accept", httpheader.RandomAccept())
	req.Header.Set("Accept-Language", httpheader.RandomAcceptLanguage())
	req.Header.Set("User-Agent", httpheader.RandomUserAgent())
	if r.Config.HTTPHost != "" {
		req.Host = r.Config.HTTPHost
	}
	// Implementation note: the following cookiejar accepts all cookies
	// from all domains. As such, would not be safe for usage where cookies
	// metter, but it's totally fine for performing measurements.
	jar, err := cookiejar.New(nil)
	runtimex.PanicOnError(err, "cookiejar.New failed")
	httpClient := &http.Client{
		Jar:       jar,
		Transport: httptransport.New(r.HTTPConfig),
	}
	if r.Config.NoFollowRedirects {
		httpClient.CheckRedirect = func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
	defer httpClient.CloseIdleConnections()
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(ioutil.Discard, resp.Body)
	return err
}

func (r Runner) dnsLookup(ctx context.Context, hostname string) error {
	resolver := httptransport.NewResolver(r.HTTPConfig)
	_, err := resolver.LookupHost(ctx, hostname)
	return err
}

func (r Runner) tlsHandshake(ctx context.Context, address string) error {
	tlsDialer := httptransport.NewTLSDialer(r.HTTPConfig)
	conn, err := tlsDialer.DialTLSContext(ctx, "tcp", address)
	if conn != nil {
		conn.Close()
	}
	return err
}

func (r Runner) tcpConnect(ctx context.Context, address string) error {
	dialer := httptransport.NewDialer(r.HTTPConfig)
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if conn != nil {
		conn.Close()
	}
	return err
}
