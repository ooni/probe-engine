package enginenetx

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/apex/log"
	"github.com/google/go-cmp/cmp"
	"github.com/ooni/probe-engine/pkg/kvstore"
	"github.com/ooni/probe-engine/pkg/netemx"
	"github.com/ooni/probe-engine/pkg/runtimex"
)

func TestStatsPolicyV2(t *testing.T) {
	// prepare the content of the stats
	twentyMinutesAgo := time.Now().Add(-20 * time.Minute)

	const bridgeAddress = netemx.AddressApiOONIIo

	expectTacticsStats := []*statsTactic{{
		CountStarted:               5,
		CountTCPConnectError:       0,
		CountTCPConnectInterrupt:   0,
		CountTLSHandshakeError:     0,
		CountTLSHandshakeInterrupt: 0,
		CountTLSVerificationError:  0,
		CountSuccess:               5, // this one always succeeds, so it should be there
		HistoTCPConnectError:       map[string]int64{},
		HistoTLSHandshakeError:     map[string]int64{},
		HistoTLSVerificationError:  map[string]int64{},
		LastUpdated:                twentyMinutesAgo,
		Tactic: &httpsDialerTactic{
			Address:        bridgeAddress,
			InitialDelay:   0,
			Port:           "443",
			SNI:            "www.repubblica.it",
			VerifyHostname: "api.ooni.io",
		},
	}, {
		CountStarted:               3,
		CountTCPConnectError:       0,
		CountTCPConnectInterrupt:   0,
		CountTLSHandshakeError:     1,
		CountTLSHandshakeInterrupt: 0,
		CountTLSVerificationError:  0,
		CountSuccess:               2, // this one sometimes succeded so it should be added
		HistoTCPConnectError:       map[string]int64{},
		HistoTLSHandshakeError:     map[string]int64{},
		HistoTLSVerificationError:  map[string]int64{},
		LastUpdated:                twentyMinutesAgo,
		Tactic: &httpsDialerTactic{
			Address:        bridgeAddress,
			InitialDelay:   0,
			Port:           "443",
			SNI:            "www.kernel.org",
			VerifyHostname: "api.ooni.io",
		},
	}, {
		CountStarted:               3,
		CountTCPConnectError:       0,
		CountTCPConnectInterrupt:   0,
		CountTLSHandshakeError:     3, // this one always failed, so should not be added
		CountTLSHandshakeInterrupt: 0,
		CountTLSVerificationError:  0,
		CountSuccess:               0,
		HistoTCPConnectError:       map[string]int64{},
		HistoTLSHandshakeError:     map[string]int64{},
		HistoTLSVerificationError:  map[string]int64{},
		LastUpdated:                twentyMinutesAgo,
		Tactic: &httpsDialerTactic{
			Address:        bridgeAddress,
			InitialDelay:   0,
			Port:           "443",
			SNI:            "theconversation.com",
			VerifyHostname: "api.ooni.io",
		},
	}, {
		CountStarted:               4,
		CountTCPConnectError:       0,
		CountTCPConnectInterrupt:   0,
		CountTLSHandshakeError:     0,
		CountTLSHandshakeInterrupt: 0,
		CountTLSVerificationError:  0,
		CountSuccess:               4,
		HistoTCPConnectError:       map[string]int64{},
		HistoTLSHandshakeError:     map[string]int64{},
		HistoTLSVerificationError:  map[string]int64{},
		LastUpdated:                twentyMinutesAgo,
		Tactic:                     nil, // the nil policy here should cause this entry to be filtered out
	}, {
		CountStarted:               0,
		CountTCPConnectError:       0,
		CountTCPConnectInterrupt:   0,
		CountTLSHandshakeError:     0,
		CountTLSHandshakeInterrupt: 0,
		CountTLSVerificationError:  0,
		CountSuccess:               0,
		HistoTCPConnectError:       map[string]int64{},
		HistoTLSHandshakeError:     map[string]int64{},
		HistoTLSVerificationError:  map[string]int64{},
		LastUpdated:                time.Time{}, // the zero time should exclude this one
		Tactic: &httpsDialerTactic{
			Address:        bridgeAddress,
			InitialDelay:   0,
			Port:           "443",
			SNI:            "ilpost.it",
			VerifyHostname: "api.ooni.io",
		},
	}}

	// createStatsManager creates a stats manager given some baseline stats
	createStatsManager := func(domainEndpoint string, tactics ...*statsTactic) *statsManager {
		container := &statsContainer{
			DomainEndpoints: map[string]*statsDomainEndpoint{
				domainEndpoint: {
					Tactics: map[string]*statsTactic{},
				},
			},
			Version: statsContainerVersion,
		}

		for _, tx := range tactics {
			if tx.Tactic != nil {
				container.DomainEndpoints[domainEndpoint].Tactics[tx.Tactic.tacticSummaryKey()] = tx
			}
		}

		kvStore := &kvstore.Memory{}
		if err := kvStore.Set(statsKey, runtimex.Try1(json.Marshal(container))); err != nil {
			t.Fatal(err)
		}

		const trimInterval = 30 * time.Second
		return newStatsManager(kvStore, log.Log, trimInterval)
	}

	t.Run("when we have relevant stats", func(t *testing.T) {
		// create stats manager
		stats := createStatsManager("api.ooni.io:443", expectTacticsStats...)
		defer stats.Close()

		// create the policy
		policy := &statsPolicyV2{
			Stats: stats,
		}

		// obtain the tactics from the saved stats
		var tactics []*httpsDialerTactic
		for entry := range policy.LookupTactics(context.Background(), "api.ooni.io", "443") {
			tactics = append(tactics, entry)
		}

		// compute the list of results we expect to see from the stats data
		var expect []*httpsDialerTactic
		for _, entry := range expectTacticsStats {
			if entry.CountSuccess <= 0 || entry.Tactic == nil {
				continue // we SHOULD NOT include entries that systematically failed
			}
			t := entry.Tactic.Clone()
			t.InitialDelay = 0
			expect = append(expect, t)
		}

		// perform the actual comparison
		if diff := cmp.Diff(expect, tactics); diff != "" {
			t.Fatal(diff)
		}
	})

	t.Run("when there are no relevant stats", func(t *testing.T) {
		// create stats manager
		stats := createStatsManager("api.ooni.io:443" /*, nothing */)
		defer stats.Close()

		// create the policy
		policy := &statsPolicyV2{
			Stats: stats,
		}

		// obtain the tactics from the saved stats
		var tactics []*httpsDialerTactic
		for entry := range policy.LookupTactics(context.Background(), "api.ooni.io", "443") {
			tactics = append(tactics, entry)
		}

		// compute the list of results we expect to see from the stats data
		var expect []*httpsDialerTactic

		// perform the actual comparison
		if diff := cmp.Diff(expect, tactics); diff != "" {
			t.Fatal(diff)
		}
	})
}

