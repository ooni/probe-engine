package model

import (
	"bytes"
	"encoding/json"
	"errors"
	"net"
)

// PrivacySettings contains privacy settings for submitting measurements.
type PrivacySettings struct {
	// IncludeASN indicates whether to include the ASN
	IncludeASN bool

	// IncludeCountry indicates whether to include the country
	IncludeCountry bool

	// IncludeIP indicates whether to include the IP
	IncludeIP bool
}

// Apply applies the privacy settings to the measurement, possibly
// scrubbing the probeIP out of it.
func (ps PrivacySettings) Apply(m *Measurement, probeIP string) (err error) {
	if ps.IncludeASN == false {
		m.ProbeASN = DefaultProbeASNString
	}
	if ps.IncludeCountry == false {
		m.ProbeCC = DefaultProbeCC
	}
	if ps.IncludeIP == false {
		m.ProbeIP = DefaultProbeIP
		err = ps.MaybeRewriteTestKeys(m, probeIP, json.Marshal)
	}
	return
}

// MaybeRewriteTestKeys is the function called by Apply that
// ensures that m's serialization doesn't include the IP
func (ps PrivacySettings) MaybeRewriteTestKeys(
	m *Measurement, currentIP string,
	marshal func(interface{}) ([]byte, error),
) error {
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
