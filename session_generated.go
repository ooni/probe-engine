package engine

import "github.com/ooni/probe-engine/model"

// TODO(bassosimone): merge/refactor this file into session.go
// and/or move specific functions together into another file but
// for sure remove/rename this file.

// MaybeProbeIP is like ProbeIP except that we return the real
// value only if the privacy settings allows us. Otherwise, we just
// return the default value, model.DefaultProbeIP.
func (s *Session) MaybeProbeIP() string {
	// We now behave like there's no permission to share the IP.
	return model.DefaultProbeIP
}

// MaybeProbeASN is like ProbeASN except that we return the real
// value only if the privacy settings allows us. Otherwise, we just
// return the default value, model.DefaultProbeASN.
func (s *Session) MaybeProbeASN() uint {
	// We now behave like we have permission to share the ASN.
	return s.ProbeASN()
}

// MaybeProbeASNString is like ProbeASNString except that we return the real
// value only if the privacy settings allows us. Otherwise, we just
// return the default value, model.DefaultProbeASNString.
func (s *Session) MaybeProbeASNString() string {
	// We now behave like we have permission to share the ASN.
	return s.ProbeASNString()
}

// MaybeProbeCC is like ProbeCC except that we return the real
// value only if the privacy settings allows us. Otherwise, we just
// return the default value, model.DefaultProbeCC.
func (s *Session) MaybeProbeCC() string {
	// We now behave like we have permission to share the CC.
	return s.ProbeCC()
}

// MaybeProbeNetworkName is like ProbeNetworkName except that we return the real
// value only if the privacy settings allows us. Otherwise, we just
// return the default value, model.DefaultProbeNetworkName.
func (s *Session) MaybeProbeNetworkName() string {
	// We now behave like we have permission to share the ASN.
	return s.ProbeNetworkName()
}

// MaybeResolverIP is like ResolverIP except that we return the real
// value only if the privacy settings allows us. Otherwise, we just
// return the default value, model.DefaultResolverIP.
func (s *Session) MaybeResolverIP() string {
	// The resolver IP in the worst case allows to know the user's ASN
	// therefore it's fine for us to share it.
	return s.ResolverIP()
}

// MaybeResolverASN is like ResolverASN except that we return the real
// value only if the privacy settings allows us. Otherwise, we just
// return the default value, model.DefaultResolverASN.
func (s *Session) MaybeResolverASN() uint {
	// We now behave like we have permission to share the ASN.
	return s.ResolverASN()
}

// MaybeResolverASNString is like ResolverASNString except that we return the real
// value only if the privacy settings allows us. Otherwise, we just
// return the default value, model.DefaultResolverASNString.
func (s *Session) MaybeResolverASNString() string {
	// We now behave like we have permission to share the ASN.
	return s.ResolverASNString()
}

// MaybeResolverNetworkName is like ResolverNetworkName except that we return the real
// value only if the privacy settings allows us. Otherwise, we just
// return the default value, model.DefaultResolverNetworkName.
func (s *Session) MaybeResolverNetworkName() string {
	// We now behave like we have permission to share the ASN.
	return s.ResolverNetworkName()
}
