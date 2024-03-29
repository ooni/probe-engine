package tunnel_test

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/pkg/engine"
	"github.com/ooni/probe-engine/pkg/tunnel"
)

func TestFakeStartStop(t *testing.T) {
	// no need to skip because the bootstrap is obviously fast
	tunnelDir, err := ioutil.TempDir("testdata", "fake")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	sess, err := engine.NewSession(ctx, engine.SessionConfig{
		Logger:          log.Log,
		SoftwareName:    "miniooni",
		SoftwareVersion: "0.1.0-dev",
		TunnelDir:       tunnelDir,
	})
	if err != nil {
		t.Fatal(err)
	}
	tunnel, _, err := tunnel.Start(context.Background(), &tunnel.Config{
		Name:      "fake",
		Session:   sess,
		TunnelDir: tunnelDir,
	})
	if err != nil {
		t.Fatal(err)
	}
	if tunnel.SOCKS5ProxyURL() == nil {
		t.Fatal("expected non nil URL here")
	}
	if tunnel.BootstrapTime() <= 0 {
		t.Fatal("expected positive bootstrap time here")
	}
	tunnel.Stop()
}
