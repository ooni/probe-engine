package engine

import (
	"context"
	"errors"
	"testing"

	"github.com/ooni/probe-engine/model"
)

type locationLookupResourceUpdater struct {
	err error
}

func (c locationLookupResourceUpdater) MaybeUpdateResources(ctx context.Context) error {
	return c.err
}

func TestLocationLookupCannotUpdateResources(t *testing.T) {
	expected := errors.New("mocked error")
	op := LocationLookup{
		ResourceUpdater: locationLookupResourceUpdater{err: expected},
	}
	ctx := context.Background()
	out, err := op.Do(ctx)
	if !errors.Is(err, expected) {
		t.Fatalf("not the error we expected: %+v", err)
	}
	if out.ASN != model.DefaultProbeASN {
		t.Fatal("invalid ASN value")
	}
	if out.CountryCode != model.DefaultProbeCC {
		t.Fatal("invalid CountryCode value")
	}
	if out.NetworkName != model.DefaultProbeNetworkName {
		t.Fatal("invalid NetworkName value")
	}
	if out.ProbeIP != model.DefaultProbeIP {
		t.Fatal("invalid ProbeIP value")
	}
	if out.ResolverASN != model.DefaultResolverASN {
		t.Fatal("invalid ResolverASN value")
	}
	if out.ResolverIP != model.DefaultResolverIP {
		t.Fatal("invalid ResolverIP value")
	}
	if out.ResolverNetworkName != model.DefaultResolverNetworkName {
		t.Fatal("invalid ResolverNetworkName value")
	}
}

type locationLookupProbeIPLookupper struct {
	ip  string
	err error
}

func (c locationLookupProbeIPLookupper) LookupProbeIP(ctx context.Context) (string, error) {
	return c.ip, c.err
}

func TestLocationLookupCannotLookupProbeIP(t *testing.T) {
	expected := errors.New("mocked error")
	op := LocationLookup{
		ResourceUpdater:  locationLookupResourceUpdater{},
		ProbeIPLookupper: locationLookupProbeIPLookupper{err: expected},
	}
	ctx := context.Background()
	out, err := op.Do(ctx)
	if !errors.Is(err, expected) {
		t.Fatalf("not the error we expected: %+v", err)
	}
	if out.ASN != model.DefaultProbeASN {
		t.Fatal("invalid ASN value")
	}
	if out.CountryCode != model.DefaultProbeCC {
		t.Fatal("invalid CountryCode value")
	}
	if out.NetworkName != model.DefaultProbeNetworkName {
		t.Fatal("invalid NetworkName value")
	}
	if out.ProbeIP != model.DefaultProbeIP {
		t.Fatal("invalid ProbeIP value")
	}
	if out.ResolverASN != model.DefaultResolverASN {
		t.Fatal("invalid ResolverASN value")
	}
	if out.ResolverIP != model.DefaultResolverIP {
		t.Fatal("invalid ResolverIP value")
	}
	if out.ResolverNetworkName != model.DefaultResolverNetworkName {
		t.Fatal("invalid ResolverNetworkName value")
	}
}

type locationLookupASNLookupper struct {
	err  error
	asn  uint
	name string
}

func (c locationLookupASNLookupper) LookupASN(path string, ip string) (uint, string, error) {
	return c.asn, c.name, c.err
}

type locationLookupPathsProvider struct {
	asnDatabasePath     string
	countryDatabasePath string
}

func (c locationLookupPathsProvider) ASNDatabasePath() string {
	return c.asnDatabasePath
}

func (c locationLookupPathsProvider) CountryDatabasePath() string {
	return c.countryDatabasePath
}

func TestLocationLookupCannotLookupProbeASN(t *testing.T) {
	expected := errors.New("mocked error")
	op := LocationLookup{
		ResourceUpdater:   locationLookupResourceUpdater{},
		ProbeIPLookupper:  locationLookupProbeIPLookupper{ip: "1.2.3.4"},
		ProbeASNLookupper: locationLookupASNLookupper{err: expected},
		PathsProvider:     locationLookupPathsProvider{},
	}
	ctx := context.Background()
	out, err := op.Do(ctx)
	if !errors.Is(err, expected) {
		t.Fatalf("not the error we expected: %+v", err)
	}
	if out.ASN != model.DefaultProbeASN {
		t.Fatal("invalid ASN value")
	}
	if out.CountryCode != model.DefaultProbeCC {
		t.Fatal("invalid CountryCode value")
	}
	if out.NetworkName != model.DefaultProbeNetworkName {
		t.Fatal("invalid NetworkName value")
	}
	if out.ProbeIP != "1.2.3.4" {
		t.Fatal("invalid ProbeIP value")
	}
	if out.ResolverASN != model.DefaultResolverASN {
		t.Fatal("invalid ResolverASN value")
	}
	if out.ResolverIP != model.DefaultResolverIP {
		t.Fatal("invalid ResolverIP value")
	}
	if out.ResolverNetworkName != model.DefaultResolverNetworkName {
		t.Fatal("invalid ResolverNetworkName value")
	}
}

