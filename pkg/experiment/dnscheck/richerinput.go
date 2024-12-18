package dnscheck

import (
	"context"

	"github.com/ooni/probe-engine/pkg/experimentconfig"
	"github.com/ooni/probe-engine/pkg/model"
	"github.com/ooni/probe-engine/pkg/reflectx"
	"github.com/ooni/probe-engine/pkg/targetloading"
)

// Target is a richer-input target that this experiment should measure.
type Target struct {
	// Config contains the configuration.
	Config *Config

	// URL is the input URL.
	URL string
}

var _ model.ExperimentTarget = &Target{}

// Category implements [model.ExperimentTarget].
func (t *Target) Category() string {
	return model.DefaultCategoryCode
}

// Country implements [model.ExperimentTarget].
func (t *Target) Country() string {
	return model.DefaultCountryCode
}

// Input implements [model.ExperimentTarget].
func (t *Target) Input() string {
	return t.URL
}

// Options implements [model.ExperimentTarget].
func (t *Target) Options() []string {
	return experimentconfig.DefaultOptionsSerializer(t.Config)
}

// String implements [model.ExperimentTarget].
func (t *Target) String() string {
	return t.URL
}

// NewLoader constructs a new [model.ExperimentTargerLoader] instance.
//
// This function PANICS if options is not an instance of [*dnscheck.Config].
func NewLoader(loader *targetloading.Loader, gopts any) model.ExperimentTargetLoader {
	// Panic if we cannot convert the options to the expected type.
	//
	// We do not expect a panic here because the type is managed by the registry package.
	options := gopts.(*Config)

	// Construct the proper loader instance.
	return &targetLoader{
		defaultInput: defaultInput,
		loader:       loader,
		options:      options,
	}
}

// targetLoader loads targets for this experiment.
type targetLoader struct {
	defaultInput []model.ExperimentTarget
	loader       *targetloading.Loader
	options      *Config
}

// Load implements model.ExperimentTargetLoader.
func (tl *targetLoader) Load(ctx context.Context) ([]model.ExperimentTarget, error) {
	// If inputs and files are all empty and there are no options, let's use the backend
	if len(tl.loader.StaticInputs) <= 0 && len(tl.loader.SourceFiles) <= 0 &&
		reflectx.StructOrStructPtrIsZero(tl.options) {
		return tl.loadFromBackend(ctx)
	}

	// Otherwise, attempt to load the static inputs from CLI and files
	inputs, err := targetloading.LoadStatic(tl.loader)

	// Handle the case where we couldn't
	if err != nil {
		return nil, err
	}

	// Build the list of targets that we should measure.
	var targets []model.ExperimentTarget
	for _, input := range inputs {
		targets = append(targets, &Target{
			Config: tl.options,
			URL:    input,
		})
	}
	return targets, nil
}

var defaultInput = []model.ExperimentTarget{
	//
	// https://dns.google/dns-query
	//
	// Measure HTTP/3 first and then HTTP/2 (see https://github.com/ooni/probe/issues/2675).
	//
	// Make sure we include the typical IP addresses for the domain.
	//
	&Target{
		URL: "https://dns.google/dns-query",
		Config: &Config{
			HTTP3Enabled: true,
			DefaultAddrs: "8.8.8.8 8.8.4.4",
		},
	},
	&Target{
		URL: "https://dns.google/dns-query",
		Config: &Config{
			DefaultAddrs: "8.8.8.8 8.8.4.4",
		},
	},
	&Target{
		URL: "https://cloudflare-dns.com/dns-query",
		Config: &Config{
			HTTP3Enabled: true,
			DefaultAddrs: "1.1.1.1 1.0.0.1",
		},
	},
	&Target{
		URL: "https://cloudflare-dns.com/dns-query",
		Config: &Config{
			DefaultAddrs: "1.1.1.1 1.0.0.1",
		},
	},
	&Target{
		URL: "https://dns.quad9.net/dns-query",
		Config: &Config{
			HTTP3Enabled: true,
			DefaultAddrs: "9.9.9.9",
		},
	},
	&Target{
		URL: "https://dns.quad9.net/dns-query",
		Config: &Config{
			DefaultAddrs: "9.9.9.9",
		},
	},
	&Target{
		URL: "https://dns.adguard.com/dns-query",
		Config: &Config{
			HTTP3Enabled: true,
		},
	},
	&Target{
		URL:    "https://dns.adguard.com/dns-query",
		Config: &Config{},
	},
	&Target{
		URL: "https://dns.alidns.com/dns-query",
		Config: &Config{
			HTTP3Enabled: true,
		},
	},
	&Target{
		URL:    "https://dns.alidns.com/dns-query",
		Config: &Config{},
	},
	&Target{
		URL: "https://doh.opendns.com/dns-query",
		Config: &Config{
			HTTP3Enabled: true,
		},
	},
	&Target{
		URL:    "https://doh.opendns.com/dns-query",
		Config: &Config{},
	},
	&Target{
		URL: "https://dns.nextdns.io/dns-query",
		Config: &Config{
			HTTP3Enabled: true,
		},
	},
	&Target{
		URL:    "https://dns.nextdns.io/dns-query",
		Config: &Config{},
	},

	&Target{
		URL: "https://dns.switch.ch/dns-query",
		Config: &Config{
			HTTP3Enabled: true,
		},
	},
	&Target{
		URL:    "https://dns.switch.ch/dns-query",
		Config: &Config{},
	},
}

