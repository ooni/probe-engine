// Package oonitemplates contains templates for experiments.
//
// Every experiment should possibly be based on code inside of
// this package. In the future we should perhaps unify the code
// in here with the code in oonidatamodel.
//
// This has been forked from ooni/netx/x/porcelain because it was
// causing too much changes to keep this code in there.
package oonitemplates

import (
	"context"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	goptlib "git.torproject.org/pluggable-transports/goptlib.git"
	"github.com/m-lab/go/rtx"
	"github.com/ooni/netx"
	"github.com/ooni/netx/handlers"
	"github.com/ooni/netx/httpx"
	"github.com/ooni/netx/modelx"
	"github.com/ooni/netx/x/scoreboard"
	"gitlab.com/yawning/obfs4.git/transports"
	obfs4base "gitlab.com/yawning/obfs4.git/transports/base"
)

type channelHandler struct {
	ch         chan<- modelx.Measurement
	lateWrites int64
}

func (h *channelHandler) OnMeasurement(m modelx.Measurement) {
	// Implementation note: when we're closing idle connections it
	// may be that they're closed once we have stopped reading
	// therefore (1) we MUST NOT close the channel to signal that
	// we're done BECAUSE THIS IS A LIE and (2) we MUST instead
	// arrange here for non-blocking sends.
	select {
	case h.ch <- m:
	case <-time.After(100 * time.Millisecond):
		atomic.AddInt64(&h.lateWrites, 1)
	}
}

// Results contains the results of every operation that we care
// about, as well as the experimental scoreboard, and information
// on the number of bytes received and sent.
//
// When counting the number of bytes sent and received, we do not
// take into account domain name resolutions performed using the
// system resolver. We estimated that using heuristics with MK but
// we currently don't have a good solution. TODO(bassosimone): this
// can be improved by emitting estimates when we know that we are
// using the system resolver, so we can pick up estimates here.
type Results struct {
	Connects      []*modelx.ConnectEvent
	HTTPRequests  []*modelx.HTTPRoundTripDoneEvent
	Resolves      []*modelx.ResolveDoneEvent
	TLSHandshakes []*modelx.TLSHandshakeDoneEvent

	Scoreboard    *scoreboard.Board
	SentBytes     int64
	ReceivedBytes int64
}

func (r *Results) onMeasurement(m modelx.Measurement) {
	if m.Connect != nil {
		r.Connects = append(r.Connects, m.Connect)
	}
	if m.HTTPRoundTripDone != nil {
		r.HTTPRequests = append(r.HTTPRequests, m.HTTPRoundTripDone)
	}
	if m.ResolveDone != nil {
		r.Resolves = append(r.Resolves, m.ResolveDone)
	}
	if m.TLSHandshakeDone != nil {
		r.TLSHandshakes = append(r.TLSHandshakes, m.TLSHandshakeDone)
	}
	if m.Read != nil {
		r.ReceivedBytes += m.Read.NumBytes // overflow unlikely
	}
	if m.Write != nil {
		r.SentBytes += m.Write.NumBytes // overflow unlikely
	}
}

func (r *Results) collect(
	output <-chan modelx.Measurement,
	handler modelx.Handler,
	main func(),
) {
	if handler == nil {
		handler = handlers.NoHandler
	}
	done := make(chan interface{})
	go func() {
		defer close(done)
		main()
	}()
	for {
		select {
		case m := <-output:
			handler.OnMeasurement(m)
			r.onMeasurement(m)
		case <-done:
			return
		}
	}
}

type dnsFallback struct {
	network, address string
}

func configureDNS(seed int64, network, address string) (modelx.DNSResolver, error) {
	resolver, err := netx.NewResolver(handlers.NoHandler, network, address)
	if err != nil {
		return nil, err
	}
	fallbacks := []dnsFallback{
		dnsFallback{
			network: "doh",
			address: "https://cloudflare-dns.com/dns-query",
		},
		dnsFallback{
			network: "doh",
			address: "https://dns.google/dns-query",
		},
		dnsFallback{
			network: "dot",
			address: "8.8.8.8:853",
		},
		dnsFallback{
			network: "dot",
			address: "8.8.4.4:853",
		},
		dnsFallback{
			network: "dot",
			address: "1.1.1.1:853",
		},
		dnsFallback{
			network: "dot",
			address: "9.9.9.9:853",
		},
	}
	random := rand.New(rand.NewSource(seed))
	random.Shuffle(len(fallbacks), func(i, j int) {
		fallbacks[i], fallbacks[j] = fallbacks[j], fallbacks[i]
	})
	var configured int
	for i := 0; configured < 2 && i < len(fallbacks); i++ {
		if fallbacks[i].network == network {
			continue
		}
		var fallback modelx.DNSResolver
		fallback, err = netx.NewResolver(
			handlers.NoHandler, fallbacks[i].network,
			fallbacks[i].address,
		)
		rtx.PanicOnError(err, "porcelain: invalid fallbacks table")
		resolver = netx.ChainResolvers(resolver, fallback)
		configured++
	}
	return resolver, nil
}

