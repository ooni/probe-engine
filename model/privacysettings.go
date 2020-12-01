package model

import (
	"bytes"
	"encoding/json"
	"errors"
	"net"
)

// TODO(bassosimone): this code should be moved into the
// measurement.go file and this file should be deleted

// Scrub scrubs the probeIP out of the measurement.
func (m *Measurement) Scrub(probeIP string) (err error) {
	// We now behave like we can share everything except the
	// probe IP, which we instead cannot ever share
	m.ProbeIP = DefaultProbeIP
	return m.MaybeRewriteTestKeys(probeIP, json.Marshal)
}

// MaybeRewriteTestKeys is the function called by Scrub that
// ensures that m's serialization doesn't include the IP
func (m *Measurement) MaybeRewriteTestKeys(
	currentIP string, marshal func(interface{}) ([]byte, error)) error {
	if net.ParseIP(currentIP) == nil {
		return errors.New("Invalid probe IP string")
	}
	data, err := marshal(m.TestKeys)
	if err != nil {
		return err
	}
	// The check using Count is to save an unnecessary copy performed by
	// ReplaceAll when there are no matches into the body. This is what
	// we would like the common case to be, meaning that the code has done
	// its job correctly and has not leaked the IP.
	bpip := []byte(currentIP)
	if bytes.Count(data, bpip) <= 0 {
		return nil
	}
	data = bytes.ReplaceAll(data, bpip, []byte(`[scrubbed]`))
	// We add an annotation such that hopefully later we can measure the
	// number of cases where we failed to sanitize properly.
	m.AddAnnotation("_probe_engine_sanitize_test_keys", "true")
	return json.Unmarshal(data, &m.TestKeys)
}
