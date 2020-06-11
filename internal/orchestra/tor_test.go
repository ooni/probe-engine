package orchestra_test

import (
	"context"
	"testing"

	"github.com/ooni/probe-engine/internal/mockable"
	"github.com/ooni/probe-engine/internal/orchestra"
)

func TestIntegrationFetchTorTargets(t *testing.T) {
	clnt := newclient()
	if err := clnt.MaybeRegister(
		context.Background(),
		mockable.OrchestraMetadataFixture(),
	); err != nil {
		t.Fatal(err)
	}
	if err := clnt.MaybeLogin(context.Background()); err != nil {
		t.Fatal(err)
	}
	data, err := clnt.FetchTorTargets(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if data == nil || len(data) <= 0 {
		t.Fatal("invalid data")
	}
}

func TestUnitFetchTorTargetsNotRegistered(t *testing.T) {
	clnt := newclient()
	state := orchestra.State{
		// Explicitly empty so the test is more clear
	}
	if err := clnt.StateFile.Set(state); err != nil {
		t.Fatal(err)
	}
	data, err := clnt.FetchTorTargets(context.Background())
	if err == nil {
		t.Fatal("expected an error here")
	}
	if data != nil {
		t.Fatal("expected nil data here")
	}
}
