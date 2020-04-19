package ndt7

import (
	"context"
	"strings"
	"testing"

	"github.com/apex/log"
)

func TestDialDownloadWithCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // immediately halt
	mgr := newDialManager("hostname.fake", nil, log.Log)
	conn, err := mgr.dialDownload(ctx)
	if err == nil || !strings.HasSuffix(err.Error(), "operation was canceled") {
		t.Fatal("not the error we expected")
	}
	if conn != nil {
		t.Fatal("expected nil conn here")
	}
}

func TestDialUploadWithCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // immediately halt
	mgr := newDialManager("hostname.fake", nil, log.Log)
	conn, err := mgr.dialUpload(ctx)
	if err == nil || !strings.HasSuffix(err.Error(), "operation was canceled") {
		t.Fatal("not the error we expected")
	}
	if conn != nil {
		t.Fatal("expected nil conn here")
	}
}
