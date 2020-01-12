package tor

import (
	"context"
	"testing"
)

func TestIntegration(t *testing.T) {
	targets, err := Query(context.Background(), Config{})
	if err != nil {
		t.Fatal(err)
	}
	if targets == nil {
		t.Fatal("expected non nil targets")
	}
}
