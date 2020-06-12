package probeservices_test

import (
	"context"
	"testing"

	"github.com/ooni/probe-engine/probeservices"
	"github.com/ooni/probe-engine/probeservices/testorchestra"
)

func TestIntegrationFetchTorTargets(t *testing.T) {
	clnt := newclient()
	if err := clnt.MaybeRegister(context.Background(), testorchestra.MetadataFixture()); err != nil {
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

func TestFetchTorTargetsNotRegistered(t *testing.T) {
	clnt := newclient()
	state := probeservices.State{
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
