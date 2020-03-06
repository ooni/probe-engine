// Package ndt7 contains the ndt7 network experiment.
package ndt7

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/dustin/go-humanize"

	upstream "github.com/m-lab/ndt7-client-go"
	"github.com/m-lab/ndt7-client-go/mlabns"
	"github.com/m-lab/ndt7-client-go/spec"

	"github.com/ooni/probe-engine/model"
)

const (
	testName    = "ndt7"
	testVersion = "0.1.0"
)

// Config contains the experiment settings
type Config struct{}

// TestKeys contains the test keys
type TestKeys struct {
	// Failure is the failure string
	Failure string `json:"failure"`

	// Download contains download results
	Download []spec.Measurement `json:"download"`

	// Upload contains upload results
	Upload []spec.Measurement `json:"upload"`
}

func discover(ctx context.Context, sess model.ExperimentSession) (string, error) {
	client := mlabns.NewClient("ndt7", sess.UserAgent())
	// Basically: (1) make sure we're using our tracing and possibly proxied
	// client rather than default; (2) if we have an explicit proxy make sure
	// we tell mlab-ns to use our IP address rather than the proxy one.
	client.HTTPClient = sess.DefaultHTTPClient()
	if sess.ExplicitProxy() {
		client.RequestMaker = func(
			method, url string, body io.Reader,
		) (*http.Request, error) {
			req, err := http.NewRequest(method, url, body)
			if err != nil {
				return nil, err
			}
			values := req.URL.Query()
			values.Set("ip", sess.ProbeIP())
			req.URL.RawQuery = values.Encode()
			return req, nil
		}
	}
	return client.Query(ctx)
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
	const maxRuntime = 15.0 // second (conservative)
	testkeys := &TestKeys{}
	measurement.TestKeys = testkeys
	client := upstream.NewClient(sess.SoftwareName(), sess.SoftwareVersion())
	FQDN, err := discover(ctx, sess)
	if err != nil {
		testkeys.Failure = err.Error()
		return err
	}
	client.FQDN = FQDN // skip client's own mlabns call
	sess.Logger().Debugf("ndt7: mlabns returned %s to us", FQDN)
	ch, err := client.StartDownload(ctx)
	if err != nil {
		testkeys.Failure = err.Error()
		return err
	}
	callbacks.OnProgress(0, fmt.Sprintf("server: %s", client.FQDN))
	for ev := range ch {
		testkeys.Download = append(testkeys.Download, ev)
		if ev.AppInfo != nil && ev.Origin == "client" {
			elapsed := float64(ev.AppInfo.ElapsedTime) / 1e06 // to seconds
			// The percentage of completion of download goes from 0 to
			// 50% of the whole experiment, hence the `/2.0`.
			percentage := elapsed / maxRuntime / 2.0
			speed := float64(ev.AppInfo.NumBytes) * 8.0 / elapsed
			message := fmt.Sprintf(
				"download-speed %s", humanize.SI(float64(speed), "bit/s"),
			)
			callbacks.OnProgress(percentage, message)
		}
		data, err := json.Marshal(ev)
		if err != nil {
			testkeys.Failure = err.Error()
			return err
		}
		sess.Logger().Debugf("%s", string(data))
	}
	ch, err = client.StartUpload(ctx)
	if err != nil {
		testkeys.Failure = err.Error()
		return err
	}
	for ev := range ch {
		testkeys.Upload = append(testkeys.Upload, ev)
		if ev.AppInfo != nil && ev.Origin == "client" {
			elapsed := float64(ev.AppInfo.ElapsedTime) / 1e06 // to seconds
			// The percentage of completion of upload goes from 50% to 100% of
			// the whole experiment, hence `0.5 +` and `/2.0`.
			percentage := 0.5 + elapsed/maxRuntime/2.0
			speed := float64(ev.AppInfo.NumBytes) * 8.0 / elapsed
			message := fmt.Sprintf(
				"upload-speed %s", humanize.SI(float64(speed), "bit/s"),
			)
			callbacks.OnProgress(percentage, message)
		}
		data, err := json.Marshal(ev)
		if err != nil {
			testkeys.Failure = err.Error()
			return err
		}
		sess.Logger().Debugf("%s", string(data))
	}
	callbacks.OnProgress(1, "done")
	return nil
}

// NewExperimentMeasurer creates a new ExperimentMeasurer.
func NewExperimentMeasurer(config Config) model.ExperimentMeasurer {
	return &measurer{config: config}
}
