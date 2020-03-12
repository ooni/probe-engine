// Package ndt7 contains the ndt7 network experiment.
package ndt7

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/m-lab/ndt7-client-go/spec"
	"github.com/ooni/probe-engine/internal/mlablocate"
	"github.com/ooni/probe-engine/model"
)

const (
	testName    = "ndt7"
	testVersion = "0.2.0"
)

// Config contains the experiment settings
type Config struct{}

// Summary is the measurement summary
type Summary struct {
	AvgRTT         float64 `json:"avg_rtt"`         // Average RTT [ms]
	Download       float64 `json:"download"`        // download speed [kbit/s]
	MSS            int64   `json:"mss"`             // MSS
	MaxRTT         float64 `json:"max_rtt"`         // Max AvgRTT sample seen [ms]
	MinRTT         float64 `json:"min_rtt"`         // Min RTT according to kernel [ms]
	Ping           float64 `json:"ping"`            // Equivalent to MinRTT [ms]
	RetransmitRate float64 `json:"retransmit_rate"` // bytes_retrans/bytes_sent [0..1]
	Upload         float64 `json:"upload"`          // upload speed [kbit/s]
}

// TestKeys contains the test keys
type TestKeys struct {
	// Download contains download results
	Download []spec.Measurement `json:"download"`

	// Failure is the failure string
	Failure *string `json:"failure"`

	// Summary contains the measurement summary
	Summary Summary `json:"summary"`

	// Upload contains upload results
	Upload []spec.Measurement `json:"upload"`
}

type measurer struct {
	config          Config
	jsonUnmarshal   func(data []byte, v interface{}) error
	preDownloadHook func()
	preUploadHook   func()
}

func (m *measurer) discover(ctx context.Context, sess model.ExperimentSession) (string, error) {
	client := mlablocate.NewClient(sess.DefaultHTTPClient(), sess.Logger(), sess.UserAgent())
	if sess.ExplicitProxy() {
		client.NewRequest = mlablocate.NewRequestWithProxy(sess.ProbeIP())
	}
	return client.Query(ctx, "ndt7")
}

func (m *measurer) ExperimentName() string {
	return testName
}

func (m *measurer) ExperimentVersion() string {
	return testVersion
}

func (m *measurer) doDownload(
	ctx context.Context, sess model.ExperimentSession,
	callbacks model.ExperimentCallbacks, tk *TestKeys,
	hostname string,
) error {
	conn, err := newDialManager(hostname).dialDownload(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	mgr := newDownloadManager(
		conn,
		func(timediff time.Duration, count int64) {
			elapsed := timediff.Seconds()
			// The percentage of completion of download goes from 0 to
			// 50% of the whole experiment, hence the `/2.0`.
			percentage := elapsed / paramMaxRuntimeUpperBound / 2.0
			speed := float64(count) * 8.0 / elapsed
			message := fmt.Sprintf("download-speed %s", humanize.SI(float64(speed), "bit/s"))
			tk.Summary.Download = speed / 1e03 /* bit/s => kbit/s */
			callbacks.OnProgress(percentage, message)
			tk.Download = append(tk.Download, spec.Measurement{
				AppInfo: &spec.AppInfo{
					ElapsedTime: int64(timediff / time.Microsecond),
					NumBytes:    count,
				},
				Origin: "client",
				Test:   "download",
			})
		},
		func(data []byte) error {
			sess.Logger().Debugf("%s", string(data))
			var measurement spec.Measurement
			if err := m.jsonUnmarshal(data, &measurement); err != nil {
				return err
			}
			if measurement.TCPInfo != nil {
				rtt := float64(measurement.TCPInfo.RTT) / 1e03 /* us => ms */
				tk.Summary.AvgRTT = rtt
				tk.Summary.MSS = int64(measurement.TCPInfo.AdvMSS)
				if tk.Summary.MaxRTT < rtt {
					tk.Summary.MaxRTT = rtt
				}
				tk.Summary.MinRTT = float64(measurement.TCPInfo.MinRTT) / 1e03 /* us => ms */
				tk.Summary.Ping = tk.Summary.MinRTT
				if measurement.TCPInfo.BytesSent > 0 {
					tk.Summary.RetransmitRate = (float64(measurement.TCPInfo.BytesRetrans) /
						float64(measurement.TCPInfo.BytesSent))
				}
				measurement.BBRInfo = nil        // don't encourage people to use it
				measurement.ConnectionInfo = nil // do we need to save it?
				measurement.Origin = "server"
				measurement.Test = "download"
				tk.Download = append(tk.Download, measurement)
			}
			return nil
		},
	)
	if err := mgr.run(ctx); err != nil {
		sess.Logger().Warnf("download: %s", err)
	}
	return nil // failure is only when we cannot connect
}

func (m *measurer) doUpload(
	ctx context.Context, sess model.ExperimentSession,
	callbacks model.ExperimentCallbacks, tk *TestKeys,
	hostname string,
) error {
	conn, err := newDialManager(hostname).dialUpload(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	mgr := newUploadManager(
		conn,
		func(timediff time.Duration, count int64) {
			elapsed := timediff.Seconds()
			// The percentage of completion of upload goes from 50% to 100% of
			// the whole experiment, hence `0.5 +` and `/2.0`.
			percentage := 0.5 + elapsed/paramMaxRuntimeUpperBound/2.0
			speed := float64(count) * 8.0 / elapsed
			message := fmt.Sprintf("upload-speed %s", humanize.SI(float64(speed), "bit/s"))
			tk.Summary.Upload = speed / 1e03 /* bit/s => kbit/s */
			callbacks.OnProgress(percentage, message)
			tk.Upload = append(tk.Upload, spec.Measurement{
				AppInfo: &spec.AppInfo{
					ElapsedTime: int64(timediff / time.Microsecond),
					NumBytes:    count,
				},
				Origin: "client",
				Test:   "upload",
			})
		},
	)
	if err := mgr.run(ctx); err != nil {
		sess.Logger().Warnf("upload: %s", err)
	}
	return nil // failure is only when we cannot connect
}

func (m *measurer) Run(
	ctx context.Context, sess model.ExperimentSession,
	measurement *model.Measurement, callbacks model.ExperimentCallbacks,
) error {
	tk := new(TestKeys)
	measurement.TestKeys = tk
	hostname, err := m.discover(ctx, sess)
	if err != nil {
		tk.Failure = failureFromError(err)
		return err
	}
	callbacks.OnProgress(0, fmt.Sprintf("downloading: %s", hostname))
	if m.preDownloadHook != nil {
		m.preDownloadHook()
	}
	if err := m.doDownload(ctx, sess, callbacks, tk, hostname); err != nil {
		tk.Failure = failureFromError(err)
		return err
	}
	callbacks.OnProgress(0.5, fmt.Sprintf("uploading: %s", hostname))
	if m.preUploadHook != nil {
		m.preUploadHook()
	}
	if err := m.doUpload(ctx, sess, callbacks, tk, hostname); err != nil {
		tk.Failure = failureFromError(err)
		return err
	}
	callbacks.OnProgress(1, "done")
	return nil
}

// NewExperimentMeasurer creates a new ExperimentMeasurer.
func NewExperimentMeasurer(config Config) model.ExperimentMeasurer {
	return &measurer{config: config, jsonUnmarshal: json.Unmarshal}
}

func failureFromError(err error) (failure *string) {
	if err != nil {
		s := err.Error()
		failure = &s
	}
	return
}
