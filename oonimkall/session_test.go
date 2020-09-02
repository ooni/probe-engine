package oonimkall_test

import (
	"errors"
	"fmt"
	"io"
	"strings"
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
	defer ctx.Close()
	ctx.Cancel()
	if err := ctx.Close(); err != nil {
		t.Fatal(err)
	}
	if ctx.GetTimeout() != 0 {
		t.Fatal("invalid Timeout value")
	}
}

func TestNewContext(t *testing.T) {
	ctx := oonimkall.NewContext()
	defer ctx.Close()
	if ctx.GetTimeout() != 0 {
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
	defer ctx.Close()
	if ctx.GetTimeout() != 0 {
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
	defer ctx.Close()
	if ctx.GetTimeout() != oonimkall.MaxContextTimeout {
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
	defer ctx.Close()
	if ctx.GetTimeout() != 1 {
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

func TestCloseNullSession(t *testing.T) {
	var sess *oonimkall.Session
	if err := sess.Close(); err != nil {
		t.Fatal("sess.Close should not fail here")
	}
}

func TestNewSessionWithNilConfig(t *testing.T) {
	sess, err := oonimkall.NewSession(nil)
	if !errors.Is(err, oonimkall.ErrNullPointer) {
		t.Fatal("not the error we expected")
	}
	if sess != nil {
		t.Fatal("expected a nil Session here")
	}
}

func TestNewSessionWithInvalidStateDir(t *testing.T) {
	sess, err := oonimkall.NewSession(&oonimkall.SessionConfig{
		StateDir: "",
	})
	if err == nil || !strings.HasSuffix(err.Error(), "no such file or directory") {
		t.Fatal("not the error we expected")
	}
	if sess != nil {
		t.Fatal("expected a nil Session here")
	}
}

func TestNewSessionWithMissingSoftwareName(t *testing.T) {
	sess, err := oonimkall.NewSession(&oonimkall.SessionConfig{
		StateDir: "../testdata/oonimkall/state",
	})
	if err == nil || err.Error() != "AssetsDir is empty" {
		t.Fatal("not the error we expected")
	}
	if sess != nil {
		t.Fatal("expected a nil Session here")
	}
}

func NewSession() (*oonimkall.Session, error) {
	return oonimkall.NewSession(&oonimkall.SessionConfig{
		AssetsDir:       "../testdata/oonimkall/assets",
		SoftwareName:    "oonimkall-test",
		SoftwareVersion: "0.1.0",
		StateDir:        "../testdata/oonimkall/state",
		TempDir:         "../testdata/",
	})
}

func TestNewSessionWorksAndWeCanClose(t *testing.T) {
	sess, err := NewSession()
	if err != nil {
		t.Fatal(err)
	}
	if err := sess.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestSessionGeolocateWithNullSession(t *testing.T) {
	var sess *oonimkall.Session
	location, err := sess.Geolocate(oonimkall.NewContext())
	if !errors.Is(err, oonimkall.ErrNullPointer) {
		t.Fatal("not the error we expected")
	}
	if location != nil {
		t.Fatal("expected nil location here")
	}
}

func TestSessionGeolocateWithNullContext(t *testing.T) {
	sess, err := NewSession()
	if err != nil {
		t.Fatal(err)
	}
	defer sess.Close()
	location, err := sess.Geolocate(nil)
	if !errors.Is(err, oonimkall.ErrNullPointer) {
		t.Fatal("not the error we expected")
	}
	if location != nil {
		t.Fatal("expected nil location here")
	}
}

func TestSessionGeolocateWithCancelledContext(t *testing.T) {
	ctx := oonimkall.NewContext()
	defer ctx.Close()
	ctx.Cancel() // cause immediate failure
	sess, err := NewSession()
	if err != nil {
		t.Fatal(err)
	}
	defer sess.Close()
	location, err := sess.Geolocate(ctx)
	t.Log(err)
	if err == nil || err.Error() != "All IP lookuppers failed" {
		t.Fatal("not the error we expected")
	}
	if location != nil {
		t.Fatal("expected nil location here")
	}
}

func TestSessionGeolocateGood(t *testing.T) {
	ctx := oonimkall.NewContext()
	defer ctx.Close()
	sess, err := NewSession()
	if err != nil {
		t.Fatal(err)
	}
	defer sess.Close()
	location, err := sess.Geolocate(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if location.ASN == "" {
		t.Fatal("location.ASN is empty")
	}
	if location.Country == "" {
		t.Fatal("location.Country is empty")
	}
	if location.IP == "" {
		t.Fatal("location.IP is empty")
	}
	if location.Org == "" {
		t.Fatal("location.Org is empty")
	}
}
