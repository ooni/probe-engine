package dash

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"runtime"
	"time"

	"github.com/ooni/probe-engine/internal/mlablocate"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/netx/trace"
)

const (
	// libraryName is the name of this library
	libraryName = "miniooni"

	// libraryVersion is the version of this library.
	libraryVersion = "0.1.0-dev"

	// magicVersion is a magic number that identifies in a unique
	// way this implementation of DASH. 0.007xxxyyy is Measurement
	// Kit. Values lower than that are Neubot.
	magicVersion = "0.008000000"
)

var (
	// errServerBusy is returned when the Neubot server is busy.
	errServerBusy = errors.New("Server busy; try again later")

	// errHTTPRequestFailed is returned when an HTTP request fails.
	errHTTPRequestFailed = errors.New("HTTP request failed")
)

type dependencies struct {
	Collect  func(ctx context.Context, authorization string) error
	Download func(
		ctx context.Context, authorization string,
		current *clientResults) error
	HTTPClientDo   func(req *http.Request) (*http.Response, error)
	HTTPNewRequest func(method, url string, body io.Reader) (*http.Request, error)
	IOUtilReadAll  func(r io.Reader) ([]byte, error)
	JSONMarshal    func(v interface{}) ([]byte, error)
	Locate         func(ctx context.Context) (string, error)
	Loop           func(ctx context.Context, ch chan<- clientResults)
	Negotiate      func(ctx context.Context) (negotiateResponse, error)
}

// client is a DASH client
type client struct {
	// ClientName is the name of the client application. This field is
	// initialized by the NewClient constructor.
	ClientName string

	// ClientVersion is the version of the client application. This field is
	// initialized by the NewClient constructor.
	ClientVersion string

	// FQDN is the server of the server to use. If the FQDN is not
	// specified, we'll use mlab-ns to discover a server.
	FQDN string

	// HTTPClient is the HTTP client used by this implementation. This field
	// is initialized by the NewClient to http.DefaultClient.
	HTTPClient *http.Client

	// Logger is the logger to use. This field is initialized by the
	// NewClient constructor to a do-nothing logger.
	Logger dashLogger

	// MLabNSClient is the mlabns client. We'll configure it with
	// defaults in NewClient and you may override it.
	MLabNSClient *mlablocate.Client

	// Scheme is the Scheme to use. By default we configure
	// it to "https", but you can override it to "http".
	Scheme string

	begin         time.Time
	callbacks     model.ExperimentCallbacks
	clientResults []clientResults
	deps          dependencies
	err           error
	numIterations int64
	saver         *trace.Saver
	serverResults []serverResults
	userAgent     string
}

func makeUserAgent(clientName, clientVersion string) string {
	return clientName + "/" + clientVersion + " " + libraryName + "/" + libraryVersion
}

func (c *client) httpClientDo(req *http.Request) (*http.Response, error) {
	return c.HTTPClient.Do(req)
}

func (c *client) locate(ctx context.Context) (string, error) {
	return c.MLabNSClient.Query(ctx, "neubot")
}

// newClient creates a new Client instance using the specified
// client application name and version.
func newClient(
	httpClient *http.Client, saver *trace.Saver, logger model.Logger,
	callbacks model.ExperimentCallbacks,
	clientName, clientVersion string) (clnt *client) {
	ua := makeUserAgent(clientName, clientVersion)
	clnt = &client{
		ClientName:    clientName,
		ClientVersion: clientVersion,
		HTTPClient:    httpClient,
		Logger:        noLogger{},
		MLabNSClient:  mlablocate.NewClient(httpClient, logger, ua),
		Scheme:        "https",
		begin:         time.Now(),
		callbacks:     callbacks,
		numIterations: 15,
		saver:         saver,
		userAgent:     ua,
	}
	clnt.deps = dependencies{
		Collect:        clnt.collect,
		Download:       clnt.download,
		HTTPClientDo:   clnt.httpClientDo,
		HTTPNewRequest: http.NewRequest,
		IOUtilReadAll:  ioutil.ReadAll,
		JSONMarshal:    json.Marshal,
		Locate:         clnt.locate,
		Loop:           clnt.loop,
		Negotiate:      clnt.negotiate,
	}
	return
}

