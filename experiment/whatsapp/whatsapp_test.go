package whatsapp_test

import (
	"context"
	"io"
	"regexp"
	"testing"

	"github.com/apex/log"
	"github.com/google/go-cmp/cmp"
	"github.com/ooni/probe-engine/atomicx"
	"github.com/ooni/probe-engine/experiment/handler"
	"github.com/ooni/probe-engine/experiment/internal/httpfailure"
	"github.com/ooni/probe-engine/experiment/urlgetter"
	"github.com/ooni/probe-engine/experiment/whatsapp"
	"github.com/ooni/probe-engine/internal/mockable"
	"github.com/ooni/probe-engine/model"
)

func TestNewExperimentMeasurer(t *testing.T) {
	measurer := whatsapp.NewExperimentMeasurer(whatsapp.Config{})
	if measurer.ExperimentName() != "whatsapp" {
		t.Fatal("unexpected name")
	}
	if measurer.ExperimentVersion() != "0.8.0" {
		t.Fatal("unexpected version")
	}
}

func TestIntegrationSuccess(t *testing.T) {
	measurer := whatsapp.NewExperimentMeasurer(whatsapp.Config{})
	ctx := context.Background()
	sess := &mockable.ExperimentSession{MockableLogger: log.Log}
	measurement := new(model.Measurement)
	callbacks := handler.NewPrinterCallbacks(log.Log)
	err := measurer.Run(ctx, sess, measurement, callbacks)
	if err != nil {
		t.Fatal(err)
	}
	tk := measurement.TestKeys.(*whatsapp.TestKeys)
	if tk.RegistrationServerFailure != nil {
		t.Fatal("invalid RegistrationServerFailure")
	}
	if tk.RegistrationServerStatus != "ok" {
		t.Fatal("invalid RegistrationServerStatus")
	}
	if len(tk.WhatsappEndpointsBlocked) != 0 {
		t.Fatal("invalid WhatsappEndpointsBlocked")
	}
	if len(tk.WhatsappEndpointsDNSInconsistent) != 0 {
		t.Fatal("invalid WhatsappEndpointsDNSInconsistent")
	}
	if tk.WhatsappEndpointsStatus != "ok" {
		t.Fatal("invalid WhatsappEndpointsStatus")
	}
	if tk.WhatsappWebFailure != nil {
		t.Fatal("invalid WhatsappWebFailure")
	}
	if tk.WhatsappWebStatus != "ok" {
		t.Fatal("invalid WhatsappWebStatus")
	}
}

