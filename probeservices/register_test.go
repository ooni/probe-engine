package probeservices_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/ooni/probe-engine/probeservices"
	"github.com/ooni/probe-engine/probeservices/testorchestra"
)

func TestMaybeRegister(t *testing.T) {
	t.Run("when metadata is not valid", func(t *testing.T) {
		clnt := newclient()
		ctx := context.Background()
		var metadata probeservices.Metadata
		err := clnt.MaybeRegister(ctx, metadata)
		if !errors.Is(err, probeservices.ErrInvalidMetadata) {
			t.Fatal("expected an error here")
		}
	})
	t.Run("when we have already registered", func(t *testing.T) {
		clnt := newclient()
		state := probeservices.State{
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
		clnt.BaseURL = "\t\t\t" // makes it fail
		ctx := context.Background()
		metadata := testorchestra.MetadataFixture()
		err := clnt.MaybeRegister(ctx, metadata)
		if err == nil || !strings.HasSuffix(err.Error(), "invalid control character in URL") {
			t.Fatal("expected an error here")
		}
	})
}

func TestMaybeRegisterIdempotent(t *testing.T) {
	clnt := newclient()
	ctx := context.Background()
	metadata := testorchestra.MetadataFixture()
	if err := clnt.MaybeRegister(ctx, metadata); err != nil {
		t.Fatal(err)
	}
	if err := clnt.MaybeRegister(ctx, metadata); err != nil {
		t.Fatal(err)
	}
	if clnt.RegisterCalls.Load() != 1 {
		t.Fatal("called register API too many times")
	}
}
