package measurexlite

import (
	"testing"
	"time"

	"github.com/ooni/probe-engine/pkg/model"
	"github.com/ooni/probe-engine/pkg/model/mocks"
	utls "gitlab.com/yawning/utls.git"
)

func TestNewTLSHandshakerUTLS(t *testing.T) {
	t.Run("NewTLSHandshakerUTLS creates a wrapped TLSHandshaker", func(t *testing.T) {
		underlying := &mocks.TLSHandshaker{}
		zeroTime := time.Now()
		trace := NewTrace(0, zeroTime)
		trace.NewTLSHandshakerUTLSFn = func(dl model.DebugLogger, id *utls.ClientHelloID) model.TLSHandshaker {
			return underlying
		}
		thx := trace.NewTLSHandshakerUTLS(model.DiscardLogger, &utls.HelloGolang)
		thxt := thx.(*tlsHandshakerTrace)
		if thxt.thx != underlying {
			t.Fatal("invalid TLS handshaker")
		}
		if thxt.tx != trace {
			t.Fatal("invalid trace")
		}
	})
}
