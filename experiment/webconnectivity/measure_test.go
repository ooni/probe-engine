package webconnectivity_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/ooni/probe-engine/experiment/urlgetter"
	"github.com/ooni/probe-engine/experiment/webconnectivity"
	"github.com/ooni/probe-engine/netx/modelx"
)

func TestDNSExperimentFailure(t *testing.T) {
	failure := "mocked_error"
	resolveOperation := modelx.ResolveOperation
	tlsHandshakeOperation := modelx.TLSHandshakeOperation
	connectOperation := modelx.ConnectOperation
	httpRoundTripOperation := modelx.HTTPRoundTripOperation
	toplevelOperation := modelx.TopLevelOperation
	type args struct {
		tk *webconnectivity.TestKeys
	}
	tests := []struct {
		name    string
		args    args
		wantOut *string
	}{{
		name: "with no error",
		args: args{
			tk: &webconnectivity.TestKeys{},
		},
	}, {
		name: "with DNS failure",
		args: args{
			tk: &webconnectivity.TestKeys{
				TestKeys: urlgetter.TestKeys{
					Failure:         &failure,
					FailedOperation: &resolveOperation,
				},
			},
		},
		wantOut: &failure,
	}, {
		name: "with TLS handshake failure",
		args: args{
			tk: &webconnectivity.TestKeys{
				TestKeys: urlgetter.TestKeys{
					Failure:         &failure,
					FailedOperation: &tlsHandshakeOperation,
				},
			},
		},
		wantOut: nil,
	}, {
		name: "with TCP connect failure",
		args: args{
			tk: &webconnectivity.TestKeys{
				TestKeys: urlgetter.TestKeys{
					Failure:         &failure,
					FailedOperation: &connectOperation,
				},
			},
		},
		wantOut: nil,
	}, {
		name: "with HTTP round trip failure",
		args: args{
			tk: &webconnectivity.TestKeys{
				TestKeys: urlgetter.TestKeys{
					Failure:         &failure,
					FailedOperation: &httpRoundTripOperation,
				},
			},
		},
		wantOut: nil,
	}, {
		name: "with toplevel failure",
		args: args{
			tk: &webconnectivity.TestKeys{
				TestKeys: urlgetter.TestKeys{
					Failure:         &failure,
					FailedOperation: &toplevelOperation,
				},
			},
		},
		wantOut: nil,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOut := webconnectivity.DNSExperimentFailure(tt.args.tk)
			if diff := cmp.Diff(tt.wantOut, gotOut); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestHTTPExperimentFailure(t *testing.T) {
	failure := "mocked_error"
	resolveOperation := modelx.ResolveOperation
	tlsHandshakeOperation := modelx.TLSHandshakeOperation
	connectOperation := modelx.ConnectOperation
	httpRoundTripOperation := modelx.HTTPRoundTripOperation
	toplevelOperation := modelx.TopLevelOperation
	type args struct {
		tk *webconnectivity.TestKeys
	}
	tests := []struct {
		name    string
		args    args
		wantOut *string
	}{{
		name: "with no error",
		args: args{
			tk: &webconnectivity.TestKeys{},
		},
	}, {
		name: "with DNS failure",
		args: args{
			tk: &webconnectivity.TestKeys{
				TestKeys: urlgetter.TestKeys{
					Failure:         &failure,
					FailedOperation: &resolveOperation,
				},
			},
		},
		wantOut: nil,
	}, {
		name: "with TLS handshake failure",
		args: args{
			tk: &webconnectivity.TestKeys{
				TestKeys: urlgetter.TestKeys{
					Failure:         &failure,
					FailedOperation: &tlsHandshakeOperation,
				},
			},
		},
		wantOut: &failure,
	}, {
		name: "with TCP connect failure",
		args: args{
			tk: &webconnectivity.TestKeys{
				TestKeys: urlgetter.TestKeys{
					Failure:         &failure,
					FailedOperation: &connectOperation,
				},
			},
		},
		wantOut: &failure,
	}, {
		name: "with HTTP round trip failure",
		args: args{
			tk: &webconnectivity.TestKeys{
				TestKeys: urlgetter.TestKeys{
					Failure:         &failure,
					FailedOperation: &httpRoundTripOperation,
				},
			},
		},
		wantOut: &failure,
	}, {
		name: "with toplevel failure",
		args: args{
			tk: &webconnectivity.TestKeys{
				TestKeys: urlgetter.TestKeys{
					Failure:         &failure,
					FailedOperation: &toplevelOperation,
				},
			},
		},
		wantOut: nil,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOut := webconnectivity.HTTPExperimentFailure(tt.args.tk)
			if diff := cmp.Diff(tt.wantOut, gotOut); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
