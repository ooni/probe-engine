package webconnectivityqa

import (
	"github.com/ooni/probe-engine/pkg/netemx"
	"github.com/ooni/probe-engine/pkg/runtimex"
)

// dnsBlockingAndroidDNSCacheNoData is the case where we're on Android and the getaddrinfo
// resolver returns the android_dns_cache_no_data error.
func dnsBlockingAndroidDNSCacheNoData() *TestCase {
	return &TestCase{
		Name:  "dnsBlockingAndroidDNSCacheNoData",
		Flags: TestCaseFlagNoV04,
		Input: "https://www.example.com/",
		Configure: func(env *netemx.QAEnv) {
			// make sure the env knows we want to emulate our getaddrinfo wrapper behavior
			env.EmulateAndroidGetaddrinfo(true)

			// remove the record so that the DNS query returns NXDOMAIN, which is then
			// converted into android_dns_cache_no_data by the emulation layer
			env.ISPResolverConfig().RemoveRecord("www.example.com")
		},
		ExpectErr: false,
		ExpectTestKeys: &TestKeys{
			DNSExperimentFailure:  "android_dns_cache_no_data",
			HTTPExperimentFailure: "android_dns_cache_no_data",
			DNSConsistency:        "inconsistent",
			XStatus:               2080, // StatusExperimentDNS | StatusAnomalyDNS
			XDNSFlags:             2,    // AnalysisDNSFlagUnexpectedFailure
			XBlockingFlags:        33,   // AnalysisBlockingFlagDNSBlocking | AnalysisBlockingFlagSuccess
			Accessible:            false,
			Blocking:              "dns",
		},
	}
}

// dnsBlockingNXDOMAIN is the case where there's DNS blocking using NXDOMAIN.
func dnsBlockingNXDOMAIN() *TestCase {
	/*
		Historical note:

		With this test case there was an MK bug where we didn't properly record the
		actual error that occurred when performing the DNS experiment.

		See <https://github.com/measurement-kit/measurement-kit/issues/1931>.
	*/
	return &TestCase{
		Name:  "dnsBlockingNXDOMAIN",
		Flags: TestCaseFlagNoV04,
		Input: "https://www.example.com/",
		Configure: func(env *netemx.QAEnv) {
			// remove the record so that the DNS query returns NXDOMAIN
			env.ISPResolverConfig().RemoveRecord("www.example.com")
		},
		ExpectErr: false,
		ExpectTestKeys: &TestKeys{
			DNSExperimentFailure:  "dns_nxdomain_error",
			HTTPExperimentFailure: "dns_nxdomain_error",
			DNSConsistency:        "inconsistent",
			XStatus:               2080, // StatusExperimentDNS | StatusAnomalyDNS
			XDNSFlags:             2,    // AnalysisDNSFlagUnexpectedFailure
			XBlockingFlags:        33,   // AnalysisBlockingFlagDNSBlocking | AnalysisBlockingFlagSuccess
			Accessible:            false,
			Blocking:              "dns",
		},
	}
}

// dnsBlockingBOGON is the case where there's DNS blocking by returning a bogon.
func dnsBlockingBOGON() *TestCase {
	return &TestCase{
		Name:  "dnsBlockingBOGON",
		Flags: 0,
		Input: "https://www.example.com/",
		Configure: func(env *netemx.QAEnv) {
			env.ISPResolverConfig().RemoveRecord("www.example.com")
			runtimex.Try0(env.ISPResolverConfig().AddRecord("www.example.com", "", "10.10.34.35"))
		},
		ExpectErr: false,
		ExpectTestKeys: &TestKeys{
			HTTPExperimentFailure: "generic_timeout_error",
			DNSExperimentFailure:  nil,
			DNSConsistency:        "inconsistent",
			XStatus:               4256, // StatusExperimentConnect | StatusAnomalyConnect | StatusAnomalyDNS
			XDNSFlags:             5,    // AnalysisFlagDNSBogon | AnalysisDNSFlagUnexpectedAddrs
			XBlockingFlags:        33,   // AnalysisBlockingFlagDNSBlocking | AnalysisBlockingFlagSuccess
			Accessible:            false,
			Blocking:              "dns",
		},
	}
}
