package oonimkall_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/oonimkall"
)

func NewSession() (*oonimkall.Session, error) {
	return oonimkall.NewSession(&oonimkall.SessionConfig{
		AssetsDir:        "../testdata/oonimkall/assets",
		ProbeServicesURL: "https://ams-pg.ooni.org/",
		SoftwareName:     "oonimkall-test",
		SoftwareVersion:  "0.1.0",
		StateDir:         "../testdata/oonimkall/state",
		TempDir:          "../testdata/",
	})
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

func ReduceErrorForGeolocate(err error) error {
	if err == nil {
		return errors.New("we expected an error here")
	}
	if errors.Is(err, context.Canceled) {
		return nil // when we have not downloaded the resources yet
	}
	if err.Error() == "All IP lookuppers failed" {
		return nil // otherwise
	}
	return fmt.Errorf("not the error we expected: %w", err)
}

func TestGeolocateWithCancelledContext(t *testing.T) {
	sess, err := NewSession()
	if err != nil {
		t.Fatal(err)
	}
	ctx := sess.NewContext()
	ctx.Cancel() // cause immediate failure
	location, err := sess.Geolocate(ctx)
	if err := ReduceErrorForGeolocate(err); err != nil {
		t.Fatal(err)
	}
	if location != nil {
		t.Fatal("expected nil location here")
	}
}

func TestGeolocateGood(t *testing.T) {
	sess, err := NewSession()
	if err != nil {
		t.Fatal(err)
	}
	ctx := sess.NewContext()
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

func ReduceErrorForSubmitter(err error) error {
	if err == nil {
		return errors.New("we expected an error here")
	}
	if errors.Is(err, context.Canceled) {
		return nil // when we have not downloaded the resources yet
	}
	if err.Error() == "all available probe services failed" {
		return nil // otherwise
	}
	return fmt.Errorf("not the error we expected: %w", err)
}

func TestSubmitWithCancelledContext(t *testing.T) {
	sess, err := NewSession()
	if err != nil {
		t.Fatal(err)
	}
	ctx := sess.NewContext()
	ctx.Cancel() // cause immediate failure
	result, err := sess.Submit(ctx, "{}")
	if err := ReduceErrorForSubmitter(err); err != nil {
		t.Fatalf("not the error we expected: %+v", err)
	}
	if result != nil {
		t.Fatal("expected nil result here")
	}
}

func TestSubmitWithInvalidJSON(t *testing.T) {
	sess, err := NewSession()
	if err != nil {
		t.Fatal(err)
	}
	ctx := sess.NewContext()
	result, err := sess.Submit(ctx, "{")
	if err == nil || err.Error() != "unexpected end of JSON input" {
		t.Fatalf("not the error we expected: %+v", err)
	}
	if result != nil {
		t.Fatal("expected nil result here")
	}
}

func DoSubmission(ctx *oonimkall.Context, sess *oonimkall.Session) error {
	inputm := model.Measurement{
		DataFormatVersion:    "0.2.0",
		MeasurementStartTime: "2019-10-28 12:51:07",
		MeasurementRuntime:   1.71,
		ProbeASN:             "AS30722",
		ProbeCC:              "IT",
		ProbeIP:              "127.0.0.1",
		ReportID:             "",
		ResolverIP:           "172.217.33.129",
		SoftwareName:         "miniooni",
		SoftwareVersion:      "0.1.0-dev",
		TestKeys:             map[string]bool{"success": true},
		TestName:             "example",
		TestVersion:          "0.1.0",
	}
	inputd, err := json.Marshal(inputm)
	if err != nil {
		return err
	}
	result, err := sess.Submit(ctx, string(inputd))
	if err != nil {
		return fmt.Errorf("session_test.go: submit failed: %w", err)
	}
	if result.UpdatedMeasurement == "" {
		return errors.New("expected non empty measurement")
	}
	if result.UpdatedReportID == "" {
		return errors.New("expected non empty report ID")
	}
	var outputm model.Measurement
	return json.Unmarshal([]byte(result.UpdatedMeasurement), &outputm)
}

func TestSubmitMeasurementGood(t *testing.T) {
	sess, err := NewSession()
	if err != nil {
		t.Fatal(err)
	}
	ctx := sess.NewContext()
	if err := DoSubmission(ctx, sess); err != nil {
		t.Fatal(err)
	}
}

func TestSubmitCancelContextAfterFirstSubmission(t *testing.T) {
	sess, err := NewSession()
	if err != nil {
		t.Fatal(err)
	}
	ctx := sess.NewContext()
	if err := DoSubmission(ctx, sess); err != nil {
		t.Fatal(err)
	}
	ctx.Cancel() // fail second submission
	err = DoSubmission(ctx, sess)
	if err == nil || !strings.HasPrefix(err.Error(), "session_test.go: submit failed") {
		t.Fatalf("not the error we expected: %+v", err)
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("not the error we expected: %+v", err)
	}
}
