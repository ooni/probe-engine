package oodatamodel_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/ooni/netx/modelx"
	"github.com/ooni/netx/x/porcelain"
	"github.com/ooni/probe-engine/experiment/oodatamodel"
)

func TestUnitNewTCPConnectListEmpty(t *testing.T) {
	out := oodatamodel.NewTCPConnectList(porcelain.Results{})
	if len(out) != 0 {
		t.Fatal("unexpected output length")
	}
}

func TestUnitNewTCPConnectListSuccess(t *testing.T) {
	out := oodatamodel.NewTCPConnectList(porcelain.Results{
		Connects: []*modelx.ConnectEvent{
			&modelx.ConnectEvent{
				RemoteAddress: "8.8.8.8:53",
			},
			&modelx.ConnectEvent{
				RemoteAddress: "8.8.4.4:853",
			},
		},
	})
	if len(out) != 2 {
		t.Fatal("unexpected output length")
	}
	if out[0].IP != "8.8.8.8" {
		t.Fatal("unexpected out[0].IP")
	}
	if out[0].Port != 53 {
		t.Fatal("unexpected out[0].Port")
	}
	if out[0].Status.Failure != nil {
		t.Fatal("unexpected out[0].Failure")
	}
	if out[0].Status.Success != true {
		t.Fatal("unexpected out[0].Success")
	}
	if out[1].IP != "8.8.4.4" {
		t.Fatal("unexpected out[1].IP")
	}
	if out[1].Port != 853 {
		t.Fatal("unexpected out[1].Port")
	}
	if out[1].Status.Failure != nil {
		t.Fatal("unexpected out[0].Failure")
	}
	if out[1].Status.Success != true {
		t.Fatal("unexpected out[0].Success")
	}
}

func TestUnitNewTCPConnectListFailure(t *testing.T) {
	out := oodatamodel.NewTCPConnectList(porcelain.Results{
		Connects: []*modelx.ConnectEvent{
			&modelx.ConnectEvent{
				RemoteAddress: "8.8.8.8:53",
				Error:         errors.New("connection_reset"),
			},
		},
	})
	if len(out) != 1 {
		t.Fatal("unexpected output length")
	}
	if out[0].IP != "8.8.8.8" {
		t.Fatal("unexpected out[0].IP")
	}
	if out[0].Port != 53 {
		t.Fatal("unexpected out[0].Port")
	}
	if *out[0].Status.Failure != "connection_reset" {
		t.Fatal("unexpected out[0].Failure")
	}
	if out[0].Status.Success != false {
		t.Fatal("unexpected out[0].Success")
	}
}

func TestUnitNewTCPConnectListInvalidInput(t *testing.T) {
	out := oodatamodel.NewTCPConnectList(porcelain.Results{
		Connects: []*modelx.ConnectEvent{
			&modelx.ConnectEvent{
				RemoteAddress: "8.8.8.8",
				Error:         errors.New("connection_reset"),
			},
		},
	})
	if len(out) != 1 {
		t.Fatal("unexpected output length")
	}
	if out[0].IP != "" {
		t.Fatal("unexpected out[0].IP")
	}
	if out[0].Port != 0 {
		t.Fatal("unexpected out[0].Port")
	}
	if *out[0].Status.Failure != "connection_reset" {
		t.Fatal("unexpected out[0].Failure")
	}
	if out[0].Status.Success != false {
		t.Fatal("unexpected out[0].Success")
	}
}

func TestUnitNewRequestsListNil(t *testing.T) {
	out := oodatamodel.NewRequestList(nil)
	if len(out) != 0 {
		t.Fatal("unexpected output length")
	}
}

func TestUnitNewRequestsListEmptyList(t *testing.T) {
	out := oodatamodel.NewRequestList(&porcelain.HTTPDoResults{})
	if len(out) != 0 {
		t.Fatal("unexpected output length")
	}
}

