// Package oodataformat contains the OONI data format.
package oodataformat

import (
	"net"
	"strconv"

	"github.com/ooni/netx/model"
)

// TCPConnectStatus contains the TCP connect status.
type TCPConnectStatus struct {
	Failure *string `json:"failure"`
	Success bool    `json:"success"`
}

// TCPConnectEntry contains one of the entries that are part
// of the "tcp_connect" key of a OONI report.
type TCPConnectEntry struct {
	IP     string           `json:"ip"`
	Port   int              `json:"port"`
	Status TCPConnectStatus `json:"status"`
}

// TCPConnectList is a list of TCPConnectEntry
type TCPConnectList []TCPConnectEntry

// NewTCPConnectList creates a new TCPConnectList
func NewTCPConnectList(events [][]model.Measurement) TCPConnectList {
	var out TCPConnectList
	for _, roundTripEvents := range events {
		for _, ev := range roundTripEvents {
			if ev.Connect != nil {
				// We assume Go is passing us legit data structs
				ip, sport, err := net.SplitHostPort(ev.Connect.RemoteAddress)
				if err != nil {
					continue
				}
				iport, err := strconv.Atoi(sport)
				if err != nil {
					continue
				}
				if iport < 0 || iport > 65535 {
					continue
				}
				out = append(out, TCPConnectEntry{
					IP:   ip,
					Port: iport,
					Status: TCPConnectStatus{
						Failure: makeFailure(ev.Connect.Error),
						Success: ev.Connect.Error == nil,
					},
				})
			}
		}
	}
	return out
}

func makeFailure(err error) (s *string) {
	if err != nil {
		serio := err.Error()
		s = &serio
	}
	return
}
