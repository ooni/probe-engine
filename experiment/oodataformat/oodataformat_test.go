package oodataformat

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/ooni/netx/model"
)

func TestTCPConnectListNil(t *testing.T) {
	tcpConnect := NewTCPConnectList(nil)
	if len(tcpConnect) != 0 {
		t.Fatal("unexpected list length")
	}
}

func TestTCPConnectListEmptyRoundTrips(t *testing.T) {
	tcpConnect := NewTCPConnectList([][]model.Measurement{
		nil, nil, nil,
	})
	if len(tcpConnect) != 0 {
		t.Fatal("unexpected list length")
	}
}

func TestTCPConnectListEmptyMeasurement(t *testing.T) {
	tcpConnect := NewTCPConnectList([][]model.Measurement{
		[]model.Measurement{},
	})
	if len(tcpConnect) != 0 {
		t.Fatal("unexpected list length")
	}
}

func TestTCPConnectListGood(t *testing.T) {
	tcpConnect := NewTCPConnectList([][]model.Measurement{[]model.Measurement{
		model.Measurement{
			Connect: &model.ConnectEvent{
				RemoteAddress: "1.1.1.1:53",
			},
		},
	}})
	if len(tcpConnect) != 1 {
		t.Fatal("unexpected list length")
	}
	e := tcpConnect[0]
	if e.IP != "1.1.1.1" {
		t.Fatal("unexpected IP")
	}
	if e.Port != 53 {
		t.Fatal("unexpected port")
	}
	if e.Status.Failure != nil {
		t.Fatal("expected nil failure")
	}
	if e.Status.Success != true {
		t.Fatal("unexpected value of success")
	}
}

func TestTCPConnectListNetworkError(t *testing.T) {
	tcpConnect := NewTCPConnectList([][]model.Measurement{[]model.Measurement{
		model.Measurement{
			Connect: &model.ConnectEvent{
				RemoteAddress: "1.1.1.1:53",
				Error:         errors.New("mocked error"),
			},
		},
	}})
	if len(tcpConnect) != 1 {
		t.Fatal("unexpected list length")
	}
	e := tcpConnect[0]
	if e.IP != "1.1.1.1" {
		t.Fatal("unexpected IP")
	}
	if e.Port != 53 {
		t.Fatal("unexpected port")
	}
	if e.Status.Failure == nil {
		t.Fatal("expected non-nil failure")
	}
	if e.Status.Success != false {
		t.Fatal("unexpected value of success")
	}
}

func TestTCPConnectListInvalidHostPort(t *testing.T) {
	tcpConnect := NewTCPConnectList([][]model.Measurement{[]model.Measurement{
		model.Measurement{
			Connect: &model.ConnectEvent{
				RemoteAddress: "antani",
				Error:         errors.New("mocked error"),
			},
		},
		model.Measurement{
			Connect: &model.ConnectEvent{
				RemoteAddress: "1.1.1.1:53",
				Error:         errors.New("mocked error"),
			},
		},
	}})
	// make sure we don't stop processing after first error
	if len(tcpConnect) != 1 {
		t.Fatal("unexpected list length")
	}
}

func TestTCPConnectListInvalidPortNumber(t *testing.T) {
	tcpConnect := NewTCPConnectList([][]model.Measurement{[]model.Measurement{
		model.Measurement{
			Connect: &model.ConnectEvent{
				RemoteAddress: "1.1.1.1:antani",
				Error:         errors.New("mocked error"),
			},
		},
		model.Measurement{
			Connect: &model.ConnectEvent{
				RemoteAddress: "1.1.1.1:53",
				Error:         errors.New("mocked error"),
			},
		},
	}})
	// make sure we don't stop processing after first error
	if len(tcpConnect) != 1 {
		t.Fatal("unexpected list length")
	}
}

func TestTCPConnectListLargePortNumber(t *testing.T) {
	tcpConnect := NewTCPConnectList([][]model.Measurement{[]model.Measurement{
		model.Measurement{
			Connect: &model.ConnectEvent{
				RemoteAddress: "1.1.1.1:65536",
				Error:         errors.New("mocked error"),
			},
		},
		model.Measurement{
			Connect: &model.ConnectEvent{
				RemoteAddress: "1.1.1.1:53",
				Error:         errors.New("mocked error"),
			},
		},
	}})
	// make sure we don't stop processing after first error
	if len(tcpConnect) != 1 {
		t.Fatal("unexpected list length")
	}
}

