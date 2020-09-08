package oonimkall

import (
	"sync/atomic"
	"testing"
	"time"
)

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
