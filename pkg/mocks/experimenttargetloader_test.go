package mocks

import (
	"context"
	"errors"
	"testing"

	"github.com/ooni/probe-engine/pkg/model"
)

func TestExperimentInputLoader(t *testing.T) {
	t.Run("Load", func(t *testing.T) {
		expected := errors.New("mocked error")
		eil := &ExperimentTargetLoader{
			MockLoad: func(ctx context.Context) ([]model.ExperimentTarget, error) {
				return nil, expected
			},
		}
		out, err := eil.Load(context.Background())
		if !errors.Is(err, expected) {
			t.Fatal("unexpected err", err)
		}
		if len(out) > 0 {
			t.Fatal("unexpected length")
		}
	})
}