func TestUnitNewRequestsListGood(t *testing.T) {
	out := oodatamodel.NewRequestList(&porcelain.HTTPDoResults{
		TestKeys: porcelain.Results{
			HTTPRequests: []*modelx.HTTPRoundTripDoneEvent{
				// need two requests to test that order is inverted
				&modelx.HTTPRoundTripDoneEvent{
					RequestBodySnap: []byte("abcdefx"),
					RequestHeaders: http.Header{
						"Content-Type": []string{
							"text/plain",
							"foobar",
						},
						"Content-Length": []string{
							"17",
						},
					},
					RequestMethod:    "GET",
					RequestURL:       "http://x.org/",
					ResponseBodySnap: []byte("abcdef"),
					ResponseHeaders: http.Header{
						"Content-Type": []string{
							"application/json",
							"foobaz",
						},
						"Server": []string{
							"antani",
						},
						"Content-Length": []string{
							"14",
						},
					},
					ResponseStatusCode: 451,
					MaxBodySnapSize:    10,
				},
				&modelx.HTTPRoundTripDoneEvent{
					Error: errors.New("antani"),
				},
			},
		},
	})
	if len(out) != 2 {
		t.Fatal("unexpected output length")
	}

	if *out[0].Failure != "antani" {
		t.Fatal("unexpected out[0].Failure")
	}
	if out[0].Request.Body.Value != "" {
		t.Fatal("unexpected out[0].Request.Body.Value")
	}
	if len(out[0].Request.Headers) != 0 {
		t.Fatal("unexpected out[0].Request.Headers")
	}
	if out[0].Request.Method != "" {
		t.Fatal("unexpected out[0].Request.Method")
	}
	if out[0].Request.URL != "" {
		t.Fatal("unexpected out[0].Request.URL")
	}
	if out[0].Request.BodyIsTruncated != false {
		t.Fatal("unexpected out[0].Request.BodyIsTruncated")
	}
	if out[0].Response.Body.Value != "" {
		t.Fatal("unexpected out[0].Response.Body.Value")
	}
	if out[0].Response.Code != 0 {
		t.Fatal("unexpected out[0].Response.Code")
	}
	if len(out[0].Response.Headers) != 0 {
		t.Fatal("unexpected out[0].Response.Headers")
	}
	if out[0].Response.BodyIsTruncated != false {
		t.Fatal("unexpected out[0].Response.BodyIsTruncated")
	}

	if out[1].Failure != nil {
		t.Fatal("unexpected out[1].Failure")
	}
	if out[1].Request.Body.Value != "abcdefx" {
		t.Fatal("unexpected out[1].Request.Body.Value")
	}
	if len(out[1].Request.Headers) != 2 {
		t.Fatal("unexpected out[1].Request.Headers")
	}
	if out[1].Request.Headers["Content-Type"].Value != "text/plain" {
		t.Fatal("unexpected out[1].Request.Headers Content-Type value")
	}
	if out[1].Request.Headers["Content-Length"].Value != "17" {
		t.Fatal("unexpected out[1].Request.Headers Content-Length value")
	}
	if out[1].Request.Method != "GET" {
		t.Fatal("unexpected out[1].Request.Method")
	}
	if out[1].Request.URL != "http://x.org/" {
		t.Fatal("unexpected out[1].Request.URL")
	}
	if out[1].Request.BodyIsTruncated != false {
		t.Fatal("unexpected out[1].Request.BodyIsTruncated")
	}
	if out[1].Response.Body.Value != "abcdef" {
		t.Fatal("unexpected out[1].Response.Body.Value")
	}
	if out[1].Response.Code != 451 {
		t.Fatal("unexpected out[1].Response.Code")
	}
	if len(out[1].Response.Headers) != 3 {
		t.Fatal("unexpected out[1].Response.Headers")
	}
	if out[1].Response.Headers["Content-Type"].Value != "application/json" {
		t.Fatal("unexpected out[1].Response.Headers Content-Type value")
	}
	if out[1].Response.Headers["Server"].Value != "antani" {
		t.Fatal("unexpected out[1].Response.Headers Server value")
	}
	if out[1].Response.Headers["Content-Length"].Value != "14" {
		t.Fatal("unexpected out[1].Response.Headers Content-Length value")
	}
	if out[1].Response.BodyIsTruncated != false {
		t.Fatal("unexpected out[1].Response.BodyIsTruncated")
	}
}

func TestUnitNewRequestsSnaps(t *testing.T) {
	out := oodatamodel.NewRequestList(&porcelain.HTTPDoResults{
		TestKeys: porcelain.Results{
			HTTPRequests: []*modelx.HTTPRoundTripDoneEvent{
				&modelx.HTTPRoundTripDoneEvent{
					RequestBodySnap:  []byte("abcd"),
					MaxBodySnapSize:  4,
					ResponseBodySnap: []byte("defg"),
				},
			},
		},
	})
	if len(out) != 1 {
		t.Fatal("unexpected output length")
	}
	if out[0].Request.BodyIsTruncated != true {
		t.Fatal("wrong out[0].Request.BodyIsTruncated")
	}
	if out[0].Response.BodyIsTruncated != true {
		t.Fatal("wrong out[0].Response.BodyIsTruncated")
	}
}

func TestMarshalHTTPBodyString(t *testing.T) {
	mbv := oodatamodel.HTTPBody{
		Value: "1234",
	}
	data, err := json.Marshal(mbv)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, []byte(`"1234"`)) {
		t.Fatal("result is unexpected")
	}
}

func TestMarshalHTTPBodyBinary(t *testing.T) {
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
	mbv := oodatamodel.HTTPBody{
		Value: string(input),
	}
	data, err := json.Marshal(mbv)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, []byte(`{"data":"V+V5+6a7DbzOvaeguqR4eBJZ7mg5pAeYxT68Vcv+NDx+G1qzIp3BLW7KW/EQJUceROItYAjqsArMBUig9Xg48Ns/nZ8lb4kAlpOvQ6xNyawT2yK+en3ZJKJSadiJwdFXqgQrotixGfbVETm7gM+G+V+djKv1xXQkOqLUQE7XEB8=","format":"base64"}`)) {
		t.Fatal("result is unexpected")
	}
}
