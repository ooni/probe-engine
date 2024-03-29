package netemx

import (
	"context"
	"net"
	"testing"

	"github.com/apex/log"
	"github.com/google/go-cmp/cmp"
	"github.com/ooni/probe-engine/pkg/netxlite"
)

func TestDNSOverUDPServerFactory(t *testing.T) {
	env := MustNewQAEnv(
		QAEnvOptionNetStack(AddressDNSGoogle8844, &DNSOverUDPServerFactory{}),
	)
	defer env.Close()

	env.AddRecordToAllResolvers("www.example.com", "", AddressWwwExampleCom)

	env.Do(func() {
		netx := &netxlite.Netx{}
		reso := netx.NewParallelUDPResolver(
			log.Log, netx.NewDialerWithoutResolver(log.Log),
			net.JoinHostPort(AddressDNSGoogle8844, "53"))
		addrs, err := reso.LookupHost(context.Background(), "www.example.com")
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff([]string{AddressWwwExampleCom}, addrs); diff != "" {
			t.Fatal(diff)
		}
	})
}