func TestTCPConnectListSmallPortNumber(t *testing.T) {
	tcpConnect := NewTCPConnectList([][]model.Measurement{[]model.Measurement{
		model.Measurement{
			Connect: &model.ConnectEvent{
				RemoteAddress: "1.1.1.1:-1",
				Error:         errors.New("mocked error"),
			},
		},
		model.Measurement{
			Connect: &model.ConnectEvent{
				RemoteAddress: "1.1.1.1:53",
				Error:         errors.New("mocked error"),
			},
		},
	}})
	// make sure we don't stop processing after first error
	if len(tcpConnect) != 1 {
		t.Fatal("unexpected list length")
	}
}

func TestRequestListNil(t *testing.T) {
	out := NewRequestList(nil)
	if len(out) != 0 {
		t.Fatal("unexpected list length")
	}
}

func TestRequestListGood(t *testing.T) {
	out := NewRequestList([][]model.Measurement{[]model.Measurement{
		model.Measurement{
			HTTPRequestHeadersDone: &model.HTTPRequestHeadersDoneEvent{
				Headers: http.Header{
					"Content-Type": []string{
						"text/plain",
						"application/json", // should miss this one
					},
				},
				Method: "GET",
				URL:    "http://generic.antani/",
			},
		},
		model.Measurement{
			HTTPResponseHeadersDone: &model.HTTPResponseHeadersDoneEvent{
				Headers: http.Header{
					"Content-Type": []string{
						"text/plain",
						"application/json", // should miss this one
					},
				},
				StatusCode: 304,
			},
		},
		model.Measurement{
			HTTPResponseBodyPart: &model.HTTPResponseBodyPartEvent{
				Data:     []byte(`abc`),
				NumBytes: 3,
			},
		},
		model.Measurement{
			HTTPResponseBodyPart: &model.HTTPResponseBodyPartEvent{
				Data:     []byte(`def`),
				Error:    io.EOF, // see if we read body anyway
				NumBytes: 3,
			},
		},
	}})
	if len(out) != 1 {
		t.Fatal("unexpected list length")
	}
	e := out[0]
	if e.Failure != nil {
		t.Fatal("unexpected failure")
	}
	if e.Request.Body.Value != "" {
		t.Fatal("unexpected request body value")
	}
	if len(e.Request.Headers) != 1 {
		t.Fatal("unexpected number of request headers")
	}
	if e.Request.Headers["Content-Type"] != "text/plain" {
		t.Fatal("unexpected request content-type value")
	}
	if e.Request.Method != "GET" {
		t.Fatal("unexpected request method value")
	}
	if e.Request.Tor.ExitIP != nil {
		t.Fatal("unexpected request Tor.ExitIP value")
	}
	if e.Request.Tor.ExitName != nil {
		t.Fatal("unexpected request Tor.ExitName value")
	}
	if e.Request.Tor.IsTor != false {
		t.Fatal("unexpected request Tor.IsTor value")
	}
	if e.Request.URL != "http://generic.antani/" {
		t.Fatal("unexpected request URL")
	}
	if e.Response.Body.Value != "abcdef" {
		t.Fatal("unexpected response body value")
	}
	if e.Response.Code != 304 {
		t.Fatal("unexpected response code value")
	}
	if len(e.Response.Headers) != 1 {
		t.Fatal("unexpected response of request headers")
	}
	if e.Response.Headers["Content-Type"] != "text/plain" {
		t.Fatal("unexpected response content-type value")
	}
}

