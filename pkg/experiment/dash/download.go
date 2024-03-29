package dash

//
// The download phase of the dash experiment.
//

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/ooni/probe-engine/pkg/netxlite"
)

// downloadConfig contains configuration for [download].
type downloadConfig struct {
	// authorization contains the authorization token to perform the download.
	authorization string

	// baseURL is the base URL to use.
	baseURL string

	// begin is the time when we started.
	begin time.Time

	// currentRate is the bitrate at which we request the next chunk.
	currentRate int64

	// deps contains the mockable dependencies.
	deps dependencies

	// elapsedTarget is the desired amount of time that the download
	// of the next chunk should take.
	elapsedTarget int64
}

// downloadResult is the result returned by [download].
type downloadResult struct {
	// elapsed is the elapsed time in seconds
	elapsed float64

	// received is the number of bytes received.
	received int64

	// requestTicks is the time when we started the request in
	// seconds relatively to the config's begin time.
	requestTicks float64

	// serverURL is the URL we downloaded from.
	serverURL string

	// timestamp is the timestamp after the download.
	timestamp int64
}

// download implements one iteration of download phase of the dash experiment. We request
// a chunk from the server and return the measured metrics.
func download(ctx context.Context, config downloadConfig) (downloadResult, error) {
	// compute the desired number of bytes to download
	nbytes := (config.currentRate * 1000 * config.elapsedTarget) >> 3

	// prepare the HTTP request
	var result downloadResult
	URL, err := url.Parse(config.baseURL)
	if err != nil {
		return result, err
	}
	URL.Path = fmt.Sprintf("%s%d", downloadPath, nbytes)
	req, err := config.deps.NewHTTPRequestWithContext(ctx, "GET", URL.String(), nil)
	if err != nil {
		return result, err
	}
	result.serverURL = URL.String()
	req.Header.Set("User-Agent", config.deps.UserAgent())
	req.Header.Set("Authorization", config.authorization)

	// issue the request and get the response
	savedTicks := time.Now()
	resp, err := config.deps.HTTPClient().Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	// make sure that the request is successful
	if resp.StatusCode != 200 {
		return result, errHTTPRequestFailed
	}

	// read the response body
	data, err := netxlite.ReadAllContext(ctx, resp.Body)
	if err != nil {
		return result, err
	}

	// compute performance metrics
	//
	// Implementation note: MK contains a comment that says that Neubot uses
	// the elapsed time since when we start receiving the response but it
	// turns out that Neubot and MK do the same. So, we do what they do. At
	// the same time, we are currently not able to include the overhead that
	// is caused by HTTP headers etc. So, we're a bit less precise.
	result.elapsed = time.Since(savedTicks).Seconds()
	result.received = int64(len(data))
	result.requestTicks = savedTicks.Sub(config.begin).Seconds()
	result.timestamp = time.Now().Unix()
	return result, nil
}
