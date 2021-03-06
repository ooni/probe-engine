package internal

import (
	"context"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/ooni/probe-engine/atomicx"
	"github.com/ooni/probe-engine/netx"
)

type FakeResolver struct {
	NumFailures *atomicx.Int64
	Err         error
	Result      []string
}

func NewFakeResolverThatFails() FakeResolver {
	return FakeResolver{NumFailures: atomicx.NewInt64(), Err: ErrNotFound}
}

func NewFakeResolverWithResult(r []string) FakeResolver {
	return FakeResolver{NumFailures: atomicx.NewInt64(), Result: r}
}

var ErrNotFound = &net.DNSError{
	Err: "no such host",
}

func (c FakeResolver) LookupHost(ctx context.Context, hostname string) ([]string, error) {
	time.Sleep(10 * time.Microsecond)
	if c.Err != nil {
		if c.NumFailures != nil {
			c.NumFailures.Add(1)
		}
		return nil, c.Err
	}
	return c.Result, nil
}

func (c FakeResolver) Network() string {
	return "fake"
}

func (c FakeResolver) Address() string {
	return ""
}

var _ netx.Resolver = FakeResolver{}

type FakeTransport struct {
	Err  error
	Func func(*http.Request) (*http.Response, error)
	Resp *http.Response
}

func (txp FakeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	time.Sleep(10 * time.Microsecond)
	if txp.Func != nil {
		return txp.Func(req)
	}
	if req.Body != nil {
		ioutil.ReadAll(req.Body)
		req.Body.Close()
	}
	if txp.Err != nil {
		return nil, txp.Err
	}
	txp.Resp.Request = req // non thread safe but it doesn't matter
	return txp.Resp, nil
}

func (txp FakeTransport) CloseIdleConnections() {}

var _ netx.HTTPRoundTripper = FakeTransport{}

type FakeBody struct {
	Data []byte
	Err  error
}

func (fb *FakeBody) Read(p []byte) (int, error) {
	time.Sleep(10 * time.Microsecond)
	if fb.Err != nil {
		return 0, fb.Err
	}
	if len(fb.Data) <= 0 {
		return 0, io.EOF
	}
	n := copy(p, fb.Data)
	fb.Data = fb.Data[n:]
	return n, nil
}

func (fb *FakeBody) Close() error {
	return nil
}

var _ io.ReadCloser = &FakeBody{}
