package oonimkall

import (
	"testing"
	"time"
)

func TestNewContextWithZeroTimeout(t *testing.T) {
	ctx, cancel := newContext(0)
	defer cancel()
	go func() {
		<-time.After(250 * time.Millisecond)
		cancel()
	}()
	<-ctx.Done()
}

func TestNewContextWithNegativeTimeout(t *testing.T) {
	ctx, cancel := newContext(-1)
	defer cancel()
	go func() {
		<-time.After(250 * time.Millisecond)
		cancel()
	}()
	<-ctx.Done()
}

func TestNewContextWithHugeTimeout(t *testing.T) {
	ctx, cancel := newContext(maxTimeout + 1)
	defer cancel()
	go func() {
		<-time.After(250 * time.Millisecond)
		cancel()
	}()
	<-ctx.Done()
}

func TestNewContextWithReasonableTimeout(t *testing.T) {
	ctx, cancel := newContext(1)
	defer cancel()
	go func() {
		<-time.After(5 * time.Second)
		cancel()
	}()
	start := time.Now()
	<-ctx.Done()
	if time.Now().Sub(start) >= 2*time.Second {
		t.Fatal("context timeout not working as intended")
	}
}