// negotiate is the preliminary phase of Neubot experiment where we connect
// to the server, negotiate test parameters, and obtain an authorization
// token that will be used by us and by the server to identify this experiment.
func (c *client) negotiate(ctx context.Context) (negotiateResponse, error) {
	var negotiateResp negotiateResponse
	data, err := c.deps.JSONMarshal(negotiateRequest{
		DASHRates: defaultRates,
	})
	if err != nil {
		return negotiateResp, err
	}
	c.Logger.Debugf("dash: body: %s", string(data))
	var URL url.URL
	URL.Scheme = c.Scheme
	URL.Host = c.FQDN
	URL.Path = negotiatePath
	req, err := c.deps.HTTPNewRequest("POST", URL.String(), bytes.NewReader(data))
	if err != nil {
		return negotiateResp, err
	}
	c.Logger.Debugf("dash: POST %s", URL.String())
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "")
	resp, err := c.deps.HTTPClientDo(req.WithContext(ctx))
	if err != nil {
		return negotiateResp, err
	}
	c.Logger.Debugf("dash: StatusCode: %d", resp.StatusCode)
	if resp.StatusCode != 200 {
		return negotiateResp, errHTTPRequestFailed
	}
	defer resp.Body.Close()
	data, err = c.deps.IOUtilReadAll(resp.Body)
	if err != nil {
		return negotiateResp, err
	}
	c.Logger.Debugf("dash: body: %s", string(data))
	err = json.Unmarshal(data, &negotiateResp)
	if err != nil {
		return negotiateResp, err
	}
	// Implementation oddity: Neubot is using an integer rather than a
	// boolean for the unchoked, with obvious semantics. I wonder why
	// I choose an integer over a boolean, given that Python does have
	// support for booleans. I don't remember 🤷.
	if negotiateResp.Authorization == "" || negotiateResp.Unchoked == 0 {
		return negotiateResp, errServerBusy
	}
	c.Logger.Debugf("dash: authorization: %s", negotiateResp.Authorization)
	return negotiateResp, nil
}

// download implements the DASH test proper. We compute the number of bytes
// to request given the current rate, download the fake DASH segment, and
// then we return the measured performance of this segment to the caller. This
// is repeated several times to emulate downloading part of a video.
func (c *client) download(
	ctx context.Context, authorization string, current *clientResults,
) error {
	nbytes := (current.Rate * 1000 * current.ElapsedTarget) >> 3
	var URL url.URL
	URL.Scheme = c.Scheme
	URL.Host = c.FQDN
	URL.Path = fmt.Sprintf("%s%d", downloadPath, nbytes)
	req, err := c.deps.HTTPNewRequest("GET", URL.String(), nil)
	if err != nil {
		return err
	}
	c.Logger.Debugf("dash: GET %s", URL.String())
	current.ServerURL = URL.String()
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Authorization", authorization)
	savedTicks := time.Now()
	resp, err := c.deps.HTTPClientDo(req.WithContext(ctx))
	if err != nil {
		return err
	}
	c.Logger.Debugf("dash: StatusCode: %d", resp.StatusCode)
	if resp.StatusCode != 200 {
		return errHTTPRequestFailed
	}
	defer resp.Body.Close()
	data, err := c.deps.IOUtilReadAll(resp.Body)
	if err != nil {
		return err
	}
	// Implementation note: MK contains a comment that says that Neubot uses
	// the elapsed time since when we start receiving the response but it
	// turns out that Neubot and MK do the same. So, we do what they do. At
	// the same time, we are currently not able to include the overhead that
	// is caused by HTTP headers etc. So, we're a bit less precise.
	current.Elapsed = time.Now().Sub(savedTicks).Seconds()
	current.Received = int64(len(data))
	current.RequestTicks = savedTicks.Sub(c.begin).Seconds()
	current.Timestamp = time.Now().Unix()
	//c.Logger.Debugf("dash: current: %+v", current) /* for debugging */
	return nil
}