func TestStatsPolicyFilterStatsTactics(t *testing.T) {
	t.Run("we do nothing when good is false", func(t *testing.T) {
		tactics := statsPolicyFilterStatsTactics(nil, false)
		if len(tactics) != 0 {
			t.Fatal("expected zero-lenght return value")
		}
	})

	t.Run("we filter out cases in which t or t.Tactic are nil or entry has no successes", func(t *testing.T) {
		expected := &statsTactic{
			CountStarted:         7,
			CountTCPConnectError: 3,
			CountSuccess:         4,
			HistoTCPConnectError: map[string]int64{
				"generic_timeout_error": 3,
			},
			LastUpdated: time.Now().Add(-11 * time.Second),
			Tactic: &httpsDialerTactic{
				Address:        "130.192.91.211",
				InitialDelay:   0,
				Port:           "443",
				SNI:            "garr.it",
				VerifyHostname: "shelob.polito.it",
			},
		}

		input := []*statsTactic{
			// nil entry
			nil,

			// entry with nil tactic
			{
				CountStarted:               0,
				CountTCPConnectError:       0,
				CountTCPConnectInterrupt:   0,
				CountTLSHandshakeError:     0,
				CountTLSHandshakeInterrupt: 0,
				CountTLSVerificationError:  0,
				CountSuccess:               0,
				HistoTCPConnectError:       map[string]int64{},
				HistoTLSHandshakeError:     map[string]int64{},
				HistoTLSVerificationError:  map[string]int64{},
				LastUpdated:                time.Time{},
				Tactic:                     nil,
			},

			// another nil entry
			nil,

			// an entry that should be OK
			expected,

			// entry that is OK except that it does not contain any
			// success so we don't expect to see it
			{
				CountStarted:           10,
				CountTLSHandshakeError: 10,
				HistoTLSHandshakeError: map[string]int64{
					"generic_timeout_error": 10,
				},
				LastUpdated: time.Now().Add(-4 * time.Second),
				Tactic: &httpsDialerTactic{
					Address:        "130.192.91.211",
					InitialDelay:   0,
					Port:           "443",
					SNI:            "polito.it",
					VerifyHostname: "shelob.polito.it",
				},
			},
		}

		got := statsPolicyFilterStatsTactics(input, true)

		if len(got) != 1 {
			t.Fatal("expected just one element")
		}

		if diff := cmp.Diff(expected.Tactic, got[0]); diff != "" {
			t.Fatal(diff)
		}
	})
}
