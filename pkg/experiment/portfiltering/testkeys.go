package portfiltering

import "github.com/ooni/probe-engine/pkg/model"

// TestKeys contains the experiment results.
type TestKeys struct {
	TCPConnect []*model.ArchivalTCPConnectResult `json:"tcp_connect"`
}
