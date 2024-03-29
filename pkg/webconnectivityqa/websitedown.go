package webconnectivityqa

import "github.com/ooni/probe-engine/pkg/netemx"

// websiteDownNXDOMAIN describes the test case where the website domain
// is NXDOMAIN according to the TH and the probe.
func websiteDownNXDOMAIN() *TestCase {
	/*
	   TODO(bassosimone): Debateable result for v0.4, while v0.5 behaves in the
	   correct way. See <https://github.com/ooni/probe-engine/issues/579>.

	   Some historical context follows:

	   Note that MK is not doing it right here because it's suppressing the
	   dns_nxdomain_error that instead is very informative. Yet, it is reporting
	   a failure in HTTP, which miniooni does not because it does not make
	   sense to perform HTTP when there are no IP addresses.

	   The following seems indeed a bug in MK where we don't properly record the
	   actual error that occurred when performing the DNS experiment.

	   See <https://github.com/measurement-kit/measurement-kit/issues/1931>.
	*/
	return &TestCase{
		Name:      "websiteDownNXDOMAIN",
		Flags:     TestCaseFlagNoV04,
		Input:     "http://www.example.xyz/", // domain not defined in the simulation
		Configure: nil,
		ExpectErr: false,
		ExpectTestKeys: &TestKeys{
			DNSExperimentFailure:  "dns_nxdomain_error",
			HTTPExperimentFailure: "dns_nxdomain_error",
			DNSConsistency:        "consistent",
			XStatus:               2052, // StatusExperimentDNS | StatusSuccessNXDOMAIN
			XBlockingFlags:        0,
			XNullNullFlags:        1, // AnalysisFlagNullNullExpectedDNSLookupFailure
			Accessible:            false,
			Blocking:              false,
		},
	}
}

// websiteDownTCPConnect describes the test case where attempting to
// connect to the website doesn't work for both probe and TH.
//
// See https://github.com/ooni/probe/issues/2299.
func websiteDownTCPConnect() *TestCase {
	return &TestCase{
		Name:      "websiteDownTCPConnect",
		Flags:     TestCaseFlagNoV04,
		Input:     "http://www.example.com:444/", // port where we're not listening.
		Configure: nil,
		ExpectErr: false,
		ExpectTestKeys: &TestKeys{
			HTTPExperimentFailure: "connection_refused",
			DNSConsistency:        "consistent",
			XStatus:               2052, // StatusExperimentDNS | StatusSuccessNXDOMAIN
			XBlockingFlags:        0,
			XNullNullFlags:        2, // AnalysisFlagNullNullExpectedTCPConnectFailure
			Accessible:            false,
			Blocking:              false,
		},
	}
}

// websiteDownNoAddrs describes the test case where the website domain
// does not return any address according to the TH and the probe.
func websiteDownNoAddrs() *TestCase {
	return &TestCase{
		Name:  "websiteDownNoAddrs",
		Flags: TestCaseFlagNoV04,
		Input: "http://www.example.com/",
		Configure: func(env *netemx.QAEnv) {

			// reconfigure with only CNAME but no addresses and do this
			// for all the resolvers of the kingdom
			env.AddRecordToAllResolvers("www.example.com", "web01.example.com" /* No addrs */)

		},
		ExpectErr: false,
		ExpectTestKeys: &TestKeys{
			DNSExperimentFailure: "dns_no_answer",
			DNSConsistency:       "consistent",
			XBlockingFlags:       0,
			XNullNullFlags:       1, // AnalysisFlagNullNullExpectedDNSLookupFailure
			Accessible:           false,
			Blocking:             false,
		},
	}
}
