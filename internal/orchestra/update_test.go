package orchestra_test

import (
	"context"
	"testing"
	"time"

	"github.com/ooni/probe-engine/internal/mockable"
	"github.com/ooni/probe-engine/internal/orchestra"
)

func TestIntegrationUpdate(t *testing.T) {
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
	if err := clnt.Update(
		context.Background(), mockable.OrchestraMetadataFixture(),
	); err != nil {
		t.Fatal(err)
	}
}

func TestUnitUpdate(t *testing.T) {
	t.Run("when metadata is not valid", func(t *testing.T) {
		clnt := newclient()
		ctx := context.Background()
		var metadata orchestra.Metadata
		if err := clnt.Update(ctx, metadata); err == nil {
			t.Fatal("expected an error here")
		}
	})
	t.Run("when we have are not registered", func(t *testing.T) {
		clnt := newclient()
		state := orchestra.State{
			// Explicitly empty so the test is more clear
		}
		if err := clnt.StateFile.Set(state); err != nil {
			t.Fatal(err)
		}
		ctx := context.Background()
		metadata := mockable.OrchestraMetadataFixture()
		if err := clnt.Update(ctx, metadata); err == nil {
			t.Fatal("expected an error here")
		}
	})
	t.Run("when we are not logged in", func(t *testing.T) {
		clnt := newclient()
		state := orchestra.State{
			ClientID: "xx-xxx-x-xxxx",
			Password: "xx",
		}
		if err := clnt.StateFile.Set(state); err != nil {
			t.Fatal(err)
		}
		ctx := context.Background()
		metadata := mockable.OrchestraMetadataFixture()
		if err := clnt.Update(ctx, metadata); err == nil {
			t.Fatal("expected an error here")
		}
	})
	t.Run("when the API call fails", func(t *testing.T) {
		clnt := newclient()
		clnt.BaseURL = "\t\t\t"
		state := orchestra.State{
			ClientID: "xx-xxx-x-xxxx",
			Expire:   time.Now().Add(time.Hour),
			Password: "xx",
			Token:    "xx-xxx-x-xxxx",
		}
		if err := clnt.StateFile.Set(state); err != nil {
			t.Fatal(err)
		}
		ctx := context.Background()
		metadata := mockable.OrchestraMetadataFixture()
		if err := clnt.Update(ctx, metadata); err == nil {
			t.Fatal("expected an error here")
		}
	})
}
