package orchestra_test

import (
	"context"
	"testing"
	"time"

	"github.com/ooni/probe-engine/internal/mockable"
	"github.com/ooni/probe-engine/internal/orchestra"
)

func TestUnitMaybeLogin(t *testing.T) {
	t.Run("when we already have a token", func(t *testing.T) {
		clnt := newclient()
		state := orchestra.State{
			Expire: time.Now().Add(time.Hour),
			Token:  "xx-xxx-x-xxxx",
		}
		if err := clnt.StateFile.Set(state); err != nil {
			t.Fatal(err)
		}
		ctx := context.Background()
		if err := clnt.MaybeLogin(ctx); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("when we have already registered", func(t *testing.T) {
		clnt := newclient()
		state := orchestra.State{
			// Explicitly empty to clarify what this test does
		}
		if err := clnt.StateFile.Set(state); err != nil {
			t.Fatal(err)
		}
		ctx := context.Background()
		if err := clnt.MaybeLogin(ctx); err == nil {
			t.Fatal("expected an error here")
		}
	})
	t.Run("when the API call fails", func(t *testing.T) {
		clnt := newclient()
		clnt.BaseURL = "\t\t\t"
		state := orchestra.State{
			ClientID: "xx-xxx-x-xxxx",
			Password: "xx",
		}
		if err := clnt.StateFile.Set(state); err != nil {
			t.Fatal(err)
		}
		ctx := context.Background()
		if err := clnt.MaybeLogin(ctx); err == nil {
			t.Fatal("expected an error here")
		}
	})
}

func TestIntegrationMaybeLoginIdempotent(t *testing.T) {
	clnt := newclient()
	ctx := context.Background()
	metadata := mockable.OrchestraMetadataFixture()
	if err := clnt.MaybeRegister(ctx, metadata); err != nil {
		t.Fatal(err)
	}
	if err := clnt.MaybeLogin(ctx); err != nil {
		t.Fatal(err)
	}
	if err := clnt.MaybeLogin(ctx); err != nil {
		t.Fatal(err)
	}
	if clnt.LoginCalls.Load() != 1 {
		t.Fatal("called login API too many times")
	}
}
