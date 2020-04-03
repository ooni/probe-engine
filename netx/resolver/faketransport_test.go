package resolver

import "context"

// FakeTransport is used by unit tests
type FakeTransport struct {
	Data []byte
	Err  error
}

// RoundTrip implements RoundTripper.RoundTrip
func (ft FakeTransport) RoundTrip(ctx context.Context, query []byte) ([]byte, error) {
	return ft.Data, ft.Err
}

func (ft FakeTransport) RequiresPadding() bool {
	return false
}

func (ft FakeTransport) Address() string {
	return ""
}

func (ft FakeTransport) Network() string {
	return "fake"
}