type locationLookupCCLookupper struct {
	err error
	cc  string
}

func (c locationLookupCCLookupper) LookupCC(path string, ip string) (string, error) {
	return c.cc, c.err
}

func TestLocationLookupCannotLookupProbeCC(t *testing.T) {
	expected := errors.New("mocked error")
	op := LocationLookup{
		ResourceUpdater:   locationLookupResourceUpdater{},
		ProbeIPLookupper:  locationLookupProbeIPLookupper{ip: "1.2.3.4"},
		ProbeASNLookupper: locationLookupASNLookupper{asn: 1234, name: "1234.com"},
		CountryLookupper:  locationLookupCCLookupper{cc: "US", err: expected},
		PathsProvider:     locationLookupPathsProvider{},
	}
	ctx := context.Background()
	out, err := op.Do(ctx)
	if !errors.Is(err, expected) {
		t.Fatalf("not the error we expected: %+v", err)
	}
	if out.ASN != 1234 {
		t.Fatal("invalid ASN value")
	}
	if out.CountryCode != model.DefaultProbeCC {
		t.Fatal("invalid CountryCode value")
	}
	if out.NetworkName != "1234.com" {
		t.Fatal("invalid NetworkName value")
	}
	if out.ProbeIP != "1.2.3.4" {
		t.Fatal("invalid ProbeIP value")
	}
	if out.ResolverASN != model.DefaultResolverASN {
		t.Fatal("invalid ResolverASN value")
	}
	if out.ResolverIP != model.DefaultResolverIP {
		t.Fatal("invalid ResolverIP value")
	}
	if out.ResolverNetworkName != model.DefaultResolverNetworkName {
		t.Fatal("invalid ResolverNetworkName value")
	}
}

type locationLookupResolverIPLookupper struct {
	ip  string
	err error
}

func (c locationLookupResolverIPLookupper) LookupResolverIP(ctx context.Context) (string, error) {
	return c.ip, c.err
}

func TestLocationLookupCannotLookupResolverIP(t *testing.T) {
	expected := errors.New("mocked error")
	op := LocationLookup{
		ResourceUpdater:      locationLookupResourceUpdater{},
		ProbeIPLookupper:     locationLookupProbeIPLookupper{ip: "1.2.3.4"},
		ProbeASNLookupper:    locationLookupASNLookupper{asn: 1234, name: "1234.com"},
		CountryLookupper:     locationLookupCCLookupper{cc: "IT"},
		PathsProvider:        locationLookupPathsProvider{},
		ResolverIPLookupper:  locationLookupResolverIPLookupper{err: expected},
		EnableResolverLookup: true,
	}
	ctx := context.Background()
	out, err := op.Do(ctx)
	if err != nil {
		t.Fatalf("not the error we expected: %+v", err)
	}
	if out.ASN != 1234 {
		t.Fatal("invalid ASN value")
	}
	if out.CountryCode != "IT" {
		t.Fatal("invalid CountryCode value")
	}
	if out.NetworkName != "1234.com" {
		t.Fatal("invalid NetworkName value")
	}
	if out.ProbeIP != "1.2.3.4" {
		t.Fatal("invalid ProbeIP value")
	}
	if out.DidResolverLookup != true {
		t.Fatal("invalid DidResolverLookup value")
	}
	if out.ResolverASN != model.DefaultResolverASN {
		t.Fatal("invalid ResolverASN value")
	}
	if out.ResolverIP != model.DefaultResolverIP {
		t.Fatal("invalid ResolverIP value")
	}
	if out.ResolverNetworkName != model.DefaultResolverNetworkName {
		t.Fatal("invalid ResolverNetworkName value")
	}
}

