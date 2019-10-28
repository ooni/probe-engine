package psiphon

import "errors"

const (
	testName    = "psiphon"
	testVersion = "0.3.0"
)

// Config contains the experiment's configuration.
type Config struct {
	// ConfigFilePath is the path where Psiphon config file is located.
	ConfigFilePath string `ooni:"configuration file path"`

	// WorkDir is the directory where Psiphon should store
	// its configuration database.
	WorkDir string `ooni:"experiment working directory"`
}

// ErrDisabled indicates that we disabled psiphon at compile time
var ErrDisabled = errors.New("Psiphon disabled at compile time")
