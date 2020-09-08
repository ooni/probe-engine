package oonimkall

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestClampTimeout(t *testing.T) {
	if clampTimeout(-1, maxTimeout) != -1 {
		t.Fatal("unexpected result here")
	}
	if clampTimeout(0, maxTimeout) != 0 {
		t.Fatal("unexpected result here")
	}
	if clampTimeout(60, maxTimeout) != 60 {
		t.Fatal("unexpected result here")
	}
	if clampTimeout(maxTimeout, maxTimeout) != maxTimeout {
		t.Fatal("unexpected result here")
	}
	if clampTimeout(maxTimeout+1, maxTimeout) != maxTimeout {
		t.Fatal("unexpected result here")
	}
}

func TestNewContextWithZeroTimeout(t *testing.T) {
	var here uint32
	ctx, cancel := newContext(0)
	defer cancel()
	go func() {
		<-time.After(250 * time.Millisecond)
		cancel()
		atomic.AddUint32(&here, 1)
	}()
	<-ctx.Done()
	if here != 1 {
		t.Fatal("context timeout not working as intended")
	}
}

func TestNewContextWithNegativeTimeout(t *testing.T) {
	var here uint32
	ctx, cancel := newContext(-1)
	defer cancel()
	go func() {
		<-time.After(250 * time.Millisecond)
		cancel()
		atomic.AddUint32(&here, 1)
	}()
	<-ctx.Done()
	if here != 1 {
		t.Fatal("context timeout not working as intended")
	}
}

func TestNewContextWithHugeTimeout(t *testing.T) {
	var here uint32
	ctx, cancel := newContext(maxTimeout + 1)
	defer cancel()
	go func() {
		<-time.After(250 * time.Millisecond)
		cancel()
		atomic.AddUint32(&here, 1)
	}()
	<-ctx.Done()
	if here != 1 {
		t.Fatal("context timeout not working as intended")
	}
}

func TestNewContextWithReasonableTimeout(t *testing.T) {
	var here uint32
	ctx, cancel := newContext(1)
	defer cancel()
	go func() {
		<-time.After(5 * time.Second)
		cancel()
		atomic.AddUint32(&here, 1)
	}()
	<-ctx.Done()
	if here != 0 {
		t.Fatal("context timeout not working as intended")
	}
}

func TestNewContextWithArtificiallyLowMaxTimeout(t *testing.T) {
	var here uint32
	const maxTimeout = 2
	ctx, cancel := newContextEx(maxTimeout+1, maxTimeout)
	defer cancel()
	go func() {
		<-time.After(30 * time.Second)
		cancel()
		atomic.AddUint32(&here, 1)
	}()
	<-ctx.Done()
	if here != 0 {
		t.Fatal("context timeout not working as intended")
	}
}
