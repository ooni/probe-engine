package patternsplitter_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/ooni/probe-engine/experiment/tlstool/internal/patternsplitter"
)

func TestSplitDialerFailure(t *testing.T) {
	expected := errors.New("mocked error")
	d := patternsplitter.SplitDialer{Dialer: patternsplitter.FakeDialer{Err: expected}}
	conn, err := d.DialContext(context.Background(), "tcp", "1.1.1.1:853")
	if !errors.Is(err, expected) {
		t.Fatalf("not the error we expected: %+v", err)
	}
	if conn != nil {
		t.Fatal("expected nil conn here")
	}
}

func TestSplitDialerSuccess(t *testing.T) {
	innerconn := &patternsplitter.FakeConn{}
	d := patternsplitter.SplitDialer{
		Dialer:  patternsplitter.FakeDialer{Conn: innerconn},
		Delay:   1234,
		Pattern: "abcdef",
	}
	conn, err := d.DialContext(context.Background(), "tcp", "1.1.1.1:853")
	if err != nil {
		t.Fatal(err)
	}
	realconn, ok := conn.(patternsplitter.SplitConn)
	if !ok {
		t.Fatal("cannot cast conn to patternsplitter.SplitConn")
	}
	if realconn.Delay != 1234 {
		t.Fatal("invalid Delay value")
	}
	if diff := cmp.Diff(realconn.Pattern, []byte("abcdef")); diff != "" {
		t.Fatal(diff)
	}
	if realconn.Conn != innerconn {
		t.Fatal("invalid Conn value")
	}
}

func TestWriteSuccessNoSplit(t *testing.T) {
	const (
		pattern = "abc.def"
		data    = "deadbeefdeafbeef"
	)
	innerconn := &patternsplitter.FakeConn{}
	conn := patternsplitter.SplitConn{
		Conn:    innerconn,
		Pattern: []byte(pattern),
	}
	count, err := conn.Write([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	if count != len(data) {
		t.Fatal("invalid count")
	}
	if len(innerconn.WriteData) != 1 {
		t.Fatal("invalid number of writes")
	}
	if string(innerconn.WriteData[0]) != data {
		t.Fatal("written invalid data")
	}
}

func TestWriteFailureNoSplit(t *testing.T) {
	const (
		pattern = "abc.def"
		data    = "deadbeefdeafbeef"
	)
	expected := errors.New("mocked error")
	innerconn := &patternsplitter.FakeConn{
		WriteError: expected,
	}
	conn := patternsplitter.SplitConn{
		Conn:    innerconn,
		Pattern: []byte(pattern),
	}
	count, err := conn.Write([]byte(data))
	if !errors.Is(err, expected) {
		t.Fatalf("not the error we expected: %+v", err)
	}
	if count != 0 {
		t.Fatal("invalid count")
	}
}

func TestWriteSuccessSplit(t *testing.T) {
	const (
		pattern = "abc.def"
		data    = "deadbeefabc.defdeafbeef"
	)
	innerconn := &patternsplitter.FakeConn{}
	conn := patternsplitter.SplitConn{
		Conn:    innerconn,
		Pattern: []byte(pattern),
	}
	count, err := conn.Write([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	if count != len(data) {
		t.Fatal("invalid count")
	}
	if len(innerconn.WriteData) != 2 {
		t.Fatal("invalid number of writes")
	}
	if string(innerconn.WriteData[0]) != "deadbeefabc" {
		t.Fatal("written invalid data")
	}
	if string(innerconn.WriteData[1]) != ".defdeafbeef" {
		t.Fatal("written invalid data")
	}
}

func TestWriteFailureSplitFirstWrite(t *testing.T) {
	const (
		pattern = "abc.def"
		data    = "deadbeefabc.defdeafbeef"
	)
	expected := errors.New("mocked error")
	innerconn := &patternsplitter.FakeConn{
		WriteError: expected,
	}
	conn := patternsplitter.SplitConn{
		Conn:    innerconn,
		Pattern: []byte(pattern),
	}
	count, err := conn.Write([]byte(data))
	if !errors.Is(err, expected) {
		t.Fatalf("not the error we expected: %+v", err)
	}
	if count != 0 {
		t.Fatal("invalid count")
	}
	if len(innerconn.WriteData) != 0 {
		t.Fatal("some data has been written")
	}
}

func TestWriteFailureSplitSecondWrite(t *testing.T) {
	const (
		pattern = "abc.def"
		data    = "deadbeefabc.defdeafbeef"
	)
	expected := errors.New("mocked error")
	innerconn := &patternsplitter.FakeConn{}
	conn := patternsplitter.SplitConn{
		BeforeSecondWrite: func() {
			innerconn.WriteError = expected // second write will then fail
		},
		Conn:    innerconn,
		Pattern: []byte(pattern),
	}
	count, err := conn.Write([]byte(data))
	if !errors.Is(err, expected) {
		t.Fatalf("not the error we expected: %+v", err)
	}
	if count != 0 {
		t.Fatal("invalid count")
	}
	if len(innerconn.WriteData) != 1 {
		t.Fatal("we expected to see just one write")
	}
	if string(innerconn.WriteData[0]) != "deadbeefabc" {
		t.Fatal("written invalid data")
	}
}
