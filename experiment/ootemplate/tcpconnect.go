package ootemplate

import (
	"context"
	"errors"
	"net"
	"strconv"

	"github.com/ooni/probe-engine/httpx/retryx"
	"github.com/ooni/probe-engine/log"
)

// TCPConnectStatus contains the TCP connect status.
type TCPConnectStatus struct {
	Failure string `json:"failure"`
	Success bool   `json:"success"`
}

// TCPConnectResults contains the results of a TCP connect.
type TCPConnectResults struct {
	IP     string           `json:"ip"`
	Port   uint16           `json:"port"`
	Status TCPConnectStatus `json:"status"`
}

// TCPConnectOnce performs a single TCP connect attempt. There is no real
// timeout except the one you can configure using the context.
func TCPConnectOnce(
	ctx context.Context, logger log.Logger, epnt string,
) (results TCPConnectResults, err error) {
	logger.Debugf("tcpconnect.go: connecting to %s", epnt)
	defer func() {
		if err != nil {
			logger.Debugf("tcpconnect.go: connecting to %s: %s", epnt, err.Error())
			results.Status.Failure = err.Error()
		} else {
			logger.Debugf("tcpconnect.go: connecting to %s: OK", epnt)
			results.Status.Success = true
		}
	}()
	var address, port string
	address, port, err = net.SplitHostPort(epnt)
	if err != nil {
		return
	}
	var portnum int
	portnum, err = strconv.Atoi(port)
	if err != nil {
		return
	}
	if portnum < 0 || portnum > 65535 {
		err = errors.New("invalid port number")
		return
	}
	results.IP = address
	results.Port = uint16(portnum)
	var conn net.Conn
	conn, err = (&net.Dialer{}).DialContext(ctx, "tcp", epnt)
	if err == nil {
		conn.Close()
	}
	return
}

// TCPConnectWithRetry performs a TCP connect and retries a few times before
// giving up. The return value indicates whether we succeded (error equal
// to nil) or failed (non nil error). We also return the result of each of
// the TCP connect attempts.
func TCPConnectWithRetry(
	ctx context.Context, logger log.Logger, epnt string,
) ([]TCPConnectResults, error) {
	var all []TCPConnectResults
	err := retryx.Do(ctx, func() error {
		results, err := TCPConnectOnce(ctx, logger, epnt)
		all = append(all, results)
		return err
	})
	return all, err
}

func tcpConnectAsyncLoop(
	ctx context.Context, out chan<- TCPConnectResults,
	logger log.Logger, epnts ...string,
) {
	defer close(out)
	for _, epnt := range epnts {
		results, _ := TCPConnectWithRetry(ctx, logger, epnt)
		for _, entry := range results {
			out <- entry
			if ctx.Err() != nil {
				return // bail out if the context has expired
			}
		}
	}
}

// TCPConnectAsync connects to an array of endpoints and returns the results
// on the returned channel, which will be closed when done.
func TCPConnectAsync(
	ctx context.Context, logger log.Logger, epnts ...string,
) <-chan TCPConnectResults {
	out := make(chan TCPConnectResults)
	go tcpConnectAsyncLoop(ctx, out, logger, epnts...)
	return out
}
