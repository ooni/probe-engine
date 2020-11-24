package engine

import (
	"testing"

	"github.com/ooni/probe-engine/model"
)

// TODO(bassosimone):
//
// 1. this file should be renamed/refactored
//
// 2. we can consider zapping a bunch of MaybeFoo methods

func TestSessionMaybeProbeIPWorksAsIntended(t *testing.T) {
	sess := &Session{location: &model.LocationInfo{
		ProbeIP: "8.8.8.8",
	}}
	out := sess.MaybeProbeIP()
	if out != model.DefaultProbeIP {
		t.Fatal("not the value we expected")
	}
}

func TestSessionMaybeProbeASNWorksAsIntended(t *testing.T) {
	sess := &Session{location: &model.LocationInfo{
		ASN: 30722,
	}}
	out := sess.MaybeProbeASN()
	if out != 30722 {
		t.Fatal("not the value we expected")
	}
}

func TestSessionMaybeProbeASNStringWorksAsIntended(t *testing.T) {
	sess := &Session{location: &model.LocationInfo{
		ASN: 30722,
	}}
	out := sess.MaybeProbeASNString()
	if out != "AS30722" {
		t.Fatal("not the value we expected")
	}
}

func TestSessionMaybeProbeCCWorksAsIntended(t *testing.T) {
	sess := &Session{location: &model.LocationInfo{
		CountryCode: "IT",
	}}
	out := sess.MaybeProbeCC()
	if out != "IT" {
		t.Fatal("not the value we expected")
	}
}

func TestSessionMaybeProbeNetworkNameWorksAsIntended(t *testing.T) {
	sess := &Session{location: &model.LocationInfo{
		NetworkName: "Vodafone Italia",
	}}
	out := sess.MaybeProbeNetworkName()
	if out != "Vodafone Italia" {
		t.Fatal("not the value we expected")
	}
}

func TestSessionMaybeResolverIPWorksAsIntended(t *testing.T) {
	sess := &Session{location: &model.LocationInfo{
		ResolverIP: "9.9.9.9",
	}}
	out := sess.MaybeResolverIP()
	if out != "9.9.9.9" {
		t.Fatal("not the value we expected")
	}
}

func TestSessionMaybeResolverASNWorksAsIntended(t *testing.T) {
	sess := &Session{location: &model.LocationInfo{
		ResolverASN: 44,
	}}
	out := sess.MaybeResolverASN()
	if out != 44 {
		t.Fatal("not the value we expected")
	}
}

func TestSessionMaybeResolverASNStringWorksAsIntended(t *testing.T) {
	sess := &Session{location: &model.LocationInfo{
		ResolverASN: 44,
	}}
	out := sess.MaybeResolverASNString()
	if out != "AS44" {
		t.Fatal("not the value we expected")
	}
}

func TestSessionMaybeResolverNetworkNameWorksAsIntended(t *testing.T) {
	sess := &Session{location: &model.LocationInfo{
		ResolverNetworkName: "Google LLC",
	}}
	out := sess.MaybeResolverNetworkName()
	if out != "Google LLC" {
		t.Fatal("not the value we expected")
	}
}