func TestIntegrationFailureAllEndpoints(t *testing.T) {
	measurer := whatsapp.NewExperimentMeasurer(whatsapp.Config{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	sess := &mockable.ExperimentSession{MockableLogger: log.Log}
	measurement := new(model.Measurement)
	callbacks := handler.NewPrinterCallbacks(log.Log)
	err := measurer.Run(ctx, sess, measurement, callbacks)
	if err != nil {
		t.Fatal(err)
	}
	tk := measurement.TestKeys.(*whatsapp.TestKeys)
	if *tk.RegistrationServerFailure != "interrupted" {
		t.Fatal("invalid RegistrationServerFailure")
	}
	if tk.RegistrationServerStatus != "blocked" {
		t.Fatal("invalid RegistrationServerStatus")
	}
	if len(tk.WhatsappEndpointsBlocked) != 16 {
		t.Fatal("invalid WhatsappEndpointsBlocked")
	}
	pattern := regexp.MustCompile("^e[0-9]{1,2}.whatsapp.net$")
	for i := 0; i < len(tk.WhatsappEndpointsBlocked); i++ {
		if !pattern.MatchString(tk.WhatsappEndpointsBlocked[i]) {
			t.Fatalf("invalid WhatsappEndpointsBlocked[%d]", i)
		}
	}
	if len(tk.WhatsappEndpointsDNSInconsistent) != 0 {
		t.Fatal("invalid WhatsappEndpointsDNSInconsistent")
	}
	if tk.WhatsappEndpointsStatus != "blocked" {
		t.Fatal("invalid WhatsappEndpointsStatus")
	}
	if *tk.WhatsappWebFailure != "interrupted" {
		t.Fatal("invalid WhatsappWebFailure")
	}
	if tk.WhatsappWebStatus != "blocked" {
		t.Fatal("invalid WhatsappWebStatus")
	}
}

func TestTestKeysComputeWebStatus(t *testing.T) {
	errorString := io.EOF.Error()
	secondErrorString := context.Canceled.Error()
	type fields struct {
		TestKeys                         urlgetter.TestKeys
		RegistrationServerFailure        *string
		RegistrationServerStatus         string
		WhatsappEndpointsBlocked         []string
		WhatsappEndpointsDNSInconsistent []string
		WhatsappEndpointsStatus          string
		WhatsappWebStatus                string
		WhatsappWebFailure               *string
		WhatsappHTTPFailure              *string
		WhatsappHTTPSFailure             *string
	}
	tests := []struct {
		name    string
		fields  fields
		failure *string
		status  string
	}{{
		name:    "with success",
		failure: nil,
		status:  "ok",
	}, {
		name: "with HTTP failure",
		fields: fields{
			WhatsappHTTPFailure: &errorString,
		},
		failure: &errorString,
		status:  "blocked",
	}, {
		name: "with HTTPS failure",
		fields: fields{
			WhatsappHTTPSFailure: &errorString,
		},
		failure: &errorString,
		status:  "blocked",
	}, {
		name: "with both HTTP and HTTPS failure",
		fields: fields{
			WhatsappHTTPFailure:  &errorString,
			WhatsappHTTPSFailure: &secondErrorString,
		},
		failure: &secondErrorString,
		status:  "blocked",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tk := &whatsapp.TestKeys{
				TestKeys:                         tt.fields.TestKeys,
				RegistrationServerFailure:        tt.fields.RegistrationServerFailure,
				RegistrationServerStatus:         tt.fields.RegistrationServerStatus,
				WhatsappEndpointsBlocked:         tt.fields.WhatsappEndpointsBlocked,
				WhatsappEndpointsDNSInconsistent: tt.fields.WhatsappEndpointsDNSInconsistent,
				WhatsappEndpointsStatus:          tt.fields.WhatsappEndpointsStatus,
				WhatsappWebStatus:                tt.fields.WhatsappWebStatus,
				WhatsappWebFailure:               tt.fields.WhatsappWebFailure,
				WhatsappHTTPFailure:              tt.fields.WhatsappHTTPFailure,
				WhatsappHTTPSFailure:             tt.fields.WhatsappHTTPSFailure,
			}
			tk.ComputeWebStatus()
			diff := cmp.Diff(tk.WhatsappWebFailure, tt.failure)
			if diff != "" {
				t.Fatal(diff)
			}
			diff = cmp.Diff(tk.WhatsappWebStatus, tt.status)
			if diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestTestKeysMixedEndpointsFailure(t *testing.T) {
	failure := io.EOF.Error()
	tk := whatsapp.NewTestKeys()
	tk.Update(urlgetter.MultiOutput{
		Input:    urlgetter.MultiInput{Target: "tcpconnect://e7.whatsapp.net:443"},
		TestKeys: urlgetter.TestKeys{Failure: &failure},
	})
	tk.Update(urlgetter.MultiOutput{
		Input:    urlgetter.MultiInput{Target: "tcpconnect://e7.whatsapp.net:5222"},
		TestKeys: urlgetter.TestKeys{},
	})
	tk.Update(urlgetter.MultiOutput{
		Input:    urlgetter.MultiInput{Target: whatsapp.RegistrationServiceURL},
		TestKeys: urlgetter.TestKeys{},
	})
	tk.Update(urlgetter.MultiOutput{
		Input:    urlgetter.MultiInput{Target: whatsapp.WebHTTPSURL},
		TestKeys: urlgetter.TestKeys{},
	})
	tk.Update(urlgetter.MultiOutput{
		Input: urlgetter.MultiInput{Target: whatsapp.WebHTTPURL},
		TestKeys: urlgetter.TestKeys{
			HTTPResponseStatus:    302,
			HTTPResponseLocations: []string{whatsapp.WebHTTPSURL},
		},
	})
	tk.ComputeWebStatus()
	if tk.RegistrationServerFailure != nil {
		t.Fatal("invalid RegistrationServerFailure")
	}
	if tk.RegistrationServerStatus != "ok" {
		t.Fatal("invalid RegistrationServerStatus")
	}
	if len(tk.WhatsappEndpointsBlocked) != 0 {
		t.Fatal("invalid WhatsappEndpointsBlocked")
	}
	if len(tk.WhatsappEndpointsDNSInconsistent) != 0 {
		t.Fatal("invalid WhatsappEndpointsDNSInconsistent")
	}
	if tk.WhatsappEndpointsStatus != "ok" {
		t.Fatal("invalid WhatsappEndpointsStatus")
	}
	if tk.WhatsappWebFailure != nil {
		t.Fatal("invalid WhatsappWebFailure")
	}
	if tk.WhatsappWebStatus != "ok" {
		t.Fatal("invalid WhatsappWebStatus")
	}
}

func TestTestKeysOnlyEndpointsFailure(t *testing.T) {
	failure := io.EOF.Error()
	tk := whatsapp.NewTestKeys()
	tk.Update(urlgetter.MultiOutput{
		Input:    urlgetter.MultiInput{Target: "tcpconnect://e7.whatsapp.net:443"},
		TestKeys: urlgetter.TestKeys{Failure: &failure},
	})
	tk.Update(urlgetter.MultiOutput{
		Input:    urlgetter.MultiInput{Target: "tcpconnect://e7.whatsapp.net:5222"},
		TestKeys: urlgetter.TestKeys{Failure: &failure},
	})
	tk.Update(urlgetter.MultiOutput{
		Input:    urlgetter.MultiInput{Target: whatsapp.RegistrationServiceURL},
		TestKeys: urlgetter.TestKeys{},
	})
	tk.Update(urlgetter.MultiOutput{
		Input:    urlgetter.MultiInput{Target: whatsapp.WebHTTPSURL},
		TestKeys: urlgetter.TestKeys{},
	})
	tk.Update(urlgetter.MultiOutput{
		Input: urlgetter.MultiInput{Target: whatsapp.WebHTTPURL},
		TestKeys: urlgetter.TestKeys{
			HTTPResponseStatus:    302,
			HTTPResponseLocations: []string{whatsapp.WebHTTPSURL},
		},
	})
	tk.ComputeWebStatus()
	if tk.RegistrationServerFailure != nil {
		t.Fatal("invalid RegistrationServerFailure")
	}
	if tk.RegistrationServerStatus != "ok" {
		t.Fatal("invalid RegistrationServerStatus")
	}
	if len(tk.WhatsappEndpointsBlocked) != 1 {
		t.Fatal("invalid WhatsappEndpointsBlocked")
	}
	if len(tk.WhatsappEndpointsDNSInconsistent) != 0 {
		t.Fatal("invalid WhatsappEndpointsDNSInconsistent")
	}
	if tk.WhatsappEndpointsStatus != "blocked" {
		t.Fatal("invalid WhatsappEndpointsStatus")
	}
	if tk.WhatsappWebFailure != nil {
		t.Fatal("invalid WhatsappWebFailure")
	}
	if tk.WhatsappWebStatus != "ok" {
		t.Fatal("invalid WhatsappWebStatus")
	}
}

func TestTestKeysOnlyRegistrationServerFailure(t *testing.T) {
	failure := io.EOF.Error()
	tk := whatsapp.NewTestKeys()
	tk.Update(urlgetter.MultiOutput{
		Input:    urlgetter.MultiInput{Target: "tcpconnect://e7.whatsapp.net:443"},
		TestKeys: urlgetter.TestKeys{},
	})
	tk.Update(urlgetter.MultiOutput{
		Input:    urlgetter.MultiInput{Target: whatsapp.RegistrationServiceURL},
		TestKeys: urlgetter.TestKeys{Failure: &failure},
	})
	tk.Update(urlgetter.MultiOutput{
		Input:    urlgetter.MultiInput{Target: whatsapp.WebHTTPSURL},
		TestKeys: urlgetter.TestKeys{},
	})
	tk.Update(urlgetter.MultiOutput{
		Input: urlgetter.MultiInput{Target: whatsapp.WebHTTPURL},
		TestKeys: urlgetter.TestKeys{
			HTTPResponseStatus:    302,
			HTTPResponseLocations: []string{whatsapp.WebHTTPSURL},
		},
	})
	tk.ComputeWebStatus()
	if *tk.RegistrationServerFailure != failure {
		t.Fatal("invalid RegistrationServerFailure")
	}
	if tk.RegistrationServerStatus != "blocked" {
		t.Fatal("invalid RegistrationServerStatus")
	}
	if len(tk.WhatsappEndpointsBlocked) != 0 {
		t.Fatal("invalid WhatsappEndpointsBlocked")
	}
	if len(tk.WhatsappEndpointsDNSInconsistent) != 0 {
		t.Fatal("invalid WhatsappEndpointsDNSInconsistent")
	}
	if tk.WhatsappEndpointsStatus != "ok" {
		t.Fatal("invalid WhatsappEndpointsStatus")
	}
	if tk.WhatsappWebFailure != nil {
		t.Fatal("invalid WhatsappWebFailure")
	}
	if tk.WhatsappWebStatus != "ok" {
		t.Fatal("invalid WhatsappWebStatus")
	}
}

func TestTestKeysOnlyWebHTTPSFailure(t *testing.T) {
	failure := io.EOF.Error()
	tk := whatsapp.NewTestKeys()
	tk.Update(urlgetter.MultiOutput{
		Input:    urlgetter.MultiInput{Target: "tcpconnect://e7.whatsapp.net:443"},
		TestKeys: urlgetter.TestKeys{},
	})
	tk.Update(urlgetter.MultiOutput{
		Input:    urlgetter.MultiInput{Target: whatsapp.RegistrationServiceURL},
		TestKeys: urlgetter.TestKeys{},
	})
	tk.Update(urlgetter.MultiOutput{
		Input:    urlgetter.MultiInput{Target: whatsapp.WebHTTPSURL},
		TestKeys: urlgetter.TestKeys{Failure: &failure},
	})
	tk.Update(urlgetter.MultiOutput{
		Input: urlgetter.MultiInput{Target: whatsapp.WebHTTPURL},
		TestKeys: urlgetter.TestKeys{
			HTTPResponseStatus:    302,
			HTTPResponseLocations: []string{whatsapp.WebHTTPSURL},
		},
	})
	tk.ComputeWebStatus()
	if tk.RegistrationServerFailure != nil {
		t.Fatal("invalid RegistrationServerFailure")
	}
	if tk.RegistrationServerStatus != "ok" {
		t.Fatal("invalid RegistrationServerStatus")
	}
	if len(tk.WhatsappEndpointsBlocked) != 0 {
		t.Fatal("invalid WhatsappEndpointsBlocked")
	}
	if len(tk.WhatsappEndpointsDNSInconsistent) != 0 {
		t.Fatal("invalid WhatsappEndpointsDNSInconsistent")
	}
	if tk.WhatsappEndpointsStatus != "ok" {
		t.Fatal("invalid WhatsappEndpointsStatus")
	}
	if *tk.WhatsappWebFailure != failure {
		t.Fatal("invalid WhatsappWebFailure")
	}
	if tk.WhatsappWebStatus != "blocked" {
		t.Fatal("invalid WhatsappWebStatus")
	}
}

func TestTestKeysOnlyWebHTTPFailureNo302(t *testing.T) {
	tk := whatsapp.NewTestKeys()
	tk.Update(urlgetter.MultiOutput{
		Input:    urlgetter.MultiInput{Target: "tcpconnect://e7.whatsapp.net:443"},
		TestKeys: urlgetter.TestKeys{},
	})
	tk.Update(urlgetter.MultiOutput{
		Input:    urlgetter.MultiInput{Target: whatsapp.RegistrationServiceURL},
		TestKeys: urlgetter.TestKeys{},
	})
	tk.Update(urlgetter.MultiOutput{
		Input:    urlgetter.MultiInput{Target: whatsapp.WebHTTPSURL},
		TestKeys: urlgetter.TestKeys{},
	})
	tk.Update(urlgetter.MultiOutput{
		Input: urlgetter.MultiInput{Target: whatsapp.WebHTTPURL},
		TestKeys: urlgetter.TestKeys{
			HTTPResponseStatus: 400,
		},
	})
	tk.ComputeWebStatus()
	if tk.RegistrationServerFailure != nil {
		t.Fatal("invalid RegistrationServerFailure")
	}
	if tk.RegistrationServerStatus != "ok" {
		t.Fatal("invalid RegistrationServerStatus")
	}
	if len(tk.WhatsappEndpointsBlocked) != 0 {
		t.Fatal("invalid WhatsappEndpointsBlocked")
	}
	if len(tk.WhatsappEndpointsDNSInconsistent) != 0 {
		t.Fatal("invalid WhatsappEndpointsDNSInconsistent")
	}
	if tk.WhatsappEndpointsStatus != "ok" {
		t.Fatal("invalid WhatsappEndpointsStatus")
	}
	if *tk.WhatsappWebFailure != httpfailure.UnexpectedStatusCode {
		t.Fatal("invalid WhatsappWebFailure")
	}
	if tk.WhatsappWebStatus != "blocked" {
		t.Fatal("invalid WhatsappWebStatus")
	}
}

func TestTestKeysOnlyWebHTTPFailureNoLocations(t *testing.T) {
	tk := whatsapp.NewTestKeys()
	tk.Update(urlgetter.MultiOutput{
		Input:    urlgetter.MultiInput{Target: "tcpconnect://e7.whatsapp.net:443"},
		TestKeys: urlgetter.TestKeys{},
	})
	tk.Update(urlgetter.MultiOutput{
		Input:    urlgetter.MultiInput{Target: whatsapp.RegistrationServiceURL},
		TestKeys: urlgetter.TestKeys{},
	})
	tk.Update(urlgetter.MultiOutput{
		Input:    urlgetter.MultiInput{Target: whatsapp.WebHTTPSURL},
		TestKeys: urlgetter.TestKeys{},
	})
	tk.Update(urlgetter.MultiOutput{
		Input: urlgetter.MultiInput{Target: whatsapp.WebHTTPURL},
		TestKeys: urlgetter.TestKeys{
			HTTPResponseStatus:    302,
			HTTPResponseLocations: nil,
		},
	})
	tk.ComputeWebStatus()
	if tk.RegistrationServerFailure != nil {
		t.Fatal("invalid RegistrationServerFailure")
	}
	if tk.RegistrationServerStatus != "ok" {
		t.Fatal("invalid RegistrationServerStatus")
	}
	if len(tk.WhatsappEndpointsBlocked) != 0 {
		t.Fatal("invalid WhatsappEndpointsBlocked")
	}
	if len(tk.WhatsappEndpointsDNSInconsistent) != 0 {
		t.Fatal("invalid WhatsappEndpointsDNSInconsistent")
	}
	if tk.WhatsappEndpointsStatus != "ok" {
		t.Fatal("invalid WhatsappEndpointsStatus")
	}
	if *tk.WhatsappWebFailure != httpfailure.UnexpectedRedirectURL {
		t.Fatal("invalid WhatsappWebFailure")
	}
	if tk.WhatsappWebStatus != "blocked" {
		t.Fatal("invalid WhatsappWebStatus")
	}
}

func TestTestKeysOnlyWebHTTPFailureNotExpectedURL(t *testing.T) {
	tk := whatsapp.NewTestKeys()
	tk.Update(urlgetter.MultiOutput{
		Input:    urlgetter.MultiInput{Target: "tcpconnect://e7.whatsapp.net:443"},
		TestKeys: urlgetter.TestKeys{},
	})
	tk.Update(urlgetter.MultiOutput{
		Input:    urlgetter.MultiInput{Target: whatsapp.RegistrationServiceURL},
		TestKeys: urlgetter.TestKeys{},
	})
	tk.Update(urlgetter.MultiOutput{
		Input:    urlgetter.MultiInput{Target: whatsapp.WebHTTPSURL},
		TestKeys: urlgetter.TestKeys{},
	})
	tk.Update(urlgetter.MultiOutput{
		Input: urlgetter.MultiInput{Target: whatsapp.WebHTTPURL},
		TestKeys: urlgetter.TestKeys{
			HTTPResponseStatus:    302,
			HTTPResponseLocations: []string{"https://x.org/"},
		},
	})
	tk.ComputeWebStatus()
	if tk.RegistrationServerFailure != nil {
		t.Fatal("invalid RegistrationServerFailure")
	}
	if tk.RegistrationServerStatus != "ok" {
		t.Fatal("invalid RegistrationServerStatus")
	}
	if len(tk.WhatsappEndpointsBlocked) != 0 {
		t.Fatal("invalid WhatsappEndpointsBlocked")
	}
	if len(tk.WhatsappEndpointsDNSInconsistent) != 0 {
		t.Fatal("invalid WhatsappEndpointsDNSInconsistent")
	}
	if tk.WhatsappEndpointsStatus != "ok" {
		t.Fatal("invalid WhatsappEndpointsStatus")
	}
	if *tk.WhatsappWebFailure != httpfailure.UnexpectedRedirectURL {
		t.Fatal("invalid WhatsappWebFailure")
	}
	if tk.WhatsappWebStatus != "blocked" {
		t.Fatal("invalid WhatsappWebStatus")
	}
}

func TestTestKeysOnlyWebHTTPFailureTooManyURLs(t *testing.T) {
	tk := whatsapp.NewTestKeys()
	tk.Update(urlgetter.MultiOutput{
		Input:    urlgetter.MultiInput{Target: "tcpconnect://e7.whatsapp.net:443"},
		TestKeys: urlgetter.TestKeys{},
	})
	tk.Update(urlgetter.MultiOutput{
		Input:    urlgetter.MultiInput{Target: whatsapp.RegistrationServiceURL},
		TestKeys: urlgetter.TestKeys{},
	})
	tk.Update(urlgetter.MultiOutput{
		Input:    urlgetter.MultiInput{Target: whatsapp.WebHTTPSURL},
		TestKeys: urlgetter.TestKeys{},
	})
	tk.Update(urlgetter.MultiOutput{
		Input: urlgetter.MultiInput{Target: whatsapp.WebHTTPURL},
		TestKeys: urlgetter.TestKeys{
			HTTPResponseStatus:    302,
			HTTPResponseLocations: []string{whatsapp.WebHTTPSURL, "https://x.org/"},
		},
	})
	tk.ComputeWebStatus()
	if tk.RegistrationServerFailure != nil {
		t.Fatal("invalid RegistrationServerFailure")
	}
	if tk.RegistrationServerStatus != "ok" {
		t.Fatal("invalid RegistrationServerStatus")
	}
	if len(tk.WhatsappEndpointsBlocked) != 0 {
		t.Fatal("invalid WhatsappEndpointsBlocked")
	}
	if len(tk.WhatsappEndpointsDNSInconsistent) != 0 {
		t.Fatal("invalid WhatsappEndpointsDNSInconsistent")
	}
	if tk.WhatsappEndpointsStatus != "ok" {
		t.Fatal("invalid WhatsappEndpointsStatus")
	}
	if *tk.WhatsappWebFailure != httpfailure.UnexpectedRedirectURL {
		t.Fatal("invalid WhatsappWebFailure")
	}
	if tk.WhatsappWebStatus != "blocked" {
		t.Fatal("invalid WhatsappWebStatus")
	}
}

func TestWeConfigureWebChecksCorrectly(t *testing.T) {
	called := atomicx.NewInt64()
	emptyConfig := urlgetter.Config{}
	configWithFailOnHTTPError := urlgetter.Config{FailOnHTTPError: true}
	configWithNoFollowRedirects := urlgetter.Config{NoFollowRedirects: true}
	measurer := whatsapp.Measurer{
		Config: whatsapp.Config{},
		Getter: func(ctx context.Context, g urlgetter.Getter) (urlgetter.TestKeys, error) {
			switch g.Target {
			case whatsapp.WebHTTPSURL:
				called.Add(1)
				if diff := cmp.Diff(g.Config, emptyConfig); diff != "" {
					panic(diff)
				}
			case whatsapp.WebHTTPURL:
				called.Add(2)
				if diff := cmp.Diff(g.Config, configWithNoFollowRedirects); diff != "" {
					panic(diff)
				}
			case whatsapp.RegistrationServiceURL:
				called.Add(4)
				if diff := cmp.Diff(g.Config, configWithFailOnHTTPError); diff != "" {
					panic(diff)
				}
			default:
				called.Add(8)
				if diff := cmp.Diff(g.Config, emptyConfig); diff != "" {
					panic(diff)
				}
			}
			return urlgetter.DefaultMultiGetter(ctx, g)
		},
	}
	ctx := context.Background()
	sess := &mockable.ExperimentSession{
		MockableLogger: log.Log,
	}
	measurement := new(model.Measurement)
	callbacks := handler.NewPrinterCallbacks(log.Log)
	if err := measurer.Run(ctx, sess, measurement, callbacks); err != nil {
		t.Fatal(err)
	}
	if called.Load() != 263 {
		t.Fatal("not called the expected number of times")
	}
}
