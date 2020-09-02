package oonimkall_test

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/oonimkall"
)

type RecordingLogger struct {
	DebugLogs []string
	InfoLogs  []string
	WarnLogs  []string
	mu        sync.Mutex
}

func (rl *RecordingLogger) Debug(msg string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.DebugLogs = append(rl.DebugLogs, msg)
}

func (rl *RecordingLogger) Info(msg string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.InfoLogs = append(rl.InfoLogs, msg)
}

func (rl *RecordingLogger) Warn(msg string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.WarnLogs = append(rl.WarnLogs, msg)
}

func LoggerEmitMessages(logger model.Logger) {
	logger.Warnf("a formatted warn message: %+v", io.EOF)
	logger.Warn("a warn string")
	logger.Infof("a formatted info message: %+v", io.EOF)
	logger.Info("a info string")
	logger.Debugf("a formatted debug message: %+v", io.EOF)
	logger.Debug("a debug string")
}

func TestNewLoggerNilConfig(t *testing.T) {
	// The objective of this test is to make sure that, even if the
	// config instance is nil, we get back something that works, that
	// is, something that does not crash when it is used.
	logger := oonimkall.NewLogger(nil)
	LoggerEmitMessages(logger)
}

func TestNewLoggerNilLogger(t *testing.T) {
	// The objective of this test is to make sure that, even if the
	// Logger instance is nil, we get back something that works, that
	// is, something that does not crash when it is used.
	logger := oonimkall.NewLogger(&oonimkall.SessionConfig{Logger: nil})
	LoggerEmitMessages(logger)
}

func (rl *RecordingLogger) VerifyNumberOfEntries(debugEntries int) error {
	if len(rl.DebugLogs) != debugEntries {
		return errors.New("unexpected number of debug messages")
	}
	if len(rl.InfoLogs) != 2 {
		return errors.New("unexpected number of info messages")
	}
	if len(rl.WarnLogs) != 2 {
		return errors.New("unexpected number of warn messages")
	}
	return nil
}

func (rl *RecordingLogger) ExpectedEntries(level string) []string {
	return []string{
		fmt.Sprintf("a formatted %s message: %+v", level, io.EOF),
		fmt.Sprintf("a %s string", level),
	}
}

func (rl *RecordingLogger) CheckNonVerboseEntries() error {
	if diff := cmp.Diff(rl.InfoLogs, rl.ExpectedEntries("info")); diff != "" {
		return errors.New(diff)
	}
	if diff := cmp.Diff(rl.WarnLogs, rl.ExpectedEntries("warn")); diff != "" {
		return errors.New(diff)
	}
	return nil
}

func (rl *RecordingLogger) CheckVerboseEntries() error {
	if diff := cmp.Diff(rl.DebugLogs, rl.ExpectedEntries("debug")); diff != "" {
		return errors.New(diff)
	}
	return nil
}
func TestNewLoggerQuietLogger(t *testing.T) {
	handler := new(RecordingLogger)
	logger := oonimkall.NewLogger(&oonimkall.SessionConfig{Logger: handler})
	LoggerEmitMessages(logger)
	handler.VerifyNumberOfEntries(0)
	if err := handler.CheckNonVerboseEntries(); err != nil {
		t.Fatal(err)
	}
}

func TestNewLoggerVerboseLogger(t *testing.T) {
	handler := new(RecordingLogger)
	logger := oonimkall.NewLogger(&oonimkall.SessionConfig{
		Logger:  handler,
		Verbose: true,
	})
	LoggerEmitMessages(logger)
	handler.VerifyNumberOfEntries(2)
	if err := handler.CheckNonVerboseEntries(); err != nil {
		t.Fatal(err)
	}
	if err := handler.CheckVerboseEntries(); err != nil {
		t.Fatal(err)
	}
}

func TestNullContextDoesNotCrash(t *testing.T) {
	var ctx *oonimkall.Context
	ctx.Cancel()
	if err := ctx.Close(); err != nil {
		t.Fatal(err)
	}
	if ctx.Timeout() != 0 {
		t.Fatal("invalid Timeout value")
	}
}

func TestNewContext(t *testing.T) {
	ctx := oonimkall.NewContext()
	if ctx.Timeout() != 0 {
		t.Fatal("invalid Timeout value")
	}
	go func() {
		<-time.After(250 * time.Millisecond)
		ctx.Cancel()
	}()
	<-ctx.Context().Done()
}

func TestNewContextWithNegativeTimeout(t *testing.T) {
	ctx := oonimkall.NewContextWithTimeout(-1)
	if ctx.Timeout() != 0 {
		t.Fatal("invalid Timeout value")
	}
	go func() {
		<-time.After(250 * time.Millisecond)
		ctx.Cancel()
	}()
	<-ctx.Context().Done()
}

func TestNewContextWithHugeTimeout(t *testing.T) {
	ctx := oonimkall.NewContextWithTimeout(oonimkall.MaxContextTimeout + 1)
	if ctx.Timeout() != oonimkall.MaxContextTimeout {
		t.Fatal("invalid Timeout value")
	}
	go func() {
		<-time.After(250 * time.Millisecond)
		ctx.Cancel()
	}()
	<-ctx.Context().Done()
}

func TestNewContextWithReasonableTimeout(t *testing.T) {
	ctx := oonimkall.NewContextWithTimeout(1)
	if ctx.Timeout() != 1 {
		t.Fatal("invalid Timeout value")
	}
	go func() {
		<-time.After(5 * time.Second)
		ctx.Cancel()
	}()
	start := time.Now()
	<-ctx.Context().Done()
	if time.Now().Sub(start) >= 2*time.Second {
		t.Fatal("context timeout not working as intended")
	}
}
