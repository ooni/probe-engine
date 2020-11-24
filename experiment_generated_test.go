package engine

import (
	"testing"

	"github.com/ooni/probe-engine/model"
)

// TODO(bassosimone): this file should be renamed/refactored.

func TestExperimentNewMeasurementProbeIPWorksAsIntended(t *testing.T) {
	sess := &Session{location: &model.LocationInfo{
		ProbeIP: "8.8.8.8",
	}}
	newExperiment := func() *Experiment {
		builder, err := sess.NewExperimentBuilder("example")
		if err != nil {
			t.Fatal(err)
		}
		return builder.NewExperiment()
	}
	exp := newExperiment()
	m := exp.newMeasurement("")
	if m.ProbeIP != model.DefaultProbeIP {
		t.Fatal("not the value we expected")
	}
}

func TestExperimentNewMeasurementProbeASNStringWorksAsIntended(t *testing.T) {
	sess := &Session{location: &model.LocationInfo{
		ASN: 30722,
	}}
	newExperiment := func() *Experiment {
		builder, err := sess.NewExperimentBuilder("example")
		if err != nil {
			t.Fatal(err)
		}
		return builder.NewExperiment()
	}
	exp := newExperiment()
	m := exp.newMeasurement("")
	if m.ProbeASN != "AS30722" {
		t.Fatal("not the value we expected")
	}
}

func TestExperimentOpenReportProbeASNStringWorksAsIntended(t *testing.T) {
	sess := &Session{location: &model.LocationInfo{
		ASN: 30722,
	}}
	newExperiment := func() *Experiment {
		builder, err := sess.NewExperimentBuilder("example")
		if err != nil {
			t.Fatal(err)
		}
		return builder.NewExperiment()
	}
	exp := newExperiment()
	rt := exp.newReportTemplate()
	if rt.ProbeASN != "AS30722" {
		t.Fatal("not the value we expected")
	}
}

func TestExperimentNewMeasurementProbeCCWorksAsIntended(t *testing.T) {
	sess := &Session{location: &model.LocationInfo{
		CountryCode: "IT",
	}}
	newExperiment := func() *Experiment {
		builder, err := sess.NewExperimentBuilder("example")
		if err != nil {
			t.Fatal(err)
		}
		return builder.NewExperiment()
	}
	exp := newExperiment()
	m := exp.newMeasurement("")
	if m.ProbeCC != "IT" {
		t.Fatal("not the value we expected")
	}
}

func TestExperimentOpenReportProbeCCWorksAsIntended(t *testing.T) {
	sess := &Session{location: &model.LocationInfo{
		CountryCode: "IT",
	}}
	newExperiment := func() *Experiment {
		builder, err := sess.NewExperimentBuilder("example")
		if err != nil {
			t.Fatal(err)
		}
		return builder.NewExperiment()
	}
	exp := newExperiment()
	rt := exp.newReportTemplate()
	if rt.ProbeCC != "IT" {
		t.Fatal("not the value we expected")
	}
}

func TestExperimentNewMeasurementProbeNetworkNameWorksAsIntended(t *testing.T) {
	sess := &Session{location: &model.LocationInfo{
		NetworkName: "Vodafone Italia",
	}}
	newExperiment := func() *Experiment {
		builder, err := sess.NewExperimentBuilder("example")
		if err != nil {
			t.Fatal(err)
		}
		return builder.NewExperiment()
	}
	exp := newExperiment()
	m := exp.newMeasurement("")
	if m.ProbeNetworkName != "Vodafone Italia" {
		t.Fatal("not the value we expected")
	}
}

func TestExperimentNewMeasurementResolverIPWorksAsIntended(t *testing.T) {
	sess := &Session{location: &model.LocationInfo{
		ResolverIP: "9.9.9.9",
	}}
	newExperiment := func() *Experiment {
		builder, err := sess.NewExperimentBuilder("example")
		if err != nil {
			t.Fatal(err)
		}
		return builder.NewExperiment()
	}
	exp := newExperiment()
	m := exp.newMeasurement("")
	if m.ResolverIP != "9.9.9.9" {
		t.Fatal("not the value we expected")
	}
}

func TestExperimentNewMeasurementResolverASNStringWorksAsIntended(t *testing.T) {
	sess := &Session{location: &model.LocationInfo{
		ResolverASN: 44,
	}}
	newExperiment := func() *Experiment {
		builder, err := sess.NewExperimentBuilder("example")
		if err != nil {
			t.Fatal(err)
		}
		return builder.NewExperiment()
	}
	exp := newExperiment()
	m := exp.newMeasurement("")
	if m.ResolverASN != "AS44" {
		t.Fatal("not the value we expected")
	}
}

func TestExperimentNewMeasurementResolverNetworkNameWorksAsIntended(t *testing.T) {
	sess := &Session{location: &model.LocationInfo{
		ResolverNetworkName: "Google LLC",
	}}
	newExperiment := func() *Experiment {
		builder, err := sess.NewExperimentBuilder("example")
		if err != nil {
			t.Fatal(err)
		}
		return builder.NewExperiment()
	}
	exp := newExperiment()
	m := exp.newMeasurement("")
	if m.ResolverNetworkName != "Google LLC" {
		t.Fatal("not the value we expected")
	}
}
