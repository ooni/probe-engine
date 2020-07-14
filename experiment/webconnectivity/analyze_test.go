package webconnectivity_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/ooni/probe-engine/experiment/urlgetter"
	"github.com/ooni/probe-engine/experiment/webconnectivity"
	"github.com/ooni/probe-engine/internal/randx"
	"github.com/ooni/probe-engine/netx/archival"
)

func TestAnalyzeInvalidURL(t *testing.T) {
	out := webconnectivity.Analyze("\t\t\t", nil)
	if out.BodyLengthMatch != nil {
		t.Fatal("unexpected BodyLengthMatch")
	}
	if out.HeadersMatch != nil {
		t.Fatal("unexpected HeadersMatch")
	}
	if out.StatusCodeMatch != nil {
		t.Fatal("unexpected StatusCodeMatch")
	}
	if out.TitleMatch != nil {
		t.Fatal("unexpected TitleMatch")
	}
	if out.Accessible != nil {
		t.Fatal("unexpected Accessible")
	}
	if out.Blocking != nil {
		t.Fatal("unexpected Blocking")
	}
}

func TestBodyLengthChecks(t *testing.T) {
	var (
		trueValue  = true
		falseValue = false
	)
	type args struct {
		tk *webconnectivity.TestKeys
	}
	tests := []struct {
		name        string
		args        args
		lengthMatch *bool
		proportion  *float64
	}{{
		name: "nothing",
		args: args{
			tk: &webconnectivity.TestKeys{},
		},
		lengthMatch: nil,
	}, {
		name: "control length is nonzero",
		args: args{
			tk: &webconnectivity.TestKeys{
				Control: webconnectivity.ControlResponse{
					HTTPRequest: webconnectivity.ControlHTTPRequestResult{
						BodyLength: 1024,
					},
				},
			},
		},
		lengthMatch: nil,
	}, {
		name: "response body is truncated",
		args: args{
			tk: &webconnectivity.TestKeys{
				TestKeys: urlgetter.TestKeys{
					Requests: []archival.RequestEntry{{
						Response: archival.HTTPResponse{
							BodyIsTruncated: true,
						},
					}},
				},
				Control: webconnectivity.ControlResponse{
					HTTPRequest: webconnectivity.ControlHTTPRequestResult{
						BodyLength: 1024,
					},
				},
			},
		},
		lengthMatch: nil,
	}, {
		name: "response body length is zero",
		args: args{
			tk: &webconnectivity.TestKeys{
				TestKeys: urlgetter.TestKeys{
					Requests: []archival.RequestEntry{{
						Response: archival.HTTPResponse{},
					}},
				},
				Control: webconnectivity.ControlResponse{
					HTTPRequest: webconnectivity.ControlHTTPRequestResult{
						BodyLength: 1024,
					},
				},
			},
		},
		lengthMatch: nil,
	}, {
		name: "match with bigger control",
		args: args{
			tk: &webconnectivity.TestKeys{
				TestKeys: urlgetter.TestKeys{
					Requests: []archival.RequestEntry{{
						Response: archival.HTTPResponse{
							Body: archival.MaybeBinaryValue{
								Value: randx.Letters(768),
							},
						},
					}},
				},
				Control: webconnectivity.ControlResponse{
					HTTPRequest: webconnectivity.ControlHTTPRequestResult{
						BodyLength: 1024,
					},
				},
			},
		},
		lengthMatch: &trueValue,
		proportion:  (func() *float64 { v := 0.75; return &v })(),
	}, {
		name: "match with bigger measurement",
		args: args{
			tk: &webconnectivity.TestKeys{
				TestKeys: urlgetter.TestKeys{
					Requests: []archival.RequestEntry{{
						Response: archival.HTTPResponse{
							Body: archival.MaybeBinaryValue{
								Value: randx.Letters(1024),
							},
						},
					}},
				},
				Control: webconnectivity.ControlResponse{
					HTTPRequest: webconnectivity.ControlHTTPRequestResult{
						BodyLength: 768,
					},
				},
			},
		},
		lengthMatch: &trueValue,
		proportion:  (func() *float64 { v := 0.75; return &v })(),
	}, {
		name: "not match with bigger control",
		args: args{
			tk: &webconnectivity.TestKeys{
				TestKeys: urlgetter.TestKeys{
					Requests: []archival.RequestEntry{{
						Response: archival.HTTPResponse{
							Body: archival.MaybeBinaryValue{
								Value: randx.Letters(8),
							},
						},
					}},
				},
				Control: webconnectivity.ControlResponse{
					HTTPRequest: webconnectivity.ControlHTTPRequestResult{
						BodyLength: 16,
					},
				},
			},
		},
		lengthMatch: &falseValue,
		proportion:  (func() *float64 { v := 0.5; return &v })(),
	}, {
		name: "match with bigger measurement",
		args: args{
			tk: &webconnectivity.TestKeys{
				TestKeys: urlgetter.TestKeys{
					Requests: []archival.RequestEntry{{
						Response: archival.HTTPResponse{
							Body: archival.MaybeBinaryValue{
								Value: randx.Letters(16),
							},
						},
					}},
				},
				Control: webconnectivity.ControlResponse{
					HTTPRequest: webconnectivity.ControlHTTPRequestResult{
						BodyLength: 8,
					},
				},
			},
		},
		lengthMatch: &falseValue,
		proportion:  (func() *float64 { v := 0.5; return &v })(),
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, proportion := webconnectivity.BodyLengthChecks(tt.args.tk)
			if diff := cmp.Diff(tt.lengthMatch, match); diff != "" {
				t.Fatal(diff)
			}
			if diff := cmp.Diff(tt.proportion, proportion); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestStatusCodeMatch(t *testing.T) {
	var (
		trueValue  = true
		falseValue = false
	)
	type args struct {
		tk *webconnectivity.TestKeys
	}
	tests := []struct {
		name    string
		args    args
		wantOut *bool
	}{{
		name: "with all zero",
		args: args{
			tk: &webconnectivity.TestKeys{},
		},
	}, {
		name: "with a request but zero status codes",
		args: args{
			tk: &webconnectivity.TestKeys{
				TestKeys: urlgetter.TestKeys{
					Requests: []archival.RequestEntry{{}},
				},
			},
		},
	}, {
		name: "with equal status codes including 5xx",
		args: args{
			tk: &webconnectivity.TestKeys{
				TestKeys: urlgetter.TestKeys{
					Requests: []archival.RequestEntry{{
						Response: archival.HTTPResponse{
							Code: 501,
						},
					}},
				},
				Control: webconnectivity.ControlResponse{
					HTTPRequest: webconnectivity.ControlHTTPRequestResult{
						StatusCode: 501,
					},
				},
			},
		},
		wantOut: &trueValue,
	}, {
		name: "with different status codes and the control being 5xx",
		args: args{
			tk: &webconnectivity.TestKeys{
				TestKeys: urlgetter.TestKeys{
					Requests: []archival.RequestEntry{{
						Response: archival.HTTPResponse{
							Code: 407,
						},
					}},
				},
				Control: webconnectivity.ControlResponse{
					HTTPRequest: webconnectivity.ControlHTTPRequestResult{
						StatusCode: 501,
					},
				},
			},
		},
		wantOut: nil,
	}, {
		name: "with different status codes and the control being not 5xx",
		args: args{
			tk: &webconnectivity.TestKeys{
				TestKeys: urlgetter.TestKeys{
					Requests: []archival.RequestEntry{{
						Response: archival.HTTPResponse{
							Code: 407,
						},
					}},
				},
				Control: webconnectivity.ControlResponse{
					HTTPRequest: webconnectivity.ControlHTTPRequestResult{
						StatusCode: 200,
					},
				},
			},
		},
		wantOut: &falseValue,
	}, {
		name: "with only response status code and no control status code",
		args: args{
			tk: &webconnectivity.TestKeys{
				TestKeys: urlgetter.TestKeys{
					Requests: []archival.RequestEntry{{
						Response: archival.HTTPResponse{
							Code: 200,
						},
					}},
				},
			},
		},
	}, {
		name: "with only control status code and no response status code",
		args: args{
			tk: &webconnectivity.TestKeys{
				TestKeys: urlgetter.TestKeys{
					Requests: []archival.RequestEntry{{
						Response: archival.HTTPResponse{
							Code: 0,
						},
					}},
				},
				Control: webconnectivity.ControlResponse{
					HTTPRequest: webconnectivity.ControlHTTPRequestResult{
						StatusCode: 200,
					},
				},
			},
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOut := webconnectivity.StatusCodeMatch(tt.args.tk)
			if diff := cmp.Diff(tt.wantOut, gotOut); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestHeadersMatch(t *testing.T) {
	var (
		trueValue  = true
		falseValue = false
	)
	type args struct {
		tk *webconnectivity.TestKeys
	}
	tests := []struct {
		name string
		args args
		want *bool
	}{{
		name: "with no requests",
		args: args{
			tk: &webconnectivity.TestKeys{
				Control: webconnectivity.ControlResponse{
					HTTPRequest: webconnectivity.ControlHTTPRequestResult{
						Headers: map[string]string{
							"Date":   "Mon Jul 13 21:05:43 CEST 2020",
							"Antani": "Mascetti",
						},
						StatusCode: 200,
					},
				},
			},
		},
		want: nil,
	}, {
		name: "with basically nothing",
		args: args{
			tk: &webconnectivity.TestKeys{},
		},
		want: nil,
	}, {
		name: "with request and no response status code",
		args: args{
			tk: &webconnectivity.TestKeys{
				TestKeys: urlgetter.TestKeys{
					Requests: []archival.RequestEntry{{}},
				},
				Control: webconnectivity.ControlResponse{
					HTTPRequest: webconnectivity.ControlHTTPRequestResult{
						Headers: map[string]string{
							"Date":   "Mon Jul 13 21:05:43 CEST 2020",
							"Antani": "Mascetti",
						},
						StatusCode: 200,
					},
				},
			},
		},
		want: nil,
	}, {
		name: "with no control status code",
		args: args{
			tk: &webconnectivity.TestKeys{
				TestKeys: urlgetter.TestKeys{
					Requests: []archival.RequestEntry{{
						Response: archival.HTTPResponse{
							Headers: map[string]archival.MaybeBinaryValue{
								"Date": {Value: "Mon Jul 13 21:10:08 CEST 2020"},
							},
							Code: 200,
						},
					}},
				},
				Control: webconnectivity.ControlResponse{},
			},
		},
		want: nil,
	}, {
		name: "with no uncommon headers",
		args: args{
			tk: &webconnectivity.TestKeys{
				TestKeys: urlgetter.TestKeys{
					Requests: []archival.RequestEntry{{
						Response: archival.HTTPResponse{
							Headers: map[string]archival.MaybeBinaryValue{
								"Date": {Value: "Mon Jul 13 21:10:08 CEST 2020"},
							},
							Code: 200,
						},
					}},
				},
				Control: webconnectivity.ControlResponse{
					HTTPRequest: webconnectivity.ControlHTTPRequestResult{
						Headers: map[string]string{
							"Date": "Mon Jul 13 21:05:43 CEST 2020",
						},
						StatusCode: 200,
					},
				},
			},
		},
		want: &trueValue,
	}, {
		name: "with equal uncommon headers",
		args: args{
			tk: &webconnectivity.TestKeys{
				TestKeys: urlgetter.TestKeys{
					Requests: []archival.RequestEntry{{
						Response: archival.HTTPResponse{
							Headers: map[string]archival.MaybeBinaryValue{
								"Date":   {Value: "Mon Jul 13 21:10:08 CEST 2020"},
								"Antani": {Value: "MASCETTI"},
							},
							Code: 200,
						},
					}},
				},
				Control: webconnectivity.ControlResponse{
					HTTPRequest: webconnectivity.ControlHTTPRequestResult{
						Headers: map[string]string{
							"Date":   "Mon Jul 13 21:05:43 CEST 2020",
							"Antani": "MELANDRI",
						},
						StatusCode: 200,
					},
				},
			},
		},
		want: &trueValue,
	}, {
		name: "with different uncommon headers",
		args: args{
			tk: &webconnectivity.TestKeys{
				TestKeys: urlgetter.TestKeys{
					Requests: []archival.RequestEntry{{
						Response: archival.HTTPResponse{
							Headers: map[string]archival.MaybeBinaryValue{
								"Date":   {Value: "Mon Jul 13 21:10:08 CEST 2020"},
								"Antani": {Value: "MASCETTI"},
							},
							Code: 200,
						},
					}},
				},
				Control: webconnectivity.ControlResponse{
					HTTPRequest: webconnectivity.ControlHTTPRequestResult{
						Headers: map[string]string{
							"Date":     "Mon Jul 13 21:05:43 CEST 2020",
							"Melandri": "MASCETTI",
						},
						StatusCode: 200,
					},
				},
			},
		},
		want: &falseValue,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := webconnectivity.HeadersMatch(tt.args.tk)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestTitleMatch(t *testing.T) {
	var (
		trueValue  = true
		falseValue = false
	)
	type args struct {
		tk *webconnectivity.TestKeys
	}
	tests := []struct {
		name    string
		args    args
		wantOut *bool
	}{{
		name: "with all empty",
		args: args{
			tk: &webconnectivity.TestKeys{},
		},
		wantOut: nil,
	}, {
		name: "with a request and no response",
		args: args{
			tk: &webconnectivity.TestKeys{
				TestKeys: urlgetter.TestKeys{
					Requests: []archival.RequestEntry{{}},
				},
			},
		},
		wantOut: nil,
	}, {
		name: "with a response with truncated body",
		args: args{
			tk: &webconnectivity.TestKeys{
				TestKeys: urlgetter.TestKeys{
					Requests: []archival.RequestEntry{{
						Response: archival.HTTPResponse{
							Code:            200,
							BodyIsTruncated: true,
						},
					}},
				},
			},
		},
		wantOut: nil,
	}, {
		name: "with a response with good body",
		args: args{
			tk: &webconnectivity.TestKeys{
				TestKeys: urlgetter.TestKeys{
					Requests: []archival.RequestEntry{{
						Response: archival.HTTPResponse{
							Code: 200,
							Body: archival.MaybeBinaryValue{Value: "<HTML/>"},
						},
					}},
				},
			},
		},
		wantOut: nil,
	}, {
		name: "with all good but no titles",
		args: args{
			tk: &webconnectivity.TestKeys{
				TestKeys: urlgetter.TestKeys{
					Requests: []archival.RequestEntry{{
						Response: archival.HTTPResponse{
							Code: 200,
							Body: archival.MaybeBinaryValue{Value: "<HTML/>"},
						},
					}},
				},
				Control: webconnectivity.ControlResponse{
					HTTPRequest: webconnectivity.ControlHTTPRequestResult{
						StatusCode: 200,
						Title:      "",
					},
				},
			},
		},
		wantOut: nil,
	}, {
		name: "reasonably common case where it succeeds",
		args: args{
			tk: &webconnectivity.TestKeys{
				TestKeys: urlgetter.TestKeys{
					Requests: []archival.RequestEntry{{
						Response: archival.HTTPResponse{
							Code: 200,
							Body: archival.MaybeBinaryValue{
								Value: "<HTML><TITLE>La community di MSN</TITLE></HTML>"},
						},
					}},
				},
				Control: webconnectivity.ControlResponse{
					HTTPRequest: webconnectivity.ControlHTTPRequestResult{
						StatusCode: 200,
						Title:      "MSN Community",
					},
				},
			},
		},
		wantOut: &trueValue,
	}, {
		name: "reasonably common case where it fails",
		args: args{
			tk: &webconnectivity.TestKeys{
				TestKeys: urlgetter.TestKeys{
					Requests: []archival.RequestEntry{{
						Response: archival.HTTPResponse{
							Code: 200,
							Body: archival.MaybeBinaryValue{
								Value: "<HTML><TITLE>La communità di MSN</TITLE></HTML>"},
						},
					}},
				},
				Control: webconnectivity.ControlResponse{
					HTTPRequest: webconnectivity.ControlHTTPRequestResult{
						StatusCode: 200,
						Title:      "MSN Community",
					},
				},
			},
		},
		wantOut: &falseValue,
	}, {
		name: "when the title is too long",
		args: args{
			tk: &webconnectivity.TestKeys{
				TestKeys: urlgetter.TestKeys{
					Requests: []archival.RequestEntry{{
						Response: archival.HTTPResponse{
							Code: 200,
							Body: archival.MaybeBinaryValue{
								Value: "<HTML><TITLE>" + randx.Letters(1024) + "</TITLE></HTML>"},
						},
					}},
				},
				Control: webconnectivity.ControlResponse{
					HTTPRequest: webconnectivity.ControlHTTPRequestResult{
						StatusCode: 200,
						Title:      "MSN Community",
					},
				},
			},
		},
		wantOut: nil,
	}, {
		name: "reasonably common case where it succeeds with case variations",
		args: args{
			tk: &webconnectivity.TestKeys{
				TestKeys: urlgetter.TestKeys{
					Requests: []archival.RequestEntry{{
						Response: archival.HTTPResponse{
							Code: 200,
							Body: archival.MaybeBinaryValue{
								Value: "<HTML><TiTLe>La commUNity di MSN</tITLE></HTML>"},
						},
					}},
				},
				Control: webconnectivity.ControlResponse{
					HTTPRequest: webconnectivity.ControlHTTPRequestResult{
						StatusCode: 200,
						Title:      "MSN COmmunity",
					},
				},
			},
		},
		wantOut: &trueValue,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOut := webconnectivity.TitleMatch(tt.args.tk)
			if diff := cmp.Diff(tt.wantOut, gotOut); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
