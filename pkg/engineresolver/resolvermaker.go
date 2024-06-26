package engineresolver

//
// High-level code for creating a new child resolver
//

import (
	"math/rand"
	"strings"
	"time"

	"github.com/ooni/probe-engine/pkg/model"
)

// resolvemaker contains rules for making a resolver.
type resolvermaker struct {
	url   string
	score float64
}

// systemResolverURL is the URL of the system resolver.
const systemResolverURL = "system:///"

// allmakers contains all the makers in a list. We use the http3
// prefix to indicate we wanna use http3. The code will translate
// this to https and set the proper netx options.
var allmakers = []*resolvermaker{{
	url: "https://cloudflare-dns.com/dns-query",
}, {
	url: "http3://cloudflare-dns.com/dns-query",
}, {
	url: "https://dns.google/dns-query",
}, {
	url: "http3://dns.google/dns-query",
}, {
	url: "https://dns.quad9.net/dns-query",
}, {
	url: systemResolverURL,
}, {
	url: "https://mozilla.cloudflare-dns.com/dns-query",
}, {
	url: "http3://mozilla.cloudflare-dns.com/dns-query",
}, {
	url: "https://wikimedia-dns.org/dns-query",
}}

// allbyurl contains all the resolvermakers by URL
var allbyurl = resolverMakeInitialState()

// resolverMakeInitialState initializes the initial
// state by giving a nonzero initial score
// to all resolvers except for the system resolver. We set
// the system resolver score to be 0.5, so that it's less
// likely than other resolvers in this list.
//
// We used to set this value to 0, but this proved to
// create issues when it was the only available resolver,
// see https://github.com/ooni/probe/issues/2544.
func resolverMakeInitialState() map[string]*resolvermaker {
	output := make(map[string]*resolvermaker)
	rng := rand.New(rand.NewSource(time.Now().UnixNano())) // #nosec G404 -- not really important
	for _, e := range allmakers {
		output[e.url] = e
		if e.url != systemResolverURL {
			e.score = rng.Float64()
		} else {
			e.score = 0.5
		}
	}
	return output
}

// logger returns the configured logger or a default
func (r *Resolver) logger() model.Logger {
	return model.ValidLoggerOrDefault(r.Logger)
}

// newChildResolver creates a new child model.Resolver.
func (r *Resolver) newChildResolver(h3 bool, URL string) (model.Resolver, error) {
	if r.newChildResolverFn != nil {
		return r.newChildResolverFn(h3, URL)
	}
	return newChildResolver(
		r.logger(),
		URL,
		h3,
		r.ByteCounter, // newChildResolver handles the nil case
		r.ProxyURL,    // ditto
	)
}

// newresolver creates a new resolver with the given config and URL. This is
// where we expand http3 to https and set the h3 options.
func (r *Resolver) newresolver(URL string) (model.Resolver, error) {
	h3 := strings.HasPrefix(URL, "http3://")
	if h3 {
		URL = strings.Replace(URL, "http3://", "https://", 1)
	}
	return r.newChildResolver(h3, URL)
}

// getresolver returns a resolver with the given URL. This function caches
// already allocated resolvers so we only allocate them once.
func (r *Resolver) getresolver(URL string) (model.Resolver, error) {
	defer r.mu.Unlock()
	r.mu.Lock()
	if re, found := r.res[URL]; found {
		return re, nil // already created
	}
	re, err := r.newresolver(URL)
	if err != nil {
		return nil, err // config err?
	}
	if r.res == nil {
		r.res = make(map[string]model.Resolver)
	}
	r.res[URL] = re
	return re, nil
}

// closeall closes the cached resolvers.
func (r *Resolver) closeall() {
	defer r.mu.Unlock()
	r.mu.Lock()
	for _, re := range r.res {
		re.CloseIdleConnections()
	}
	r.res = nil
}
