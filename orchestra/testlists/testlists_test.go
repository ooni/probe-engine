package testlists_test

import (
	"context"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/orchestra/testlists"
	"github.com/ooni/probe-engine/session"
)

func TestIntegration(t *testing.T) {
	sess := session.New(
		log.Log,
		"ooniprobe-engine",
		"0.1.0",
		"../../testdata/",
		nil, nil,
	)
	client := testlists.NewClient(sess)
	client.SetEnabledCategories([]string{"NEWS", "CULTR"})
	log.SetLevel(log.DebugLevel)
	urls, err := client.Do(context.Background(), "IT", 128)
	if err != nil {
		t.Fatal(err)
	}
	log.Infof("%+v", urls)
}
