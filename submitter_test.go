package engine

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/model"
)

func TestSubmitterNotEnabled(t *testing.T) {
	ctx := context.Background()
	submitter, err := NewSubmitter(ctx, SubmitterConfig{
		Enabled: false,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := submitter.(stubSubmitter); !ok {
		t.Fatal("we did not get a stubSubmitter instance")
	}
	m := new(model.Measurement)
	if err := submitter.Submit(ctx, m); err != nil {
		t.Fatal(err)
	}
}

type FakeSubmitter struct {
	Calls uint32
	Error error
}

func (fs *FakeSubmitter) Submit(ctx context.Context, m *model.Measurement) error {
	atomic.AddUint32(&fs.Calls, 1)
	return fs.Error
}

var _ Submitter = &FakeSubmitter{}

type FakeSubmitterSession struct {
	Error     error
	Submitter Submitter
}

func (fse FakeSubmitterSession) NewSubmitter(ctx context.Context) (Submitter, error) {
	return fse.Submitter, fse.Error
}

var _ SubmitterSession = FakeSubmitterSession{}

func TestNewSubmitterFails(t *testing.T) {
	expected := errors.New("mocked error")
	ctx := context.Background()
	submitter, err := NewSubmitter(ctx, SubmitterConfig{
		Enabled: true,
		Session: FakeSubmitterSession{Error: expected},
	})
	if !errors.Is(err, expected) {
		t.Fatalf("not the error we expected: %+v", err)
	}
	if submitter != nil {
		t.Fatal("expected nil submitter here")
	}
}

func TestNewSubmitterWithFailedSubmission(t *testing.T) {
	expected := errors.New("mocked error")
	ctx := context.Background()
	fakeSubmitter := &FakeSubmitter{Error: expected}
	submitter, err := NewSubmitter(ctx, SubmitterConfig{
		Enabled: true,
		Logger:  log.Log,
		Session: FakeSubmitterSession{Submitter: fakeSubmitter},
	})
	if err != nil {
		t.Fatal(err)
	}
	measurement := new(model.Measurement)
	err = submitter.Submit(context.Background(), measurement)
	if !errors.Is(err, expected) {
		t.Fatalf("not the error we expected: %+v", err)
	}
	if fakeSubmitter.Calls != 1 {
		t.Fatal("unexpected number of calls")
	}
}
