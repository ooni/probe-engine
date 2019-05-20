package ndt7_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/nettest/ndt7"
	"github.com/ooni/probe-engine/session"
)

const (
	softwareName    = "ooniprobe-example"
	softwareVersion = "0.0.1"
)

func TestIntegration(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	ctx := context.Background()

	sess := session.New(log.Log, softwareName, softwareVersion)
	sess.WorkDir = "../../testdata"
	if err := sess.FetchResourcesIdempotent(ctx); err != nil {
		t.Fatal(err)
	}
	if err := sess.LookupCollectors(ctx); err != nil {
		t.Fatal(err)
	}
	if err := sess.LookupTestHelpers(ctx); err != nil {
		t.Fatal(err)
	}
	if err := sess.LookupProbeIP(ctx); err != nil {
		t.Fatal(err)
	}
	if err := sess.LookupProbeASN(sess.ASNDatabasePath()); err != nil {
		t.Fatal(err)
	}
	if err := sess.LookupProbeCC(sess.CountryDatabasePath()); err != nil {
		t.Fatal(err)
	}
	if err := sess.LookupProbeNetworkName(sess.ASNDatabasePath()); err != nil {
		t.Fatal(err)
	}
	if err := sess.LookupResolverIP(ctx); err != nil {
		t.Fatal(err)
	}

	nt := ndt7.NewNettest(sess)
	if err := nt.OpenReport(ctx); err != nil {
		t.Fatal(err)
	}
	defer nt.CloseReport(ctx)

	measurement := nt.NewMeasurement("")
	if err := ndt7.Run(ctx, &measurement, func(event ndt7.Event) {
		data, err := json.Marshal(event)
		if err != nil {
			panic(err) // should not happen
		}
		log.Debug(string(data))
	}); err != nil {
		t.Fatal(err)
	}
	if err := nt.SubmitMeasurement(ctx, &measurement); err != nil {
		t.Fatal(err)
	}
}
