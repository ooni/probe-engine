// Package whatsapp contains the WhatsApp network experiment.
//
// See https://github.com/ooni/spec/blob/master/nettests/ts-018-whatsapp.md.
package whatsapp

import (
	"context"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/ooni/probe-engine/experiment/urlgetter"
	"github.com/ooni/probe-engine/model"
)

const (
	registrationServiceURL = "https://v.whatsapp.net/v2/register"
	testName               = "whatsapp"
	testVersion            = "0.7.0"
	webHTTPURL             = "http://web.whatsapp.com"
	webHTTPSURL            = "https://web.whatsapp.com"
)

var endpointPattern = regexp.MustCompile("^tcpconnect://e[0-9]{1,2}.whatsapp.net:[0-9]{3,5}$")

// Config contains the experiment config.
type Config struct {
	AllEndpoints bool `ooni:"Whether to test all WhatsApp endpoints"`
}

// TestKeys contains the experiment results
type TestKeys struct {
	urlgetter.TestKeys
	RegistrationServerFailure        *string  `json:"registration_server_failure"`
	RegistrationServerStatus         string   `json:"registration_server_status"`
	WhatsappEndpointsBlocked         []string `json:"whatsapp_endpoints_blocked"`
	WhatsappEndpointsDNSInconsistent []string `json:"whatsapp_endpoints_dns_inconsistent"`
	WhatsappEndpointsStatus          string   `json:"whatsapp_endpoints_status"`
	WhatsappWebStatus                string   `json:"whatsapp_web_status"`
	WhatsappWebFailure               *string  `json:"whatsapp_web_failure"`
	WhatsappHTTPFailure              *string  `json:"-"`
	WhatsappHTTPSFailure             *string  `json:"-"`
}

// NewTestKeys returns a new instance of the test keys.
func NewTestKeys() *TestKeys {
	failure := "unknown_failure"
	return &TestKeys{
		RegistrationServerFailure:        &failure,
		RegistrationServerStatus:         "blocked",
		WhatsappEndpointsBlocked:         []string{},
		WhatsappEndpointsDNSInconsistent: []string{},
		WhatsappEndpointsStatus:          "blocked",
		WhatsappWebFailure:               &failure,
		WhatsappWebStatus:                "blocked",
		WhatsappHTTPFailure:              &failure,
		WhatsappHTTPSFailure:             &failure,
	}
}

// Update updates the TestKeys using the given MultiOutput result.
func (tk *TestKeys) Update(v urlgetter.MultiOutput) {
	// update the easy to update entries first
	tk.NetworkEvents = append(tk.NetworkEvents, v.TestKeys.NetworkEvents...)
	tk.Queries = append(tk.Queries, v.TestKeys.Queries...)
	tk.Requests = append(tk.Requests, v.TestKeys.Requests...)
	tk.TCPConnect = append(tk.TCPConnect, v.TestKeys.TCPConnect...)
	tk.TLSHandshakes = append(tk.TLSHandshakes, v.TestKeys.TLSHandshakes...)
	// set the status of WhatsApp endpoints
	if endpointPattern.MatchString(v.Input.Target) {
		if v.TestKeys.Failure != nil {
			endpoint := strings.ReplaceAll(v.Input.Target, "tcpconnect://", "")
			tk.WhatsappEndpointsBlocked = append(tk.WhatsappEndpointsBlocked, endpoint)
			return
		}
		tk.WhatsappEndpointsStatus = "ok"
		return
	}
	// set the status of the registration service
	if v.Input.Target == registrationServiceURL {
		tk.RegistrationServerFailure = v.TestKeys.Failure
		if v.TestKeys.Failure == nil {
			tk.RegistrationServerStatus = "ok"
		}
		return
	}
	switch v.Input.Target {
	case webHTTPSURL:
		tk.WhatsappHTTPSFailure = v.TestKeys.Failure
	case webHTTPURL:
		tk.WhatsappHTTPFailure = v.TestKeys.Failure
	}
}

// ComputeWebStatus sets the web status fields.
func (tk *TestKeys) ComputeWebStatus() {
	if tk.WhatsappHTTPFailure == nil && tk.WhatsappHTTPSFailure == nil {
		tk.WhatsappWebFailure = nil
		tk.WhatsappWebStatus = "ok"
		return
	}
	tk.WhatsappWebStatus = "blocked" // must be here because of unit tests
	if tk.WhatsappHTTPSFailure != nil {
		tk.WhatsappWebFailure = tk.WhatsappHTTPSFailure
		return
	}
	tk.WhatsappWebFailure = tk.WhatsappHTTPFailure
}

type measurer struct {
	config Config
}

func (m measurer) ExperimentName() string {
	return testName
}

func (m measurer) ExperimentVersion() string {
	return testVersion
}

func (m measurer) Run(
	ctx context.Context, sess model.ExperimentSession,
	measurement *model.Measurement, callbacks model.ExperimentCallbacks,
) error {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	urlgetter.RegisterExtensions(measurement)
	// generate all the inputs
	var inputs []urlgetter.MultiInput
	for idx := 1; idx <= 16; idx++ {
		for _, port := range []string{"443", "5222"} {
			inputs = append(inputs, urlgetter.MultiInput{
				Target: fmt.Sprintf("tcpconnect://e%d.whatsapp.net:%s", idx, port),
			})
		}
	}
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	rnd.Shuffle(len(inputs), func(i, j int) {
		inputs[i], inputs[j] = inputs[j], inputs[i]
	})
	if m.config.AllEndpoints == false {
		inputs = inputs[0:1]
	}
	inputs = append(inputs, urlgetter.MultiInput{Target: registrationServiceURL})
	inputs = append(inputs, urlgetter.MultiInput{Target: webHTTPSURL})
	inputs = append(inputs, urlgetter.MultiInput{Target: webHTTPURL})
	// measure in parallel
	multi := urlgetter.Multi{Begin: time.Now(), Session: sess}
	testkeys := NewTestKeys()
	testkeys.Agent = "redirect"
	measurement.TestKeys = testkeys
	for entry := range multi.Collect(ctx, inputs, "whatsapp", callbacks) {
		testkeys.Update(entry)
	}
	testkeys.ComputeWebStatus()
	return nil
}

// NewExperimentMeasurer creates a new ExperimentMeasurer.
func NewExperimentMeasurer(config Config) model.ExperimentMeasurer {
	return measurer{config: config}
}
