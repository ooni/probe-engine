package engine

import (
	"errors"

	"github.com/ooni/probe-engine/model"
)

// Saver saves a measurement on some persistent storage.
type Saver interface {
	SaveMeasurement(m *model.Measurement) error
}

// SaverConfig is the configuration for creating a new Saver.
type SaverConfig struct {
	// Enabled is true if saving is enabled.
	Enabled bool

	// Experiment is the experiment we're currently running.
	Experiment SaverExperiment

	// FilePath is the filepath where to append the measurement as a
	// serialized JSON followed by a newline character.
	FilePath string
}

// SaverExperiment is an experiment according to the Saver.
type SaverExperiment interface {
	SaveMeasurement(m *model.Measurement, filepath string) error
}

// NewSaver creates a new instance of Saver.
func NewSaver(config SaverConfig) (Saver, error) {
	if !config.Enabled {
		return fakeSaver{}, nil
	}
	if config.FilePath == "" {
		return nil, errors.New("saver: passed an empty filepath")
	}
	return realSaver{
		Experiment: config.Experiment,
		FilePath:   config.FilePath,
	}, nil
}

type fakeSaver struct{}

func (fs fakeSaver) SaveMeasurement(m *model.Measurement) error {
	return nil
}

var _ Saver = fakeSaver{}

type realSaver struct {
	Experiment SaverExperiment
	FilePath   string
}

func (rs realSaver) SaveMeasurement(m *model.Measurement) error {
	return rs.Experiment.SaveMeasurement(m, rs.FilePath)
}

var _ Saver = realSaver{}