func TestLocationLookupCannotLookupResolverNetworkName(t *testing.T) {
	expected := errors.New("mocked error")
	op := LocationLookup{
		ResourceUpdater:      locationLookupResourceUpdater{},
		ProbeIPLookupper:     locationLookupProbeIPLookupper{ip: "1.2.3.4"},
		ProbeASNLookupper:    locationLookupASNLookupper{asn: 1234, name: "1234.com"},
		CountryLookupper:     locationLookupCCLookupper{cc: "IT"},
		PathsProvider:        locationLookupPathsProvider{},
		ResolverIPLookupper:  locationLookupResolverIPLookupper{ip: "4.3.2.1"},
		ResolverASNLookupper: locationLookupASNLookupper{err: expected},
		EnableResolverLookup: true,
	}
	ctx := context.Background()
	out, err := op.Do(ctx)
	if err != nil {
		t.Fatalf("not the error we expected: %+v", err)
	}
	if out.ASN != 1234 {
		t.Fatal("invalid ASN value")
	}
	if out.CountryCode != "IT" {
		t.Fatal("invalid CountryCode value")
	}
	if out.NetworkName != "1234.com" {
		t.Fatal("invalid NetworkName value")
	}
	if out.ProbeIP != "1.2.3.4" {
		t.Fatal("invalid ProbeIP value")
	}
	if out.DidResolverLookup != true {
		t.Fatal("invalid DidResolverLookup value")
	}
	if out.ResolverASN != model.DefaultResolverASN {
		t.Fatalf("invalid ResolverASN value: %+v", out.ResolverASN)
	}
	if out.ResolverIP != "4.3.2.1" {
		t.Fatalf("invalid ResolverIP value: %+v", out.ResolverIP)
	}
	if out.ResolverNetworkName != model.DefaultResolverNetworkName {
		t.Fatal("invalid ResolverNetworkName value")
	}
}

func TestLocationLookupSuccessWithResolverLookup(t *testing.T) {
	op := LocationLookup{
		ResourceUpdater:      locationLookupResourceUpdater{},
		ProbeIPLookupper:     locationLookupProbeIPLookupper{ip: "1.2.3.4"},
		ProbeASNLookupper:    locationLookupASNLookupper{asn: 1234, name: "1234.com"},
		CountryLookupper:     locationLookupCCLookupper{cc: "IT"},
		PathsProvider:        locationLookupPathsProvider{},
		ResolverIPLookupper:  locationLookupResolverIPLookupper{ip: "4.3.2.1"},
		ResolverASNLookupper: locationLookupASNLookupper{asn: 4321, name: "4321.com"},
		EnableResolverLookup: true,
	}
	ctx := context.Background()
	out, err := op.Do(ctx)
	if err != nil {
		t.Fatalf("not the error we expected: %+v", err)
	}
	if out.ASN != 1234 {
		t.Fatal("invalid ASN value")
	}
	if out.CountryCode != "IT" {
		t.Fatal("invalid CountryCode value")
	}
	if out.NetworkName != "1234.com" {
		t.Fatal("invalid NetworkName value")
	}
	if out.ProbeIP != "1.2.3.4" {
		t.Fatal("invalid ProbeIP value")
	}
	if out.DidResolverLookup != true {
		t.Fatal("invalid DidResolverLookup value")
	}
	if out.ResolverASN != 4321 {
		t.Fatalf("invalid ResolverASN value: %+v", out.ResolverASN)
	}
	if out.ResolverIP != "4.3.2.1" {
		t.Fatalf("invalid ResolverIP value: %+v", out.ResolverIP)
	}
	if out.ResolverNetworkName != "4321.com" {
		t.Fatal("invalid ResolverNetworkName value")
	}
}

func TestLocationLookupSuccessWithoutResolverLookup(t *testing.T) {
	op := LocationLookup{
		ResourceUpdater:      locationLookupResourceUpdater{},
		ProbeIPLookupper:     locationLookupProbeIPLookupper{ip: "1.2.3.4"},
		ProbeASNLookupper:    locationLookupASNLookupper{asn: 1234, name: "1234.com"},
		CountryLookupper:     locationLookupCCLookupper{cc: "IT"},
		PathsProvider:        locationLookupPathsProvider{},
		ResolverIPLookupper:  locationLookupResolverIPLookupper{ip: "4.3.2.1"},
		ResolverASNLookupper: locationLookupASNLookupper{asn: 4321, name: "4321.com"},
	}
	ctx := context.Background()
	out, err := op.Do(ctx)
	if err != nil {
		t.Fatalf("not the error we expected: %+v", err)
	}
	if out.ASN != 1234 {
		t.Fatal("invalid ASN value")
	}
	if out.CountryCode != "IT" {
		t.Fatal("invalid CountryCode value")
	}
	if out.NetworkName != "1234.com" {
		t.Fatal("invalid NetworkName value")
	}
	if out.ProbeIP != "1.2.3.4" {
		t.Fatal("invalid ProbeIP value")
	}
	if out.DidResolverLookup != false {
		t.Fatal("invalid DidResolverLookup value")
	}
	if out.ResolverASN != model.DefaultResolverASN {
		t.Fatalf("invalid ResolverASN value: %+v", out.ResolverASN)
	}
	if out.ResolverIP != model.DefaultResolverIP {
		t.Fatalf("invalid ResolverIP value: %+v", out.ResolverIP)
	}
	if out.ResolverNetworkName != model.DefaultResolverNetworkName {
		t.Fatal("invalid ResolverNetworkName value")
	}
}
