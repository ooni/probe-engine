package engine

import (
	"context"
	"errors"
	"testing"

	"github.com/ooni/probe-engine/model"
)

type FakeInputProcessorExperiment struct {
	Err error
	M   []*model.Measurement
}

func (fipe *FakeInputProcessorExperiment) MeasureWithContext(
	ctx context.Context, input string) (*model.Measurement, error) {
	if fipe.Err != nil {
		return nil, fipe.Err
	}
	m := new(model.Measurement)
	// Here we add annotations to ensure that the input processor
	// is MERGING annotations as opposed to overwriting them.
	m.AddAnnotation("antani", "antani")
	m.AddAnnotation("foo", "baz") // would be bar below
	m.Input = model.MeasurementTarget(input)
	fipe.M = append(fipe.M, m)
	return m, nil
}

func TestInputProcessorMeasurementFailed(t *testing.T) {
	expected := errors.New("mocked error")
	ip := InputProcessor{
		Experiment: NewInputProcessorExperimentWrapper(
			&FakeInputProcessorExperiment{Err: expected},
		),
		Inputs: []model.URLInfo{{
			URL: "https://www.kernel.org/",
		}},
	}
	ctx := context.Background()
	if err := ip.Run(ctx); !errors.Is(err, expected) {
		t.Fatalf("not the error we expected: %+v", err)
	}
}

type FakeInputProcessorSubmitter struct {
	Err error
	M   []*model.Measurement
}

func (fips *FakeInputProcessorSubmitter) SubmitAndUpdateMeasurementContext(
	ctx context.Context, m *model.Measurement) error {
	fips.M = append(fips.M, m)
	return fips.Err
}

func TestInputProcessorSubmissionFailed(t *testing.T) {
	fipe := &FakeInputProcessorExperiment{}
	expected := errors.New("mocked error")
	ip := InputProcessor{
		Annotations: map[string]string{
			"foo": "bar",
		},
		Experiment: NewInputProcessorExperimentWrapper(fipe),
		Inputs: []model.URLInfo{{
			URL: "https://www.kernel.org/",
		}},
		Options: []string{"fake=true"},
		Submitter: NewInputProcessorSubmitterWrapper(
			&FakeInputProcessorSubmitter{Err: expected},
		),
	}
	ctx := context.Background()
	if err := ip.Run(ctx); !errors.Is(err, expected) {
		t.Fatalf("not the error we expected: %+v", err)
	}
	if len(fipe.M) != 1 {
		t.Fatal("no measurements generated")
	}
	m := fipe.M[0]
	if m.Input != "https://www.kernel.org/" {
		t.Fatal("invalid input")
	}
	if len(m.Annotations) != 2 {
		t.Fatal("invalid number of annotations")
	}
	if m.Annotations["foo"] != "bar" {
		t.Fatal("invalid annotation: foo")
	}
	if m.Annotations["antani"] != "antani" {
		t.Fatal("invalid annotation: antani")
	}
	if len(m.Options) != 1 || m.Options[0] != "fake=true" {
		t.Fatal("options not set")
	}
}

type FakeInputProcessorSaver struct {
	Err error
	M   []*model.Measurement
}

func (fips *FakeInputProcessorSaver) SaveMeasurement(m *model.Measurement) error {
	fips.M = append(fips.M, m)
	return fips.Err
}

func TestInputProcessorSaveOnDiskFailed(t *testing.T) {
	expected := errors.New("mocked error")
	ip := InputProcessor{
		Experiment: NewInputProcessorExperimentWrapper(
			&FakeInputProcessorExperiment{},
		),
		Inputs: []model.URLInfo{{
			URL: "https://www.kernel.org/",
		}},
		Options: []string{"fake=true"},
		Saver: NewInputProcessorSaverWrapper(
			&FakeInputProcessorSaver{Err: expected},
		),
		Submitter: NewInputProcessorSubmitterWrapper(
			&FakeInputProcessorSubmitter{Err: nil},
		),
	}
	ctx := context.Background()
	if err := ip.Run(ctx); !errors.Is(err, expected) {
		t.Fatalf("not the error we expected: %+v", err)
	}
}

func TestInputProcessorGood(t *testing.T) {
	fipe := &FakeInputProcessorExperiment{}
	saver := &FakeInputProcessorSaver{Err: nil}
	submitter := &FakeInputProcessorSubmitter{Err: nil}
	ip := InputProcessor{
		Experiment: NewInputProcessorExperimentWrapper(fipe),
		Inputs: []model.URLInfo{{
			URL: "https://www.kernel.org/",
		}, {
			URL: "https://www.slashdot.org/",
		}},
		Options:   []string{"fake=true"},
		Saver:     NewInputProcessorSaverWrapper(saver),
		Submitter: NewInputProcessorSubmitterWrapper(submitter),
	}
	ctx := context.Background()
	if err := ip.Run(ctx); err != nil {
		t.Fatal(err)
	}
	if len(fipe.M) != 2 || len(saver.M) != 2 || len(submitter.M) != 2 {
		t.Fatal("not all measurements saved")
	}
	if submitter.M[0].Input != "https://www.kernel.org/" {
		t.Fatal("invalid submitter.M[0].Input")
	}
	if submitter.M[1].Input != "https://www.slashdot.org/" {
		t.Fatal("invalid submitter.M[1].Input")
	}
	if saver.M[0].Input != "https://www.kernel.org/" {
		t.Fatal("invalid saver.M[0].Input")
	}
	if saver.M[1].Input != "https://www.slashdot.org/" {
		t.Fatal("invalid saver.M[1].Input")
	}
}