// extendedInput is an extended input target list for dnscheck.
// TODO(decfox): we should have a flag to return the extended list in special cases
// while using the default list for normal runs.
//
//lint:ignore U1000 ignore unused var
var extendedInput = []model.ExperimentTarget{
	//
	// https://dns.google/dns-query
	//
	// Measure HTTP/3 first and then HTTP/2 (see https://github.com/ooni/probe/issues/2675).
	//
	// Make sure we include the typical IP addresses for the domain.
	//
	&Target{
		URL: "https://dns.google/dns-query",
		Config: &Config{
			HTTP3Enabled: true,
			DefaultAddrs: "8.8.8.8 8.8.4.4",
		},
	},
	&Target{
		URL: "https://dns.google/dns-query",
		Config: &Config{
			DefaultAddrs: "8.8.8.8 8.8.4.4",
		},
	},
	&Target{
		URL: "https://cloudflare-dns.com/dns-query",
		Config: &Config{
			HTTP3Enabled: true,
			DefaultAddrs: "1.1.1.1 1.0.0.1",
		},
	},
	&Target{
		URL: "https://cloudflare-dns.com/dns-query",
		Config: &Config{
			DefaultAddrs: "1.1.1.1 1.0.0.1",
		},
	},
	&Target{
		URL: "https://dns.quad9.net/dns-query",
		Config: &Config{
			HTTP3Enabled: true,
			DefaultAddrs: "9.9.9.9",
		},
	},
	&Target{
		URL: "https://dns.quad9.net/dns-query",
		Config: &Config{
			DefaultAddrs: "9.9.9.9",
		},
	},
	&Target{
		URL: "https://family.cloudflare-dns.com/dns-query",
		Config: &Config{
			HTTP3Enabled: true,
		},
	},
	&Target{
		URL:    "https://family.cloudflare-dns.com/dns-query",
		Config: &Config{},
	},
	&Target{
		URL: "https://dns11.quad9.net/dns-query",
		Config: &Config{
			HTTP3Enabled: true,
		},
	},
	&Target{
		URL:    "https://dns11.quad9.net/dns-query",
		Config: &Config{},
	},
	&Target{
		URL: "https://dns9.quad9.net/dns-query",
		Config: &Config{
			HTTP3Enabled: true,
		},
	},
	&Target{
		URL:    "https://dns9.quad9.net/dns-query",
		Config: &Config{},
	},
	&Target{
		URL: "https://dns12.quad9.net/dns-query",
		Config: &Config{
			HTTP3Enabled: true,
		},
	},
	&Target{
		URL:    "https://dns12.quad9.net/dns-query",
		Config: &Config{},
	},
	&Target{
		URL: "https://1dot1dot1dot1.cloudflare-dns.com/dns-query",
		Config: &Config{
			HTTP3Enabled: true,
		},
	},
	&Target{
		URL:    "https://1dot1dot1dot1.cloudflare-dns.com/dns-query",
		Config: &Config{},
	},
	&Target{
		URL: "https://dns.adguard.com/dns-query",
		Config: &Config{
			HTTP3Enabled: true,
		},
	},
	&Target{
		URL:    "https://dns.adguard.com/dns-query",
		Config: &Config{},
	},
	&Target{
		URL: "https://dns-family.adguard.com/dns-query",
		Config: &Config{
			HTTP3Enabled: true,
		},
	},
	&Target{
		URL:    "https://dns-family.adguard.com/dns-query",
		Config: &Config{},
	},
	&Target{
		URL: "https://dns.cloudflare.com/dns-query",
		Config: &Config{
			HTTP3Enabled: true,
		},
	},
	&Target{
		URL:    "https://dns.cloudflare.com/dns-query",
		Config: &Config{},
	},
	&Target{
		URL: "https://adblock.doh.mullvad.net/dns-query",
		Config: &Config{
			HTTP3Enabled: true,
		},
	},
	&Target{
		URL:    "https://adblock.doh.mullvad.net/dns-query",
		Config: &Config{},
	},
	&Target{
		URL: "https://dns.alidns.com/dns-query",
		Config: &Config{
			HTTP3Enabled: true,
		},
	},
	&Target{
		URL:    "https://dns.alidns.com/dns-query",
		Config: &Config{},
	},
	&Target{
		URL: "https://doh.opendns.com/dns-query",
		Config: &Config{
			HTTP3Enabled: true,
		},
	},
	&Target{
		URL:    "https://doh.opendns.com/dns-query",
		Config: &Config{},
	},
	&Target{
		URL: "https://dns.nextdns.io/dns-query",
		Config: &Config{
			HTTP3Enabled: true,
		},
	},
	&Target{
		URL:    "https://dns.nextdns.io/dns-query",
		Config: &Config{},
	},
	&Target{
		URL: "https://dns10.quad9.net/dns-query",
		Config: &Config{
			HTTP3Enabled: true,
		},
	},
	&Target{
		URL:    "https://dns10.quad9.net/dns-query",
		Config: &Config{},
	},
	&Target{
		URL: "https://security.cloudflare-dns.com/dns-query",
		Config: &Config{
			HTTP3Enabled: true,
		},
	},
	&Target{
		URL:    "https://security.cloudflare-dns.com/dns-query",
		Config: &Config{},
	},
	&Target{
		URL: "https://dns.switch.ch/dns-query",
		Config: &Config{
			HTTP3Enabled: true,
		},
	},
	&Target{
		URL:    "https://dns.switch.ch/dns-query",
		Config: &Config{},
	},
	// TODO(DecFox): implement a backend API to serve this list as dnscheck inputs
}

func (tl *targetLoader) loadFromBackend(_ context.Context) ([]model.ExperimentTarget, error) {
	// TODO(https://github.com/ooni/probe/issues/1390): serve DNSCheck
	// inputs using richer input (aka check-in v2).
	return defaultInput, nil
}
