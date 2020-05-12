package ndt7

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/netx/dialer"
	"github.com/ooni/probe-engine/netx/resolver"
)

type dialManager struct {
	hostname        string
	logger          model.Logger
	port            string
	proxyURL        *url.URL
	readBufferSize  int
	scheme          string
	tlsConfig       *tls.Config
	userAgent       string
	writeBufferSize int
}

func newDialManager(
	hostname string, proxyURL *url.URL, logger model.Logger,
	userAgent string) dialManager {
	return dialManager{
		hostname:        hostname,
		logger:          logger,
		port:            "443",
		proxyURL:        proxyURL,
		readBufferSize:  paramMaxMessageSize,
		scheme:          "wss",
		userAgent:       userAgent,
		writeBufferSize: paramMaxMessageSize,
	}
}

func (mgr dialManager) dialWithTestName(ctx context.Context, testName string) (*websocket.Conn, error) {
	var reso resolver.Resolver = resolver.SystemResolver{}
	reso = resolver.LoggingResolver{Resolver: reso, Logger: mgr.logger}
	var dlr dialer.Dialer = new(net.Dialer)
	dlr = dialer.TimeoutDialer{Dialer: dlr}
	dlr = dialer.ErrorWrapperDialer{Dialer: dlr}
	dlr = dialer.LoggingDialer{Dialer: dlr, Logger: mgr.logger}
	dlr = dialer.DNSDialer{Dialer: dlr, Resolver: reso}
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
	headers.Add("User-Agent", mgr.userAgent)
	mgr.logrequest(URL.String(), headers)
	conn, _, err := dialer.DialContext(ctx, URL.String(), headers)
	mgr.logresponse(err)
	return conn, err
}

func (mgr dialManager) logrequest(url string, headers http.Header) {
	mgr.logger.Debugf("> GET %s", url)
	for key, values := range headers {
		for _, v := range values {
			mgr.logger.Debugf("> %s: %s", key, v)
		}
	}
	mgr.logger.Debug("> Connection: upgrade")
	mgr.logger.Debug("> Upgrade: websocket")
	mgr.logger.Debug(">")
}

func (mgr dialManager) logresponse(err error) {
	if err != nil {
		mgr.logger.Debugf("< %+v", err)
		return
	}
	mgr.logger.Debug("< 101")
	mgr.logger.Debug("< Connection: upgrade")
	mgr.logger.Debug("< Upgrade: websocket")
	mgr.logger.Debug("<")
}

func (mgr dialManager) dialDownload(ctx context.Context) (*websocket.Conn, error) {
	return mgr.dialWithTestName(ctx, "download")
}

func (mgr dialManager) dialUpload(ctx context.Context) (*websocket.Conn, error) {
	return mgr.dialWithTestName(ctx, "upload")
}
