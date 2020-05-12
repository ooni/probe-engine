package sessiontunnel_test

import (
	"context"
	"errors"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/internal/mockable"
	"github.com/ooni/probe-engine/internal/sessiontunnel"
)

func TestNoTunnel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	tunnel, err := sessiontunnel.Start(ctx, sessiontunnel.Config{
		Name: "",
		Session: &mockable.ExperimentSession{
			MockableLogger: log.Log,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if tunnel != nil {
		t.Fatal("expected nil tunnel here")
	}
}

func TestPsiphonTunnel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	tunnel, err := sessiontunnel.Start(ctx, sessiontunnel.Config{
		Name: "psiphon",
		Session: &mockable.ExperimentSession{
			MockableLogger: log.Log,
		},
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatal("not the error we expected")
	}
	if tunnel != nil {
		t.Fatal("expected nil tunnel here")
	}
}

func TestTorTunnel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	tunnel, err := sessiontunnel.Start(ctx, sessiontunnel.Config{
		Name: "tor",
		Session: &mockable.ExperimentSession{
			MockableLogger: log.Log,
		},
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatal("not the error we expected")
	}
	if tunnel != nil {
		t.Fatal("expected nil tunnel here")
	}
}

func TestInvalidTunnel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	tunnel, err := sessiontunnel.Start(ctx, sessiontunnel.Config{
		Name: "antani",
		Session: &mockable.ExperimentSession{
			MockableLogger: log.Log,
		},
	})
	if err == nil || err.Error() != "unsupported tunnel" {
		t.Fatal("not the error we expected")
	}
	t.Log(tunnel)
	if tunnel != nil {
		t.Fatal("expected nil tunnel here")
	}
}
