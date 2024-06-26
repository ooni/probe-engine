package tlsmiddlebox

//
// TCP Connect for tlsmiddlebox
//

import (
	"context"
	"time"

	"github.com/ooni/probe-engine/pkg/logx"
	"github.com/ooni/probe-engine/pkg/measurexlite"
	"github.com/ooni/probe-engine/pkg/model"
)

// TCPConnect performs a TCP connect to filter working addresses
func (m *Measurer) TCPConnect(ctx context.Context, index int64, zeroTime time.Time,
	logger model.Logger, address string, tk *TestKeys) error {
	trace := measurexlite.NewTrace(index, zeroTime)
	dialer := trace.NewDialerWithoutResolver(logger)
	ol := logx.NewOperationLogger(logger, "TCPConnect #%d %s", index, address)
	conn, err := dialer.DialContext(ctx, "tcp", address)
	ol.Stop(err)
	_ = measurexlite.MaybeClose(conn)
	tcpEvents := trace.TCPConnects()
	tk.addTCPConnect(tcpEvents)
	return err
}
