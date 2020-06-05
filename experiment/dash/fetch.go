package dash

import (
	"context"
	"runtime"
	"time"

	"github.com/ooni/probe-engine/netx/trace"
)

type fetchConfig struct {
	authorization string
	begin         time.Time
	deps          downloadDeps
	errch         chan error
	fqdn          string
	numIterations int64
	outch         chan clientResults
	realAddress   string
	saver         *trace.Saver
}

func fetch(ctx context.Context, config fetchConfig) {
	// Note: according to a comment in MK sources 3000 kbit/s was the
	// minimum speed recommended by Netflix for SD quality in 2017.
	//
	// See: <https://help.netflix.com/en/node/306>.
	const initialBitrate = 3000
	current := clientResults{
		ElapsedTarget: 2,
		Platform:      runtime.GOOS,
		Rate:          initialBitrate,
		RealAddress:   config.realAddress,
		Version:       magicVersion,
	}
	var connectTime float64
	// Because the new implementation emulates a player and we always
	// want to have a single frame in the playout buffer, we need to
	// download `numIterations + 1` frames to play numIterations frames.
	for current.Iteration <= config.numIterations {
		result, err := download(ctx, downloadConfig{
			authorization: config.authorization,
			begin:         config.begin,
			currentRate:   current.Rate,
			deps:          config.deps,
			elapsedTarget: current.ElapsedTarget,
			fqdn:          config.fqdn,
		})
		if err != nil {
			// Implementation note: ndt7 controls the connection much
			// more than us and it can tell whether an error occurs when
			// connecting or later. We cannot say that very precisely
			// because, in principle, we may reconnect. So we always
			// return error here. This comment is being introduced so
			// that we don't do https://github.com/ooni/probe-engine/pull/526
			// again, because that isn't accurate.
			config.errch <- err
			return
		}
		current.Elapsed = result.elapsed
		current.Received = result.received
		current.RequestTicks = result.requestTicks
		current.Timestamp = result.timestamp
		current.ServerURL = result.serverURL
		// Read the events so far and possibly update our measurement
		// of the latest connect time. We should have one sample in most
		// cases, because the connection should be persistent.
		for _, ev := range config.saver.Read() {
			if ev.Name == "connect" {
				connectTime = ev.Duration.Seconds()
			}
		}
		current.ConnectTime = connectTime
		config.outch <- current
		current.Iteration++
		speed := float64(current.Received) / float64(current.Elapsed)
		speed *= 8.0    // to bits per second
		speed /= 1000.0 // to kbit/s
		current.Rate = int64(speed)
	}
	config.errch <- nil
}
