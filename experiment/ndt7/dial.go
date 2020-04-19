package ndt7

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/ooni/probe-engine/netx/dialer"
)

type dialManager struct {
	hostname        string
	port            string
	proxyURL        *url.URL
	readBufferSize  int
	scheme          string
	tlsConfig       *tls.Config
	writeBufferSize int
}

func newDialManager(hostname string, proxyURL *url.URL) dialManager {
	return dialManager{
		hostname:        hostname,
		port:            "443",
		proxyURL:        proxyURL,
		readBufferSize:  paramMaxMessageSize,
		scheme:          "wss",
		writeBufferSize: paramMaxMessageSize,
	}
}

func (mgr dialManager) dialWithTestName(ctx context.Context, testName string) (*websocket.Conn, error) {
	var dlr dialer.Dialer = dialer.DNSDialer{
		Dialer: dialer.ErrorWrapperDialer{
			Dialer: dialer.TimeoutDialer{
				Dialer: new(net.Dialer),
			},
		},
		Resolver: new(net.Resolver),
	}
	dlr = dialer.ProxyDialer{Dialer: dlr, ProxyURL: mgr.proxyURL}
	dlr = dialer.ByteCounterDialer{Dialer: dlr}
	dialer := websocket.Dialer{
		NetDialContext:  dlr.DialContext,
		ReadBufferSize:  mgr.readBufferSize,
		TLSClientConfig: mgr.tlsConfig,
		WriteBufferSize: mgr.writeBufferSize,
	}
	URL := url.URL{
		Scheme: mgr.scheme,
		Host:   mgr.hostname + ":" + mgr.port,
	}
	URL.Path = "/ndt/v7/" + testName
	headers := http.Header{}
	headers.Add("Sec-WebSocket-Protocol", "net.measurementlab.ndt.v7")
	conn, _, err := dialer.DialContext(ctx, URL.String(), headers)
	return conn, err
}

func (mgr dialManager) dialDownload(ctx context.Context) (*websocket.Conn, error) {
	return mgr.dialWithTestName(ctx, "download")
}

func (mgr dialManager) dialUpload(ctx context.Context) (*websocket.Conn, error) {
	return mgr.dialWithTestName(ctx, "upload")
}
