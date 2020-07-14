package webconnectivity

import (
	"context"
	"fmt"
	"net/url"

	"github.com/ooni/probe-engine/experiment/urlgetter"
	"github.com/ooni/probe-engine/model"
)

// DNSLookupConfig contains settings for the DNS lookup.
type DNSLookupConfig struct {
	Session model.ExperimentSession
	URL     *url.URL
}

// DNSLookupResult contains the result of the DNS lookup.
type DNSLookupResult struct {
	Addrs    map[string]int64
	Failure  *string
	TestKeys urlgetter.TestKeys
}

// DNSLookup performs the DNS lookup part of Web Connectivity.
func DNSLookup(ctx context.Context, config DNSLookupConfig) (out DNSLookupResult) {
	target := fmt.Sprintf("dnslookup://%s", config.URL.Hostname())
	config.Session.Logger().Infof("dnslookup %s...", target)
	result, err := urlgetter.Getter{Session: config.Session, Target: target}.Get(ctx)
	out.Addrs = make(map[string]int64)
	for _, query := range result.Queries {
		for _, answer := range query.Answers {
			if answer.IPv4 != "" {
				out.Addrs[answer.IPv4] = answer.ASN
				continue
			}
			if answer.IPv6 != "" {
				out.Addrs[answer.IPv6] = answer.ASN
			}
		}
	}
	config.Session.Logger().Infof("dnslookup %s... %+v %+v", target, err, out.Addrs)
	out.Failure = result.FailedOperation
	out.TestKeys = result
	return
}
