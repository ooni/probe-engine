package main

import (
	"errors"
	"os"
	"runtime"
	"testing"

	"github.com/ooni/probe-engine/cmd/jafar/iptables"
	"github.com/ooni/probe-engine/cmd/jafar/shellx"
)

func ensureWeStartOverWithIPTables() {
	iptables.NewCensoringPolicy().Waive()
}

func TestIntegrationNoCommand(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("skipping test on non Linux systems")
	}
	ensureWeStartOverWithIPTables()
	*dnsProxyAddress = "127.0.0.1:0"
	*httpProxyAddress = "127.0.0.1:0"
	*tlsProxyAddress = "127.0.0.1:0"
	go func() {
		mainCh <- os.Interrupt
	}()
	main()
}

func TestIntegrationWithCommand(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("skipping test on non Linux systems")
	}
	ensureWeStartOverWithIPTables()
	*dnsProxyAddress = "127.0.0.1:0"
	*httpProxyAddress = "127.0.0.1:0"
	*tlsProxyAddress = "127.0.0.1:0"
	*mainCommand = "whoami"
	defer func() {
		*mainCommand = ""
	}()
	main()
}

func TestMustx(t *testing.T) {
	t.Run("with no error", func(t *testing.T) {
		var called int
		mustx(nil, "", func(int) {
			called++
		})
		if called != 0 {
			t.Fatal("should not happen")
		}
	})
	t.Run("with non-exit-code error", func(t *testing.T) {
		var (
			called   int
			exitcode int
		)
		mustx(errors.New("antani"), "", func(ec int) {
			called++
			exitcode = ec
		})
		if called != 1 {
			t.Fatal("not called?!")
		}
		if exitcode != 1 {
			t.Fatal("unexpected exitcode value")
		}
	})
	t.Run("with exit-code error", func(t *testing.T) {
		var (
			called   int
			exitcode int
		)
		err := shellx.Run("curl", "-sf", "") // cause exitcode == 3
		mustx(err, "", func(ec int) {
			called++
			exitcode = ec
		})
		if called != 1 {
			t.Fatal("not called?!")
		}
		if exitcode != 3 {
			t.Fatal("unexpected exitcode value")
		}
	})
}
