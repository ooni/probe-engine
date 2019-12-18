package orchestra

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/internal/kvstore"
	"github.com/ooni/probe-engine/internal/orchestra/metadata"
	"github.com/ooni/probe-engine/internal/orchestra/statefile"
	"github.com/ooni/probe-engine/internal/orchestra/testorchestra"
)

func TestIntegrationUpdate(t *testing.T) {
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
	if err := clnt.Update(
		context.Background(), testorchestra.MetadataFixture(),
	); err != nil {
		t.Fatal(err)
	}
}

func TestUnitMaybeRegister(t *testing.T) {
	t.Run("when metadata is not valid", func(t *testing.T) {
		clnt := newclient()
		ctx := context.Background()
		var metadata metadata.Metadata
		if err := clnt.MaybeRegister(ctx, metadata); err == nil {
			t.Fatal("expected an error here")
		}
	})
	t.Run("when we have already registered", func(t *testing.T) {
		clnt := newclient()
		state := statefile.State{
			ClientID: "xx-xxx-x-xxxx",
			Password: "xx",
		}
		if err := clnt.StateFile.Set(state); err != nil {
			t.Fatal(err)
		}
		ctx := context.Background()
		metadata := testorchestra.MetadataFixture()
		if err := clnt.MaybeRegister(ctx, metadata); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("when the API call fails", func(t *testing.T) {
		clnt := newclient()
		clnt.RegistryBaseURL = "\t\t\t"
		ctx := context.Background()
		metadata := testorchestra.MetadataFixture()
		if err := clnt.MaybeRegister(ctx, metadata); err == nil {
			t.Fatal("expected an error here")
		}
	})
}

func TestIntegrationMaybeRegisterIdempotent(t *testing.T) {
	clnt := newclient()
	ctx := context.Background()
	metadata := testorchestra.MetadataFixture()
	if err := clnt.MaybeRegister(ctx, metadata); err != nil {
		t.Fatal(err)
	}
	if err := clnt.MaybeRegister(ctx, metadata); err != nil {
		t.Fatal(err)
	}
	if clnt.registerCalls != 1 {
		t.Fatal("called register API too many times")
	}
}

func TestUnitMaybeLogin(t *testing.T) {
	t.Run("when we already have a token", func(t *testing.T) {
		clnt := newclient()
		state := statefile.State{
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
		state := statefile.State{
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
		clnt.RegistryBaseURL = "\t\t\t"
		state := statefile.State{
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
	metadata := testorchestra.MetadataFixture()
	if err := clnt.MaybeRegister(ctx, metadata); err != nil {
		t.Fatal(err)
	}
	if err := clnt.MaybeLogin(ctx); err != nil {
		t.Fatal(err)
	}
	if err := clnt.MaybeLogin(ctx); err != nil {
		t.Fatal(err)
	}
	if clnt.loginCalls != 1 {
		t.Fatal("called login API too many times")
	}
}

func TestUnitUpdate(t *testing.T) {
	t.Run("when metadata is not valid", func(t *testing.T) {
		clnt := newclient()
		ctx := context.Background()
		var metadata metadata.Metadata
		if err := clnt.Update(ctx, metadata); err == nil {
			t.Fatal("expected an error here")
		}
	})
	t.Run("when we have are not registered", func(t *testing.T) {
		clnt := newclient()
		state := statefile.State{
			// Explicitly empty so the test is more clear
		}
		if err := clnt.StateFile.Set(state); err != nil {
			t.Fatal(err)
		}
		ctx := context.Background()
		metadata := testorchestra.MetadataFixture()
		if err := clnt.Update(ctx, metadata); err == nil {
			t.Fatal("expected an error here")
		}
	})
	t.Run("when we are not logged in", func(t *testing.T) {
		clnt := newclient()
		state := statefile.State{
			ClientID: "xx-xxx-x-xxxx",
			Password: "xx",
		}
		if err := clnt.StateFile.Set(state); err != nil {
			t.Fatal(err)
		}
		ctx := context.Background()
		metadata := testorchestra.MetadataFixture()
		if err := clnt.Update(ctx, metadata); err == nil {
			t.Fatal("expected an error here")
		}
	})
	t.Run("when the API call fails", func(t *testing.T) {
		clnt := newclient()
		clnt.RegistryBaseURL = "\t\t\t"
		state := statefile.State{
			ClientID: "xx-xxx-x-xxxx",
			Expire:   time.Now().Add(time.Hour),
			Password: "xx",
			Token:    "xx-xxx-x-xxxx",
		}
		if err := clnt.StateFile.Set(state); err != nil {
			t.Fatal(err)
		}
		ctx := context.Background()
		metadata := testorchestra.MetadataFixture()
		if err := clnt.Update(ctx, metadata); err == nil {
			t.Fatal("expected an error here")
		}
	})
}

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

func TestUnitGetPsiphonConfigNotRegistered(t *testing.T) {
	clnt := newclient()
	state := statefile.State{
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

func newclient() *Client {
	clnt := NewClient(
		http.DefaultClient,
		log.Log,
		"miniooni/0.1.0-dev",
		statefile.New(kvstore.NewMemoryKeyValueStore()),
	)
	clnt.OrchestrateBaseURL = "https://ps-test.ooni.io"
	clnt.RegistryBaseURL = "https://ps-test.ooni.io"
	return clnt
}