func TestRequestListErrors(t *testing.T) {
	out := NewRequestList([][]model.Measurement{[]model.Measurement{
		model.Measurement{
			Resolve: &model.ResolveEvent{
				Error: errors.New("e1"),
			},
		},
		model.Measurement{
			Connect: &model.ConnectEvent{
				Error: errors.New("e2"),
			},
		},
		model.Measurement{
			Read: &model.ReadEvent{
				Error: errors.New("e4"),
			},
		},
		model.Measurement{
			Write: &model.WriteEvent{
				Error: errors.New("e5"),
			},
		},
		model.Measurement{
			HTTPRequestHeadersDone: &model.HTTPRequestHeadersDoneEvent{
				Headers: http.Header{
					"Content-Type": []string{
						"text/plain",
						"application/json", // should miss this one
					},
				},
				Method: "GET",
				URL:    "http://generic.antani/",
			},
		},
		model.Measurement{
			HTTPResponseHeadersDone: &model.HTTPResponseHeadersDoneEvent{
				Headers: http.Header{
					"Content-Type": []string{
						"text/plain",
						"application/json", // should miss this one
					},
				},
				StatusCode: 304,
			},
		},
		model.Measurement{
			HTTPResponseBodyPart: &model.HTTPResponseBodyPartEvent{
				Data:     []byte(`abc`),
				NumBytes: 3,
			},
		},
		model.Measurement{
			HTTPResponseBodyPart: &model.HTTPResponseBodyPartEvent{
				Data:     []byte(`def`),
				Error:    errors.New("e6"),
				NumBytes: 3,
			},
		},
	}})
	if len(out) != 1 {
		t.Fatal("unexpected list length")
	}
	e := out[0]
	if e.Failure == nil || *e.Failure != "e6" {
		t.Fatal("unexpected failure")
	}
	if e.Request.Body.Value != "" {
		t.Fatal("unexpected request body value")
	}
	if len(e.Request.Headers) != 1 {
		t.Fatal("unexpected number of request headers")
	}
	if e.Request.Headers["Content-Type"] != "text/plain" {
		t.Fatal("unexpected request content-type value")
	}
	if e.Request.Method != "GET" {
		t.Fatal("unexpected request method value")
	}
	if e.Request.Tor.ExitIP != nil {
		t.Fatal("unexpected request Tor.ExitIP value")
	}
	if e.Request.Tor.ExitName != nil {
		t.Fatal("unexpected request Tor.ExitName value")
	}
	if e.Request.Tor.IsTor != false {
		t.Fatal("unexpected request Tor.IsTor value")
	}
	if e.Request.URL != "http://generic.antani/" {
		t.Fatal("unexpected request URL")
	}
	if e.Response.Body.Value != "abcdef" {
		t.Fatal("unexpected response body value")
	}
	if e.Response.Code != 304 {
		t.Fatal("unexpected response code value")
	}
	if len(e.Response.Headers) != 1 {
		t.Fatal("unexpected response of request headers")
	}
	if e.Response.Headers["Content-Type"] != "text/plain" {
		t.Fatal("unexpected response content-type value")
	}
}

func TestMarshalBodyString(t *testing.T) {
	body := HTTPBody{
		Value: "1234",
	}
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, []byte(`"1234"`)) {
		t.Fatal("result is unexpected")
	}
}

func TestMarshalBodyBinary(t *testing.T) {
	input := []uint8{
		0x57, 0xe5, 0x79, 0xfb, 0xa6, 0xbb, 0x0d, 0xbc, 0xce, 0xbd, 0xa7, 0xa0,
		0xba, 0xa4, 0x78, 0x78, 0x12, 0x59, 0xee, 0x68, 0x39, 0xa4, 0x07, 0x98,
		0xc5, 0x3e, 0xbc, 0x55, 0xcb, 0xfe, 0x34, 0x3c, 0x7e, 0x1b, 0x5a, 0xb3,
		0x22, 0x9d, 0xc1, 0x2d, 0x6e, 0xca, 0x5b, 0xf1, 0x10, 0x25, 0x47, 0x1e,
		0x44, 0xe2, 0x2d, 0x60, 0x08, 0xea, 0xb0, 0x0a, 0xcc, 0x05, 0x48, 0xa0,
		0xf5, 0x78, 0x38, 0xf0, 0xdb, 0x3f, 0x9d, 0x9f, 0x25, 0x6f, 0x89, 0x00,
		0x96, 0x93, 0xaf, 0x43, 0xac, 0x4d, 0xc9, 0xac, 0x13, 0xdb, 0x22, 0xbe,
		0x7a, 0x7d, 0xd9, 0x24, 0xa2, 0x52, 0x69, 0xd8, 0x89, 0xc1, 0xd1, 0x57,
		0xaa, 0x04, 0x2b, 0xa2, 0xd8, 0xb1, 0x19, 0xf6, 0xd5, 0x11, 0x39, 0xbb,
		0x80, 0xcf, 0x86, 0xf9, 0x5f, 0x9d, 0x8c, 0xab, 0xf5, 0xc5, 0x74, 0x24,
		0x3a, 0xa2, 0xd4, 0x40, 0x4e, 0xd7, 0x10, 0x1f,
	}
	body := HTTPBody{
		Value: string(input),
	}
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(data))
	if !bytes.Equal(data, []byte(`{"data":"V+V5+6a7DbzOvaeguqR4eBJZ7mg5pAeYxT68Vcv+NDx+G1qzIp3BLW7KW/EQJUceROItYAjqsArMBUig9Xg48Ns/nZ8lb4kAlpOvQ6xNyawT2yK+en3ZJKJSadiJwdFXqgQrotixGfbVETm7gM+G+V+djKv1xXQkOqLUQE7XEB8=","format":"base64"}`)) {
		t.Fatal("result is unexpected")
	}
}

func TestRequestListEmptyRoundTrips(t *testing.T) {
	out := NewRequestList([][]model.Measurement{
		nil, nil, nil,
	})
	t.Logf("%+v", out)
	if len(out) != 0 {
		t.Fatal("unexpected list length")
	}
}