// DNSLookupConfig contains DNSLookup settings.
type DNSLookupConfig struct {
	Handler       modelx.Handler
	Hostname      string
	ServerAddress string
	ServerNetwork string
}

// DNSLookupResults contains the results of a DNSLookup
type DNSLookupResults struct {
	TestKeys  Results
	Addresses []string
	Error     error
}

// DNSLookup performs a DNS lookup.
func DNSLookup(
	ctx context.Context, config DNSLookupConfig,
) *DNSLookupResults {
	var (
		mu      sync.Mutex
		results = new(DNSLookupResults)
	)
	channel := make(chan modelx.Measurement)
	root := &modelx.MeasurementRoot{
		Beginning: time.Now(),
		Handler: &channelHandler{
			ch: channel,
		},
	}
	ctx = modelx.WithMeasurementRoot(ctx, root)
	resolver, err := netx.NewResolver(
		handlers.NoHandler,
		config.ServerNetwork,
		config.ServerAddress,
	)
	if err != nil {
		results.Error = err
		return results
	}
	results.TestKeys.collect(channel, config.Handler, func() {
		addrs, err := resolver.LookupHost(ctx, config.Hostname)
		mu.Lock()
		defer mu.Unlock()
		results.Addresses, results.Error = addrs, err
	})
	results.TestKeys.Scoreboard = &root.X.Scoreboard
	return results
}

// HTTPDoConfig contains HTTPDo settings.
type HTTPDoConfig struct {
	Accept             string
	AcceptLanguage     string
	Body               []byte
	DNSServerAddress   string
	DNSServerNetwork   string
	Handler            modelx.Handler
	InsecureSkipVerify bool
	Method             string
	ProxyFunc          func(*http.Request) (*url.URL, error)
	URL                string
	UserAgent          string

	// MaxEventsBodySnapSize controls the snap size that
	// we're using for bodies returned as modelx.Measurement.
	//
	// Same rules as modelx.MeasurementRoot.MaxBodySnapSize.
	MaxEventsBodySnapSize int64

	// MaxResponseBodySnapSize controls the snap size that
	// we're using for the HTTPDoResults.BodySnap.
	//
	// Same rules as modelx.MeasurementRoot.MaxBodySnapSize.
	MaxResponseBodySnapSize int64
}

// HTTPDoResults contains the results of a HTTPDo
type HTTPDoResults struct {
	TestKeys            Results
	StatusCode          int64
	Headers             http.Header
	BodySnap            []byte
	Error               error
	SNIBlockingFollowup *modelx.XSNIBlockingFollowup
}

