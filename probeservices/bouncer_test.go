package probeservices_test

import (
	"context"
	"testing"

	"github.com/apex/log"
)

func TestGetCollectors(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	collectors, err := makeClient().GetCollectors(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(collectors) <= 1 {
		t.Fatal("no returned collectors?!")
	}
}

func TestGetTestHelpers(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	testhelpers, err := makeClient().GetTestHelpers(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(testhelpers) <= 1 {
		t.Fatal("no returned test helpers?!")
	}
}
