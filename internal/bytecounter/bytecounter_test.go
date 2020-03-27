package bytecounter_test

import (
	"context"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"testing"

	"github.com/ooni/probe-engine/internal/bytecounter"
)

func dorequest(ctx context.Context, url string) error {
	txp := http.DefaultTransport.(*http.Transport).Clone()
	defer txp.CloseIdleConnections()
	dialer := bytecounter.Dialer{Dialer: new(net.Dialer)}
	txp.DialContext = dialer.DialContext
	client := &http.Client{Transport: txp}
	req, err := http.NewRequestWithContext(ctx, "GET", "http://www.google.com", nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if _, err := io.Copy(ioutil.Discard, resp.Body); err != nil {
		return err
	}
	return resp.Body.Close()
}

func TestIntegration(t *testing.T) {
	sess := bytecounter.New()
	ctx := context.Background()
	ctx = bytecounter.WithSessionByteCounter(ctx, sess)
	if err := dorequest(ctx, "http://www.google.com"); err != nil {
		t.Fatal(err)
	}
	exp := bytecounter.New()
	ctx = bytecounter.WithExperimentByteCounter(ctx, exp)
	if err := dorequest(ctx, "http://facebook.com"); err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", sess)
	t.Logf("%+v", exp)
}