// HTTPDo performs a HTTP request
func HTTPDo(
	origCtx context.Context, config HTTPDoConfig,
) *HTTPDoResults {
	var (
		mu      sync.Mutex
		results = new(HTTPDoResults)
	)
	channel := make(chan modelx.Measurement)
	// TODO(bassosimone): tell client to use specific CA bundle?
	root := &modelx.MeasurementRoot{
		Beginning: time.Now(),
		Handler: &channelHandler{
			ch: channel,
		},
		MaxBodySnapSize: config.MaxEventsBodySnapSize,
	}
	ctx := modelx.WithMeasurementRoot(origCtx, root)
	client := httpx.NewClientWithProxyFunc(handlers.NoHandler, config.ProxyFunc)
	resolver, err := configureDNS(
		time.Now().UnixNano(),
		config.DNSServerNetwork,
		config.DNSServerAddress,
	)
	if err != nil {
		results.Error = err
		return results
	}
	client.SetResolver(resolver)
	if config.InsecureSkipVerify {
		client.ForceSkipVerify()
	}
	// TODO(bassosimone): implement sending body
	req, err := http.NewRequest(config.Method, config.URL, nil)
	if err != nil {
		results.Error = err
		return results
	}
	if config.Accept != "" {
		req.Header.Set("Accept", config.Accept)
	}
	if config.AcceptLanguage != "" {
		req.Header.Set("Accept-Language", config.AcceptLanguage)
	}
	req.Header.Set("User-Agent", config.UserAgent)
	req = req.WithContext(ctx)
	results.TestKeys.collect(channel, config.Handler, func() {
		defer client.HTTPClient.CloseIdleConnections()
		resp, err := client.HTTPClient.Do(req)
		if err != nil {
			mu.Lock()
			results.Error = err
			mu.Unlock()
			return
		}
		mu.Lock()
		results.StatusCode = int64(resp.StatusCode)
		results.Headers = resp.Header
		mu.Unlock()
		defer resp.Body.Close()
		reader := io.LimitReader(
			resp.Body, modelx.ComputeBodySnapSize(
				config.MaxResponseBodySnapSize,
			),
		)
		data, err := ioutil.ReadAll(reader)
		mu.Lock()
		results.BodySnap, results.Error = data, err
		mu.Unlock()
	})
	results.TestKeys.Scoreboard = &root.X.Scoreboard
	results.SNIBlockingFollowup = maybeRunTLSChecks(
		origCtx, config.Handler, &root.X,
	)
	return results
}

// TLSConnectConfig contains TLSConnect settings.
type TLSConnectConfig struct {
	Address            string
	DNSServerAddress   string
	DNSServerNetwork   string
	Handler            modelx.Handler
	InsecureSkipVerify bool
	SNI                string
}

// TLSConnectResults contains the results of a TLSConnect
type TLSConnectResults struct {
	TestKeys Results
	Error    error
}

// TLSConnect performs a TLS connect.
func TLSConnect(
	ctx context.Context, config TLSConnectConfig,
) *TLSConnectResults {
	var (
		mu      sync.Mutex
		results = new(TLSConnectResults)
	)
	channel := make(chan modelx.Measurement)
	root := &modelx.MeasurementRoot{
		Beginning: time.Now(),
		Handler: &channelHandler{
			ch: channel,
		},
	}
	ctx = modelx.WithMeasurementRoot(ctx, root)
	dialer := netx.NewDialer(handlers.NoHandler)
	// TODO(bassosimone): tell dialer to use specific CA bundle?
	resolver, err := configureDNS(
		time.Now().UnixNano(),
		config.DNSServerNetwork,
		config.DNSServerAddress,
	)
	if err != nil {
		results.Error = err
		return results
	}
	dialer.SetResolver(resolver)
	if config.InsecureSkipVerify {
		dialer.ForceSkipVerify()
	}
	// TODO(bassosimone): can this call really fail?
	dialer.ForceSpecificSNI(config.SNI)
	results.TestKeys.collect(channel, config.Handler, func() {
		conn, err := dialer.DialTLSContext(ctx, "tcp", config.Address)
		if conn != nil {
			defer conn.Close()
		}
		mu.Lock()
		defer mu.Unlock()
		results.Error = err
	})
	results.TestKeys.Scoreboard = &root.X.Scoreboard
	return results
}

func maybeRunTLSChecks(
	ctx context.Context, handler modelx.Handler, x *modelx.XResults,
) (out *modelx.XSNIBlockingFollowup) {
	for _, ev := range x.Scoreboard.TLSHandshakeReset {
		for _, followup := range ev.RecommendedFollowups {
			if followup == "sni_blocking" {
				out = sniBlockingFollowup(ctx, handler, ev.Domain)
				break
			}
		}
	}
	return
}

// TODO(bassosimone): we should make this configurable
const sniBlockingHelper = "example.com:443"

func sniBlockingFollowup(
	ctx context.Context, handler modelx.Handler, domain string,
) (out *modelx.XSNIBlockingFollowup) {
	config := TLSConnectConfig{
		Address: sniBlockingHelper,
		Handler: handler,
		SNI:     domain,
	}
	measurements := TLSConnect(ctx, config)
	out = &modelx.XSNIBlockingFollowup{
		Connects:      measurements.TestKeys.Connects,
		HTTPRequests:  measurements.TestKeys.HTTPRequests,
		Resolves:      measurements.TestKeys.Resolves,
		TLSHandshakes: measurements.TestKeys.TLSHandshakes,
	}
	return
}

