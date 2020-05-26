// Package stunreachability contains the STUN reachability experiment.
package stunreachability

import (
	"context"
	"time"

	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/netx/archival"
	"github.com/ooni/probe-engine/netx/dialer"
	"github.com/ooni/probe-engine/netx/httptransport"
	"github.com/ooni/probe-engine/netx/trace"
	"github.com/pion/stun"
)

const (
	testName    = "stun_reachability"
	testVersion = "0.0.1"
)

// Config contains the experiment config.
type Config struct{}

// TestKeys contains the experiment's result.
type TestKeys struct {
	Endpoint      string                   `json:"endpoint"`
	Failure       *string                  `json:"failure"`
	NetworkEvents []archival.NetworkEvent  `json:"network_events"`
	Queries       []archival.DNSQueryEntry `json:"queries"`
}

func registerExtensions(m *model.Measurement) {
	archival.ExtDNS.AddTo(m)
	archival.ExtNetevents.AddTo(m)
}

type measurer struct {
	config Config
}

func (m *measurer) ExperimentName() string {
	return testName
}

func (m *measurer) ExperimentVersion() string {
	return testVersion
}

func (m *measurer) Run(
	ctx context.Context, sess model.ExperimentSession,
	measurement *model.Measurement, callbacks model.ExperimentCallbacks,
) error {
	tk := new(TestKeys)
	measurement.TestKeys = tk
	registerExtensions(measurement)
	if err := tk.run(ctx, sess, measurement, callbacks); err != nil {
		s := err.Error()
		tk.Failure = &s
	}
	return nil
}

func (tk *TestKeys) run(
	ctx context.Context, sess model.ExperimentSession,
	measurement *model.Measurement, callbacks model.ExperimentCallbacks,
) error {
	const defaultAddress = "stun.l.google.com:19302"
	endpoint := string(measurement.Input)
	if endpoint == "" {
		endpoint = defaultAddress
	}
	tk.Endpoint = endpoint
	saver := new(trace.Saver)
	begin := time.Now()
	err := tk.do(httptransport.NewDialer(httptransport.Config{
		ContextByteCounting: true,
		DialSaver:           saver,
		Logger:              sess.Logger(),
		ReadWriteSaver:      saver,
		ResolveSaver:        saver,
	}), endpoint)
	events := saver.Read()
	tk.NetworkEvents = append(
		tk.NetworkEvents, archival.NewNetworkEventsList(begin, events)...,
	)
	tk.Queries = append(
		tk.Queries, archival.NewDNSQueriesList(begin, events, sess.ASNDatabasePath())...,
	)
	return err
}

func (tk *TestKeys) do(dialer dialer.Dialer, endpoint string) error {
	// It's fine to use a background context because we're using UDP
	conn, err := dialer.DialContext(context.Background(), "udp", endpoint)
	if err != nil {
		return err
	}
	defer conn.Close()
	client, err := stun.NewClient(conn, stun.WithNoConnClose)
	if err != nil {
		return err
	}
	message := stun.MustBuild(stun.TransactionID, stun.BindingRequest)
	ch := make(chan error)
	err = client.Start(message, func(ev stun.Event) {
		// As mentioned below this code will run after Start has returned.
		if ev.Error != nil {
			ch <- ev.Error
			return
		}
		var xorAddr stun.XORMappedAddress
		if decodeErr := xorAddr.GetFrom(ev.Message); decodeErr != nil {
			ch <- ev.Error
			return
		}
		ch <- nil
	})
	// Implementation note: if we successfully started, then the callback
	// will be called when we receive a response or fail.
	if err != nil {
		return err
	}
	return <-ch
}

// NewExperimentMeasurer creates a new ExperimentMeasurer.
func NewExperimentMeasurer(config Config) model.ExperimentMeasurer {
	return &measurer{config: config}
}
