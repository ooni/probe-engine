package bytecounter_test

import (
	"testing"

	"github.com/ooni/probe-engine/netx/bytecounter"
)

func TestUnit(t *testing.T) {
	counter := bytecounter.New()
	counter.CountBytesReceived(16384)
	counter.CountBytesSent(2048)
	if counter.BytesSent() != 2048 {
		t.Fatal("invalid bytes sent")
	}
	if counter.BytesReceived() != 16384 {
		t.Fatal("invalid bytes received")
	}
	if v := counter.KibiBytesSent(); v < 1.9 && v > 2.1 {
		t.Fatal("invalid kibibytes sent")
	}
	if v := counter.KibiBytesReceived(); v < 15.9 && v > 16.1 {
		t.Fatal("invalid kibibytes sent")
	}
}