// TCPConnectConfig contains TCPConnect settings.
type TCPConnectConfig struct {
	Address          string
	DNSServerAddress string
	DNSServerNetwork string
	Handler          modelx.Handler
}

// TCPConnectResults contains the results of a TCPConnect
type TCPConnectResults struct {
	TestKeys Results
	Error    error
}

// TCPConnect performs a TCP connect.
func TCPConnect(
	ctx context.Context, config TCPConnectConfig,
) *TCPConnectResults {
	var (
		mu      sync.Mutex
		results = new(TCPConnectResults)
	)
	channel := make(chan modelx.Measurement)
	root := &modelx.MeasurementRoot{
		Beginning: time.Now(),
		Handler: &channelHandler{
			ch: channel,
		},
	}
	ctx = modelx.WithMeasurementRoot(ctx, root)
	dialer := netx.NewDialer(handlers.NoHandler)
	// TODO(bassosimone): tell dialer to use specific CA bundle?
	resolver, err := configureDNS(
		time.Now().UnixNano(),
		config.DNSServerNetwork,
		config.DNSServerAddress,
	)
	if err != nil {
		results.Error = err
		return results
	}
	dialer.SetResolver(resolver)
	results.TestKeys.collect(channel, config.Handler, func() {
		conn, err := dialer.DialContext(ctx, "tcp", config.Address)
		if conn != nil {
			defer conn.Close()
		}
		mu.Lock()
		defer mu.Unlock()
		results.Error = err
	})
	results.TestKeys.Scoreboard = &root.X.Scoreboard
	return results
}

func init() {
	rtx.Must(transports.Init(), "transport.Init() failed")
}

// OBFS4ConnectConfig contains OBFS4Connect settings.
type OBFS4ConnectConfig struct {
	Address          string
	DNSServerAddress string
	DNSServerNetwork string
	Handler          modelx.Handler
	Params           goptlib.Args
	StateBaseDir     string
	transportsGet    func(name string) obfs4base.Transport
	ioutilTempDir    func(dir string, prefix string) (string, error)
}

// OBFS4ConnectResults contains the results of a OBFS4Connect
type OBFS4ConnectResults struct {
	TestKeys Results
	Error    error
}

// OBFS4Connect performs a TCP connect.
func OBFS4Connect(
	ctx context.Context, config OBFS4ConnectConfig,
) *OBFS4ConnectResults {
	var (
		mu      sync.Mutex
		results = new(OBFS4ConnectResults)
	)
	channel := make(chan modelx.Measurement)
	root := &modelx.MeasurementRoot{
		Beginning: time.Now(),
		Handler: &channelHandler{
			ch: channel,
		},
	}
	ctx = modelx.WithMeasurementRoot(ctx, root)
	dialer := netx.NewDialer(handlers.NoHandler)
	// TODO(bassosimone): tell dialer to use specific CA bundle?
	resolver, err := configureDNS(
		time.Now().UnixNano(),
		config.DNSServerNetwork,
		config.DNSServerAddress,
	)
	if err != nil {
		results.Error = err
		return results
	}
	dialer.SetResolver(resolver)
	transportsGet := config.transportsGet
	if transportsGet == nil {
		transportsGet = transports.Get
	}
	txp := transportsGet("obfs4")
	ioutilTempDir := config.ioutilTempDir
	if ioutilTempDir == nil {
		ioutilTempDir = ioutil.TempDir
	}
	dirname, err := ioutilTempDir(config.StateBaseDir, "obfs4")
	if err != nil {
		results.Error = err
		return results
	}
	factory, err := txp.ClientFactory(dirname)
	if err != nil {
		results.Error = err
		return results
	}
	parsedargs, err := factory.ParseArgs(&config.Params)
	if err != nil {
		results.Error = err
		return results
	}
	results.TestKeys.collect(channel, config.Handler, func() {
		dialfunc := func(network, address string) (net.Conn, error) {
			return dialer.DialContext(ctx, network, address)
		}
		conn, err := factory.Dial("tcp", config.Address, dialfunc, parsedargs)
		if conn != nil {
			defer conn.Close()
		}
		mu.Lock()
		defer mu.Unlock()
		results.Error = err
	})
	results.TestKeys.Scoreboard = &root.X.Scoreboard
	return results
}
