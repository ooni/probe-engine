package segmenter_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ooni/probe-engine/experiment/tlstool/internal/segmenter"
)

func TestDialerFailure(t *testing.T) {
	expected := errors.New("mocked error")
	d := segmenter.Dialer{Dialer: segmenter.FakeDialer{Err: expected}}
	conn, err := d.DialContext(context.Background(), "tcp", "1.1.1.1:853")
	if !errors.Is(err, expected) {
		t.Fatalf("not the error we expected: %+v", err)
	}
	if conn != nil {
		t.Fatal("expected nil conn here")
	}
}

func TestDialerSuccess(t *testing.T) {
	innerconn := &segmenter.FakeConn{}
	d := segmenter.Dialer{
		Dialer: segmenter.FakeDialer{Conn: innerconn},
		Delay:  1234,
	}
	conn, err := d.DialContext(context.Background(), "tcp", "1.1.1.1:853")
	if err != nil {
		t.Fatal(err)
	}
	realconn, ok := conn.(segmenter.Conn)
	if !ok {
		t.Fatal("cannot cast conn to segmenter.SplitConn")
	}
	if realconn.Delay != 1234 {
		t.Fatal("invalid Delay value")
	}
	if realconn.Conn != innerconn {
		t.Fatal("invalid Conn value")
	}
}

func TestWriteEdgeCaseSmall(t *testing.T) {
	const data = "1111111"
	innerconn := &segmenter.FakeConn{}
	conn := segmenter.Conn{Conn: innerconn}
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
	if string(innerconn.WriteData[0]) != "1111111" {
		t.Fatal("first write is invalid")
	}
}

func TestWriteEdgeCaseMedium(t *testing.T) {
	const data = "1111111122"
	innerconn := &segmenter.FakeConn{}
	conn := segmenter.Conn{Conn: innerconn}
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
	if string(innerconn.WriteData[0]) != "1111111122" {
		t.Fatal("first write is invalid")
	}
}

func TestWriteEdgeCaseSmallByOne(t *testing.T) {
	const data = "111111112222"
	innerconn := &segmenter.FakeConn{}
	conn := segmenter.Conn{Conn: innerconn}
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
	if string(innerconn.WriteData[0]) != "111111112222" {
		t.Fatal("first write is invalid")
	}
}

func TestWriteEdgeCaseSmallByOneFailure(t *testing.T) {
	expected := errors.New("mocked error")
	const data = "111111112222"
	innerconn := &segmenter.FakeConn{WriteError: expected}
	conn := segmenter.Conn{Conn: innerconn}
	count, err := conn.Write([]byte(data))
	if !errors.Is(err, expected) {
		t.Fatalf("not the error we expected: %+v", err)
	}
	if count != 0 {
		t.Fatal("invalid count")
	}
	if len(innerconn.WriteData) != 0 {
		t.Fatal("invalid number of writes")
	}
}

func TestWriteSuccessMinimalCase(t *testing.T) {
	const data = "1111111122223"
	innerconn := &segmenter.FakeConn{}
	conn := segmenter.Conn{Conn: innerconn}
	count, err := conn.Write([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	if count != len(data) {
		t.Fatal("invalid count")
	}
	if len(innerconn.WriteData) != 3 {
		t.Fatal("invalid number of writes")
	}
	if string(innerconn.WriteData[0]) != "11111111" {
		t.Fatal("first write is invalid")
	}
	if string(innerconn.WriteData[1]) != "2222" {
		t.Fatal("first write is invalid")
	}
	if string(innerconn.WriteData[2]) != "3" {
		t.Fatal("first write is invalid")
	}
}

func TestWriteSuccess(t *testing.T) {
	const data = "111111112222333333333333"
	innerconn := &segmenter.FakeConn{}
	conn := segmenter.Conn{Conn: innerconn}
	count, err := conn.Write([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	if count != len(data) {
		t.Fatal("invalid count")
	}
	if len(innerconn.WriteData) != 3 {
		t.Fatal("invalid number of writes")
	}
	if string(innerconn.WriteData[0]) != "11111111" {
		t.Fatal("first write is invalid")
	}
	if string(innerconn.WriteData[1]) != "2222" {
		t.Fatal("first write is invalid")
	}
	if string(innerconn.WriteData[2]) != "333333333333" {
		t.Fatal("first write is invalid")
	}
}

func TestFirstWriteFailure(t *testing.T) {
	const data = "111111112222333333333333"
	expected := errors.New("mocked error")
	innerconn := &segmenter.FakeConn{WriteError: expected}
	conn := segmenter.Conn{Conn: innerconn}
	count, err := conn.Write([]byte(data))
	if !errors.Is(err, expected) {
		t.Fatalf("not the error we expected: %+v", err)
	}
	if count != 0 {
		t.Fatal("invalid count")
	}
	if len(innerconn.WriteData) != 0 {
		t.Fatal("invalid number of writes")
	}
}

func TestSecondWriteFailure(t *testing.T) {
	const data = "111111112222333333333333"
	expected := errors.New("mocked error")
	innerconn := &segmenter.FakeConn{}
	conn := segmenter.Conn{
		BeforeSecondWrite: func() {
			innerconn.WriteError = expected
		},
		Conn: innerconn,
	}
	count, err := conn.Write([]byte(data))
	if !errors.Is(err, expected) {
		t.Fatalf("not the error we expected: %+v", err)
	}
	if count != 0 {
		t.Fatal("invalid count")
	}
	if len(innerconn.WriteData) != 1 {
		t.Fatal("invalid number of writes")
	}
	if string(innerconn.WriteData[0]) != "11111111" {
		t.Fatal("first write is invalid")
	}
}

func TestThirdWriteFailure(t *testing.T) {
	const data = "111111112222333333333333"
	expected := errors.New("mocked error")
	innerconn := &segmenter.FakeConn{}
	conn := segmenter.Conn{
		BeforeThirdWrite: func() {
			innerconn.WriteError = expected
		},
		Conn: innerconn,
	}
	count, err := conn.Write([]byte(data))
	if !errors.Is(err, expected) {
		t.Fatalf("not the error we expected: %+v", err)
	}
	if count != 0 {
		t.Fatal("invalid count")
	}
	if len(innerconn.WriteData) != 2 {
		t.Fatal("invalid number of writes")
	}
	if string(innerconn.WriteData[0]) != "11111111" {
		t.Fatal("first write is invalid")
	}
	if string(innerconn.WriteData[1]) != "2222" {
		t.Fatal("first write is invalid")
	}
}