// collect is the final phase of the test. We send to the server what we
// measured and we receive back what it has measured.
func (c *client) collect(ctx context.Context, authorization string) error {
	data, err := c.deps.JSONMarshal(c.clientResults)
	if err != nil {
		return err
	}
	c.Logger.Debugf("dash: body: %s", string(data))
	var URL url.URL
	URL.Scheme = c.Scheme
	URL.Host = c.FQDN
	URL.Path = collectPath
	req, err := c.deps.HTTPNewRequest("POST", URL.String(), bytes.NewReader(data))
	if err != nil {
		return err
	}
	c.Logger.Debugf("dash: POST %s", URL.String())
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authorization)
	resp, err := c.deps.HTTPClientDo(req.WithContext(ctx))
	if err != nil {
		return err
	}
	c.Logger.Debugf("dash: StatusCode: %d", resp.StatusCode)
	if resp.StatusCode != 200 {
		return errHTTPRequestFailed
	}
	defer resp.Body.Close()
	data, err = c.deps.IOUtilReadAll(resp.Body)
	if err != nil {
		return err
	}
	c.Logger.Debugf("dash: body: %s", string(data))
	err = json.Unmarshal(data, &c.serverResults)
	if err != nil {
		return err
	}
	return nil
}

// loop is the main loop of the DASH test. It performs negotiation, the test
// proper, and then collection. It posts interim results on |ch|.
func (c *client) loop(ctx context.Context, ch chan<- clientResults) {
	defer close(ch)
	// Implementation note: we will soon refactor the server to eliminate the
	// possiblity of keeping clients in queue. For this reason it's becoming
	// increasingly less important to loop waiting for the ready signal. Hence
	// if the server is busy, we just return a well known error.
	var negotiateResp negotiateResponse
	negotiateResp, c.err = c.deps.Negotiate(ctx)
	if c.err != nil {
		return
	}
	// Note: according to a comment in MK sources 3000 kbit/s was the
	// minimum speed recommended by Netflix for SD quality in 2017.
	//
	// See: <https://help.netflix.com/en/node/306>.
	const initialBitrate = 3000
	current := clientResults{
		ElapsedTarget: 2,
		Platform:      runtime.GOOS,
		Rate:          initialBitrate,
		RealAddress:   negotiateResp.RealAddress,
		Version:       magicVersion,
	}
	var connectTime float64
	for current.Iteration < c.numIterations {
		c.err = c.deps.Download(ctx, negotiateResp.Authorization, &current)
		if c.err != nil {
			return
		}
		// Read the events so far and possibly update our measurement
		// of the latest connect time. We should have one sample in most
		// cases, because the connection should be persistent.
		for _, ev := range c.saver.Read() {
			if ev.Name == "connect" {
				connectTime = ev.Duration.Seconds()
			}
		}
		current.ConnectTime = connectTime
		c.clientResults = append(c.clientResults, current)
		ch <- current
		current.Iteration++
		speed := float64(current.Received) / float64(current.Elapsed)
		speed *= 8.0    // to bits per second
		speed /= 1000.0 // to kbit/s
		current.Rate = int64(speed)
	}
	c.err = c.deps.Collect(ctx, negotiateResp.Authorization)
}

// StartDownload starts the DASH download. It returns a channel where
// client measurements are posted, or an error. This function will only
// fail if we cannot even initiate the experiment. If you see some
// results on the returned channel, then maybe it means the experiment
// has somehow worked. You can see if there has been any error during
// the experiment by using the Error function.
func (c *client) StartDownload(ctx context.Context) (<-chan clientResults, error) {
	if c.FQDN == "" {
		c.Logger.Debug("dash: discovering server with mlabns")
		fqdn, err := c.deps.Locate(ctx)
		if err != nil {
			return nil, err
		}
		c.FQDN = fqdn
	}
	c.callbacks.OnProgress(0, fmt.Sprintf("server: %s", c.FQDN))
	ch := make(chan clientResults)
	go c.deps.Loop(ctx, ch)
	return ch, nil
}
