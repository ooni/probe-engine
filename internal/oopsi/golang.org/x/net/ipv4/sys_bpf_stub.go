// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !linux

package ipv4

import (
	"github.com/ooni/probe-engine/internal/oopsi/golang.org/x/net/bpf"
	"github.com/ooni/probe-engine/internal/oopsi/golang.org/x/net/internal/socket"
)

func (so *sockOpt) setAttachFilter(c *socket.Conn, f []bpf.RawInstruction) error {
	return errNotImplemented
}
