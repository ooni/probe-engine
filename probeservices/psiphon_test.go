package probeservices_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/ooni/probe-engine/probeservices"
	"github.com/ooni/probe-engine/probeservices/testorchestra"
)

func TestIntegrationFetchPsiphonConfig(t *testing.T) {
	clnt := newclient()
	if err := clnt.MaybeRegister(
		context.Background(),
		testorchestra.MetadataFixture(),
	); err != nil {
		t.Fatal(err)
	}
	if err := clnt.MaybeLogin(context.Background()); err != nil {
		t.Fatal(err)
	}
	data, err := clnt.FetchPsiphonConfig(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var config interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatal(err)
	}
}

func TestUnitFetchPsiphonConfigNotRegistered(t *testing.T) {
	clnt := newclient()
	state := probeservices.State{
		// Explicitly empty so the test is more clear
	}
	if err := clnt.StateFile.Set(state); err != nil {
		t.Fatal(err)
	}
	data, err := clnt.FetchPsiphonConfig(context.Background())
	if err == nil {
		t.Fatal("expected an error here")
	}
	if data != nil {
		t.Fatal("expected nil data here")
	}
}
