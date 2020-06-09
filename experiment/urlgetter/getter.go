package urlgetter

import (
	"context"
	"time"

	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/netx/archival"
	"github.com/ooni/probe-engine/netx/errorx"
	"github.com/ooni/probe-engine/netx/modelx"
	"github.com/ooni/probe-engine/netx/trace"
)

// The Getter gets the specified target in the context of the
// given session and with the specified config.
//
// Other OONI experiment should use the Getter to factor code when
// the Getter implements the operations they wanna perform.
type Getter struct {
	Config  Config
	Session model.ExperimentSession
	Target  string
}

// Get performs the action described by g using the given context
// and returning the test keys and eventually an error
func (g Getter) Get(ctx context.Context) (TestKeys, error) {
	begin := time.Now()
	saver := new(trace.Saver)
	tk, err := g.get(ctx, saver)
	// Make sure we have an operation in cases where we fail before
	// hitting our httptransport that does error wrapping.
	err = errorx.SafeErrWrapperBuilder{
		Error:     err,
		Operation: modelx.TopLevelOperation,
	}.MaybeBuild()
	tk.FailedOperation = archival.NewFailedOperation(err)
	tk.Failure = archival.NewFailure(err)
	events := saver.Read()
	tk.Queries = append(
		tk.Queries, archival.NewDNSQueriesList(
			begin, events, g.Session.ASNDatabasePath())...,
	)
	tk.NetworkEvents = append(
		tk.NetworkEvents, archival.NewNetworkEventsList(begin, events)...,
	)
	tk.Requests = append(
		tk.Requests, archival.NewRequestList(begin, events)...,
	)
	tk.TCPConnect = append(
		tk.TCPConnect, archival.NewTCPConnectList(begin, events)...,
	)
	tk.TLSHandshakes = append(
		tk.TLSHandshakes, archival.NewTLSHandshakesList(begin, events)...,
	)
	return tk, err
}

func (g Getter) get(ctx context.Context, saver *trace.Saver) (TestKeys, error) {
	tk := TestKeys{
		Agent:    "redirect",
		DNSCache: []string{g.Config.DNSCache},
		Tunnel:   g.Config.Tunnel,
	}
	if g.Config.NoFollowRedirects {
		tk.Agent = "agent"
	}
	// start tunnel
	if err := g.Session.MaybeStartTunnel(ctx, g.Config.Tunnel); err != nil {
		return tk, err
	}
	tk.BootstrapTime = g.Session.TunnelBootstrapTime().Seconds()
	if url := g.Session.ProxyURL(); url != nil {
		tk.SOCKSProxy = url.Host
	}
	// create configuration
	configurer := Configurer{
		Config:   g.Config,
		Logger:   g.Session.Logger(),
		ProxyURL: g.Session.ProxyURL(),
		Saver:    saver,
	}
	configuration, err := configurer.NewConfiguration()
	if err != nil {
		return tk, err
	}
	defer configuration.CloseIdleConnections()
	// run the measurement
	runner := Runner{
		Config:     g.Config,
		HTTPConfig: configuration.HTTPConfig,
		Target:     g.Target,
	}
	return tk, runner.Run(ctx)
}
