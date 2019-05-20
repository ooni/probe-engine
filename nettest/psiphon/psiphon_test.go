package psiphon_test

import (
	"context"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/nettest/psiphon"
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

	nt := psiphon.NewNettest(sess)
	if err := nt.OpenReport(ctx); err != nil {
		t.Fatal(err)
	}
	defer nt.CloseReport(ctx)

	measurement := nt.NewMeasurement("")
	err := psiphon.Run(ctx, &measurement, psiphon.Config{
		ConfigFilePath: "../../testdata/psiphon_config.json",
		Logger:         sess.Logger,
		UserAgent:      sess.UserAgent(),
		WorkDir:        sess.WorkDir,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := nt.SubmitMeasurement(ctx, &measurement); err != nil {
		t.Fatal(err)
	}
}
